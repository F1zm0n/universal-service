package receiver

import (
	"bytes"
	"fmt"
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
	fmt.Println("connected to kafka")
	return &kafkaConsumer{
		sl:   sl,
		cons: c,
	}
}

func (c kafkaConsumer) Consume() {
	for true {
		msg, err := c.cons.ReadMessage(-1)
		if err == nil {
			c.sl.Info("received message from kafka queue")
			c.sl.Info("topicname", slog.String("topic", *msg.TopicPartition.Topic))
			switch *msg.TopicPartition.Topic {
			case "email":
				err := sendReq(http.MethodPost, "http://mailer:5001/mail", msg.Value)
				if err != nil {
					c.sl.Error("error sending req", slog.String("err", err.Error()))
					continue
				}
				c.sl.Info("sent the request")
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
		fmt.Println(err)
		return ErrSendingReq
	}
	if res.StatusCode != 200 {
		return ErrStatusIsNot200
	}
	return nil
}
