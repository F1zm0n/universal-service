package prodservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type EmailPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type VerifyPayload struct {
	VerId uuid.UUID `json:"ver_id"`
}

func New(logger log.Logger) Service {
	var svc Service
	{
		svc = NewKafkaService()
		svc = LoggingMiddleware(logger)(svc)
		svc = InstrumentingMiddleware()(svc)
	}

	return svc
}

var ErrProducingToKafka = errors.New("internal server error k")

type Service interface {
	ProduceMail(ctx context.Context, email, password string) error
	ProduceVer(ctx context.Context, verId uuid.UUID) error
	ProduceRegister(ctx context.Context, email, password string) error
}

type kafkaService struct {
	conn *kafka.Producer
}

func NewKafkaService() Service {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:9092,localhost:9092",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("connected to kafka")
	return &kafkaService{
		conn: p,
	}
}

func (s kafkaService) ProduceMail(ctx context.Context, email, password string) error {
	data := EmailPayload{
		Email:    email,
		Password: password,
	}
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.produceData(ctx, viper.GetString("kafka.topics.mailer"), j)
}

func (s kafkaService) ProduceVer(ctx context.Context, verId uuid.UUID) error {
	data := VerifyPayload{
		VerId: verId,
	}
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.produceData(ctx, "verify", j)
}

func (s kafkaService) ProduceRegister(ctx context.Context, email, password string) error {
	data := EmailPayload{
		Email:    email,
		Password: password,
	}
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.produceData(ctx, viper.GetString("kafka.topics.register"), j)
}

func (s kafkaService) produceData(_ context.Context, topic string, data []byte) error {
	err := s.conn.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          data,
	}, nil)
	if err != nil {
		return ErrProducingToKafka
	}
	return nil
}
