package main

import (
	"net"
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/oklog/oklog/pkg/group"
	"github.com/spf13/viper"

	"github.com/F1zm0n/universal-mailer/pkg/mailendpoint"
	"github.com/F1zm0n/universal-mailer/pkg/mailservice"
	"github.com/F1zm0n/universal-mailer/pkg/mailtransport"
)

var configPaths = []string{"../mailer/config", "/app/config/"}

func main() {
	parseConfig(configPaths)
	httpAddr := ":5001"
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	var (
		service     = mailservice.New(logger)
		endpoints   = mailendpoint.New(service, logger)
		httpHandler = mailtransport.NewHTTPHandler(endpoints, logger)
	)
	var g group.Group
	{
		// The HTTP listener mounts the Go kit HTTP handler we created.
		httpListener, err := net.Listen("tcp", httpAddr)
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
