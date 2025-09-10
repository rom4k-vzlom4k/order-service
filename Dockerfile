FROM golang:1.25-alpine

RUN apk add --no-cache \
    bash \
    postgresql-client \
    netcat-openbsd \
    git \
    build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o main ./cmd/order-service

CMD ["/bin/sh", "-c", "\
  echo 'Waiting for Postgres...'; \
  until pg_isready -h db -p 5432; do sleep 2; done; \
  echo 'Postgres is ready!'; \
  echo 'Waiting for Kafka...'; \
  # Ждем пока контейнер Kafka запустится и порт откроется \
  while ! nc -z kafka 9092; do \
    echo 'Kafka not ready, waiting...'; \
    sleep 5; \
  done; \
  # Даем Kafka дополнительное время для инициализации \
  echo 'Kafka port open, waiting for full startup...'; \
  sleep 15; \
  echo 'Starting service'; \
  ./main \
"]