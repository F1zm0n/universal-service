package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/viper"

	"github.com/F1zm0n/universal-mailer/internal/transport"
)

var configPaths = []string{"../mailer/config", "/app/config/"}

func main() {
	parseConfig(configPaths)
	sl := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	httpAddr := viper.GetString("listen.http.port")
	trans := transport.NewHttpTransport()
	http.HandleFunc("/mail", trans.HandleMail)
	http.HandleFunc("/verify", trans.HandleVerify)

	sl.Info("serving and listening", slog.String("port", httpAddr))
	log.Fatal(http.ListenAndServe(":"+httpAddr, nil))
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
