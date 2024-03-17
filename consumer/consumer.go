package receiver

import (
	"bytes"
	"log"
	"log/slog"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/spf13/viper"
)

var (
	mailerUrl   = viper.GetString("req.http.urls.mail")
	mailerTopic = viper.GetString("kafka.topics.mailer")
)

type Consumer interface {
	Consume()
}

type kafkaConsumer struct {
	cons *kafka.Consumer
	sl   *slog.Logger
}

func NewKafkaConsumer(sl *slog.Logger, topics []string) Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		log.Fatal(err)
	}

	myTop, err := c.Subscription()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("subscribed to ", myTop)

	sl.Info("connected to kafka queue")
	return &kafkaConsumer{
		sl:   sl,
		cons: c,
	}
}

func (c kafkaConsumer) Consume() {
	for true {
		msg, err := c.cons.ReadMessage(-1)
		if err == nil {
			c.sl.Info(
				"received message from kafka queue",
				slog.String("topic", *msg.TopicPartition.Topic),
			)
			switch *msg.TopicPartition.Topic {
			case "email":
				l := c.sl.With(slog.String("topic", "email"))
				l.Info("sending request")
				err := sendReq(http.MethodPost, "http://mailer:5001/mail", msg.Value)
				if err != nil {
					l.Error("error sending req", slog.String("err", err.Error()))
					continue
				}
			case "verify":
				l := c.sl.With(slog.String("topic", "verify"))
				l.Info("sending request")
				err := sendReq(http.MethodPost, "http://mailer:5001/verify", msg.Value)
				if err != nil {
					l.Error("error sending req", slog.String("err", err.Error()))
					continue
				}
			case "register":
				l := c.sl.With(slog.String("topic", "register"))
				l.Info("sending request")
				err := sendReq(http.MethodPost, "http://auth:5001/register", msg.Value)
				if err != nil {
					l.Error("error sending req", slog.String("err", err.Error()))
					continue
				}
			}
		} else if !err.(kafka.Error).IsTimeout() {
			c.sl.Info("error consuming from kafka")
		}
	}
}

func sendReq(method string, url string, data []byte) error {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return ErrSendingReq
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ErrSendingReq
	}
	if res.StatusCode != 200 {
		return ErrStatusIsNot200
	}
	return nil
}
