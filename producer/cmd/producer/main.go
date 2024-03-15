package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log"
	"github.com/oklog/oklog/pkg/group"
	"github.com/spf13/viper"

	"github.com/F1zm0n/universal-producer/pkg/prodendpoint"
	"github.com/F1zm0n/universal-producer/pkg/prodservice"
	"github.com/F1zm0n/universal-producer/pkg/prodtransport"
)

// var configPaths = []string{"../producer/config/"}
var configPaths = []string{"../producer/config", "/app/config/"}

func main() {
	parseConfig(configPaths)
	httpAddr := viper.GetString("listen.http.port")

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var (
		service     = prodservice.New(logger)
		endpoint    = prodendpoint.New(service, logger)
		httpHandler = prodtransport.NewHTTPHandler(endpoint, logger)
	)

	var g group.Group
	{
		httpListener, err := net.Listen("tcp", ":"+httpAddr)
		if err != nil {
			logger.Log("transport", "HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "HTTP", "addr", httpAddr)
			return http.Serve(httpListener, httpHandler)
		}, func(error) {
			httpListener.Close()
		})
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
