package main

import (
	"context"
	"log/slog"
	"os"

	mailservice "github.com/F1zm0n/consume_mail/service"
	"github.com/F1zm0n/consume_mail/transport"
)

func main() {
	sl := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	svc := mailservice.New(sl)
	cons := transport.NewKafkaConsumer(sl, "mail", svc)

	go cons.ConsumeMail(context.Background())
	cons.ConsumeVer(context.Background())
}
