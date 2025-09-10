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
	GetAllOrders(ctx context.Context) ([]*models.Order, error)
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
        internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, order.OrderUID, order.TrackNumber,
		order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey,
		order.SmID, order.DateCreated, order.OofShard)
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
	query := `
			SELECT o.order_uid,
				o.track_number, 
				o.entry,
				d.name AS delivery_name,
				d.phone,
				d.zip,
				d.city,
				d.address,
				d.region,
				d.email,
				p.transaction,
				p.request_id,
				p.currency,
				p.provider,
				p.amount,
				p.payment_dt,
				p.bank,
				p.delivery_cost,
				p.goods_total,
				p.custom_fee,
				i.chrt_id,
				i.track_number,
				i.price,
				i.rid,
				i.name AS item_name,
				i.sale,
				i.total_price,
				i.nm_id,
				i.brand,
				i.status,
				o.locale,
				o.internal_signature,
				o.customer_id,
				o.delivery_service,
				o.shardkey,
				o.sm_id,
				o.date_created,
				o.oof_shard
			FROM orders o
			JOIN delivery d ON o.order_uid = d.order_uid
			JOIN payment p ON o.order_uid = p.order_uid
			JOIN items i ON o.order_uid = i.order_uid
			WHERE o.order_uid=$1;`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var order *models.Order
	items := []models.Item{}
	for rows.Next() {
		var o models.Order
		var d models.Delivery
		var p models.Payment
		var it models.Item
		err = rows.Scan(
			&o.OrderUID,
			&o.TrackNumber,
			&o.Entry,
			&d.Name,
			&d.Phone,
			&d.Zip,
			&d.City,
			&d.Address,
			&d.Region,
			&d.Email,
			&p.Transaction,
			&p.RequestID,
			&p.Currency,
			&p.Provider,
			&p.Amount,
			&p.PaymentDt,
			&p.Bank,
			&p.DeliveryCost,
			&p.GoodsTotal,
			&p.CustomFee,
			&it.ChrtID,
			&it.TrackNumber,
			&it.Price,
			&it.Rid,
			&it.Name,
			&it.Sale,
			&it.TotalPrice,
			&it.NmID,
			&it.Brand,
			&it.Status,
			&o.Locale,
			&o.InternalSignature,
			&o.CustomerID,
			&o.DeliveryService,
			&o.Shardkey,
			&o.SmID,
			&o.DateCreated,
			&o.OofShard,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		if order == nil {
			o.Delivery = d
			o.Payment = p
			order = &o
		}
		items = append(items, it)
	}
	if order != nil {
		order.Items = items
	}
	return order, nil
}

func (r *orderRepository) GetAllOrders(ctx context.Context) ([]*models.Order, error) {
	query := `
        SELECT 
            o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
            d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
            p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
        FROM orders o
        JOIN delivery d ON o.order_uid = d.order_uid
        JOIN payment p ON o.order_uid = p.order_uid
    `
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех заказов: %w", err)
	}
	defer rows.Close()

	ordersMap := make(map[string]*models.Order)
	for rows.Next() {
		var o models.Order
		var d models.Delivery
		var p models.Payment
		err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerID, &o.DeliveryService, &o.Shardkey, &o.SmID, &o.DateCreated, &o.OofShard,
			&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
			&p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDt, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		o.Delivery = d
		o.Payment = p
		ordersMap[o.OrderUID] = &o
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	// отдельно получает товары для каждого заказа
	for _, order := range ordersMap {
		itemsQuery := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid=$1;`
		itemRows, err := r.db.Query(ctx, itemsQuery, order.OrderUID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения товаров для заказа %s: %w", order.OrderUID, err)
		}
		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err = itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
			if err != nil {
				return nil, fmt.Errorf("ошибка сканирования товара: %w", err)
			}
			items = append(items, item)
		}
		itemRows.Close() // необходимо закрыть, чтобы не было утечки ресурсов
		order.Items = items
	}

	var orders []*models.Order
	for _, order := range ordersMap {
		orders = append(orders, order)
	}

	return orders, nil
}
