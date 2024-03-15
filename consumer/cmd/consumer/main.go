package main

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"

	receiver "github.com/F1zm0n/universal-receiver"
)

var (
	configPaths = []string{"/app/config", "../consumer/config"}
	topics      = viper.GetStringSlice("kafka.topics.list")
)

func main() {
	sl := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	parseConfig(configPaths)

	rec := receiver.NewKafkaConsumer(sl, topics)

	go rec.Consume()
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
