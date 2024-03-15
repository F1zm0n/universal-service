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
		"bootstrap.servers": "localhost:9092,kafka:9092",
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
	return &kafkaConsumer{
		sl:   sl,
		cons: c,
	}
}

func (c kafkaConsumer) Consume() {
	for {
		msg, err := c.cons.ReadMessage(-1)
		if err != nil {
			continue
		}
		c.sl.Info("received message from kafka queue")
		switch *msg.TopicPartition.Topic {
		case mailerTopic:
			err := sendReq(http.MethodPost, mailerUrl, msg.Value)
			if err != nil {
				continue
			}
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
