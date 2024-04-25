package transport

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	mailservice "github.com/F1zm0n/consume_mail/service"
)

type Consumer interface {
	ConsumeMail(ctx context.Context)
	ConsumeVer(ctx context.Context)
}

type kafkaConsumer struct {
	consM *kafka.Consumer
	consV *kafka.Consumer
	sl    *slog.Logger
	svc   mailservice.Service
}

func NewKafkaConsumer(sl *slog.Logger, topic string, svc mailservice.Service) Consumer {
	cm, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092,kafka:29092",
		"group.id":          "mail",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	cv, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092,kafka:29092",
		"group.id":          "ver",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	err = cm.Subscribe("mail", nil)
	if err != nil {
		log.Fatal(err)
	}

	err = cv.Subscribe("ver", nil)
	if err != nil {
		log.Fatal(err)
	}

	sl.Info("connected to kafka queue")
	return &kafkaConsumer{
		sl:    sl,
		consM: cm,
		consV: cv,
		svc:   svc,
	}
}

func (c kafkaConsumer) ConsumeVer(ctx context.Context) {
	for {
		msg, err := c.consV.ReadMessage(-1)
		if err == nil {
			c.sl.Info(
				"received message from kafka queue",
				slog.String("topic", *msg.TopicPartition.Topic),
			)
			l := c.sl.With(slog.String("topic", "verify"))
			l.Info("sending request")
			var ver mailservice.Verify
			err := json.Unmarshal(msg.Value, &ver)
			if err != nil {
				l.Error("error unmarshalling kafka message", slog.String("error", err.Error()))
				continue
			}
			err = c.svc.VerifyMail(ctx, ver.VerId)
			if err != nil {
				l.Error("error verifying email", slog.String("error", err.Error()))
				continue

			}
		}
	}
}

func (c kafkaConsumer) ConsumeMail(ctx context.Context) {
	for {
		msg, err := c.consM.ReadMessage(-1)
		if err == nil {
			c.sl.Info(
				"received message from kafka queue",
				slog.String("topic", *msg.TopicPartition.Topic),
			)
			l := c.sl.With(slog.String("topic", "mail"))
			l.Info("sending request")
			var verDto mailservice.VerDto
			err := json.Unmarshal(msg.Value, &verDto)
			if err != nil {
				l.Error("error unmarshalling kafka message", slog.String("error", err.Error()))
				continue
			}
			err = c.svc.SendEmail(ctx, verDto)
			if err != nil {
				l.Error("error sending email", slog.String("error", err.Error()))
				continue

			}
		}
	}
}
