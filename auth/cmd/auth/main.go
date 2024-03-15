package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"
	"github.com/oklog/oklog/pkg/group"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	authv1 "github.com/F1zm0n/uni-auth/pb"
	"github.com/F1zm0n/uni-auth/pkg/authendpoint"
	"github.com/F1zm0n/uni-auth/pkg/authservice"
	"github.com/F1zm0n/uni-auth/pkg/authtransport"
	"github.com/F1zm0n/uni-auth/repository/postgres"
)

// var configPaths = []string{"../auth/config/", "/config."}
var configPaths = []string{"/app/config/"}

func main() {
	parseConfig(configPaths)
	var (
		debugAddr = viper.GetString("listen.debug.port")
		httpAddr  = viper.GetString("listen.http.port")
		grpcAddr  = viper.GetString("listen.grpc.port")
	)
	postgres := postgres.MustNewPostgresDB()
	postgres.MustMigrateSchema()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// http.DefaultServeMux.Handle("/metrics", promhttp.Handler())
	var (
		service     = authservice.New(logger, postgres)
		endpoints   = authendpoint.New(service, logger)
		grpcServer  = authtransport.NewGRPCServer(endpoints, logger)
		httpHandler = authtransport.NewHTTPServer(endpoints, logger)
	)

	var g group.Group
	{
		debugListener, err := net.Listen("tcp", ":"+debugAddr)
		if err != nil {
			logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "debug/HTTP", "addr", ":"+debugAddr)
			return http.Serve(debugListener, http.DefaultServeMux)
		}, func(error) {
			debugListener.Close()
		})
	}

	{
		httpListener, err := net.Listen("tcp", ":"+httpAddr)
		if err != nil {
			logger.Log("transport", "HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(
			func() error {
				logger.Log("transport", "HTTP", "addr", httpAddr)
				return http.Serve(httpListener, httpHandler)
			},
			func(err error) {
				httpListener.Close()
			},
		)
	}
	{
		grpcListener, err := net.Listen("tcp", ":"+grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", grpcAddr)
			baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
			authv1.RegisterAuthServiceServer(baseServer, grpcServer)
			return baseServer.Serve(grpcListener)
		}, func(err error) {
			grpcListener.Close()
		},
		)
	}
	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	logger.Log("exit", g.Run())
}

func parseConfig(paths []string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	for _, path := range paths {
		viper.AddConfigPath(path)
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
