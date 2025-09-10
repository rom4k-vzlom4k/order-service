package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rom4k-vzlom4k/oder-service/internal/handler"
	"github.com/rom4k-vzlom4k/oder-service/internal/kafka"
	"github.com/rom4k-vzlom4k/oder-service/internal/repository"
	"github.com/rom4k-vzlom4k/oder-service/internal/service"
)

func main() {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL not set in environment")
	}

	pool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatalf("ошибка подключения к БД: %v", err)
	}
	defer pool.Close()

	// проверяет соединение с БД
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("ошибка пинга БД: %v", err)
	}
	log.Println("Успешное подключение к базе данных")

	repo := repository.NewOrderRepository(pool)
	orderService := service.NewOrderService(repo)

	// восстанавливает кэш из бд
	if err := orderService.RestoreCache(context.Background()); err != nil {
		log.Printf("ошибка восстановления кэша: %v", err)
	}

	log.Println("Ожидание подключения к Kafka...")
	producer, err := waitForKafkaConnection()
	if err != nil {
		log.Fatalf("Не удалось подключиться к Kafka: %v", err)
	}
	defer producer.Close()

	consumer, err := kafka.NewConsumer([]string{"kafka:9092"}, "order-group", "order-topic", orderService)
	if err != nil {
		log.Fatalf("ошибка создания консюмера: %v", err)
	}
	defer consumer.Close()

	go consumer.Run()

	orderHandler := handler.NewOrderHandler(orderService, producer)

	http.Handle("/", http.FileServer(http.Dir("./static/")))

	http.HandleFunc("/order/", orderHandler.GetByOrderId)
	http.HandleFunc("/save/", orderHandler.SaveOrder)

	log.Println("Сервер запущен на :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func waitForKafkaConnection() (kafka.Producer, error) {
	maxAttempts := 30
	attempt := 1

	for attempt <= maxAttempts {
		log.Printf("Попытка подключения к Kafka (%d/%d)...", attempt, maxAttempts)

		producer, err := kafka.NewProducer([]string{"kafka:9092"})
		if err == nil {
			log.Println("Успешное подключение к Kafka")
			return producer, nil
		}

		log.Printf("Ошибка подключения к Kafka: %v", err)
		time.Sleep(5 * time.Second)
		attempt++
	}

	return nil, fmt.Errorf("не удалось подключиться к Kafka после %d попыток", maxAttempts)
}
