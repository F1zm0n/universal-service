package main

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"

	receiver "github.com/F1zm0n/universal-receiver"
)

var (
	configPaths = []string{"/app/config", "../consumer/config"}
	topics      = []string{"email", "verify"}
)

func main() {
	sl := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	parseConfig(configPaths)

	rec := receiver.NewKafkaConsumer(sl, topics)

	rec.Consume()
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
