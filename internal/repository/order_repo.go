package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rom4k-vzlom4k/oder-service/pkg/models"
)

type OrderRepository interface {
	Save(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
}

type orderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Save(ctx context.Context, order *models.Order) error {
	tx, err := r.db.Begin(ctx) // Получает соединение из пула и запускает транзакцию
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx) // откат если commit не вызван
	// Добавляет заказ
	_, err = tx.Exec(ctx,
		`INSERT INTO orders (order_uid, track_number, entry, locale, 
        internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, off_shard)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, order.OrderUID, order.TrackNumber,
		order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey,
		order.SmID, order.DateCreated, order.OffShard)
	if err != nil {
		return fmt.Errorf("insert order failed: %w", err)
	}
	// Добавляет доставку
	_, err = tx.Exec(ctx,
		`INSERT INTO delivery(order_uid, name, phone, zip, city, address, region, email) 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("insert delivery failed: %w", err)
	}

	// Добавляет платеж
	_, err = tx.Exec(ctx,
		`INSERT INTO payment(transaction, order_uid, request_id, currency, provider, amount, payment_dt, bank,
		delivery_cost, goods_total, custom_fee) 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		order.Payment.Transaction, order.OrderUID, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal,
		order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("insert payment failed: %w", err)
	}

	// Добавляет items
	for _, item := range order.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, 
			nm_id, brand, status)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size,
			item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("insert item failed: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return err
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	return nil, nil
}
