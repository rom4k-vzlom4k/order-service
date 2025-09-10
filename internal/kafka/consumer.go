package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rom4k-vzlom4k/oder-service/internal/service"
	"github.com/rom4k-vzlom4k/oder-service/pkg/models"
	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Run()
	Close() error
}

type consumer struct {
	reader  *kafka.Reader
	service service.OrderService
}

func NewConsumer(brokers []string, groupID string, topic string, service service.OrderService) (Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &consumer{
		reader:  reader,
		service: service,
	}, nil
}

func (c *consumer) Run() {
	log.Println("Запуск консюмера")
	for {
		msg, err := c.reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("ошибка чтения сообщения: %v", err)
			continue
		}

		var order models.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("ошибка десериализации сообщения: %v", err)
			continue
		}

		if err := c.service.Save(context.Background(), &order); err != nil {
			log.Printf("ошибка сохранения заказа: %v", err)
		} else {
			log.Printf("Заказ %s успешно сохранен из Kafka", order.OrderUID)
		}
	}
}

func (c *consumer) Close() error {
	return c.reader.Close()
}
