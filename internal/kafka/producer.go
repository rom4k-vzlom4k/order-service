package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	WriteMessage(ctx context.Context, value []byte) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) (Producer, error) {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  "order-topic",
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	// проверка подключения
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		log.Printf("Warning: Kafka connection test failed: %v", err)
	} else {
		conn.Close()
		log.Println("Kafka producer connected successfully")
	}

	return &producer{
		writer: writer,
	}, nil
}

func (p *producer) WriteMessage(ctx context.Context, value []byte) error {
	log.Printf("Sending message to Kafka: %s", string(value))

	msg := kafka.Message{
		Value: value,
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("Failed to produce message: %v", err)
		return err
	}

	log.Println("Message sent to Kafka successfully")
	return nil
}

func (p *producer) Close() error {
	return p.writer.Close()
}
