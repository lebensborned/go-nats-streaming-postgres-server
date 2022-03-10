package db

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Repository describes a repository of app
type Repository interface {
	FindById(orderId string) (*Order, error)
	Insert(order *Order) error
	Fill() error
	Init(user, password, dbname string) error
}

// Cache + db handler abstraction
type Repo struct {
	cache map[string]*Order
	mu    sync.RWMutex
	db    *sqlx.DB
}

// NewRepo creates a object of repository
func NewRepo() *Repo {
	return &Repo{
		cache: make(map[string]*Order),
		mu:    sync.RWMutex{},
	}
}

// Init a connection to db
func (r *Repo) Init(user, password, dbname string) error {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("cant open connection to db: %v", err)
	}
	if err = db.Ping(); err != nil {
		return fmt.Errorf("unsuccesfully db ping: %v", err)
	}
	r.db = db
	err = r.Fill()
	if err != nil {
		return fmt.Errorf("cant fill repo cache in init(): %v", err)
	}
	return nil
}

// Get order from cache
func (r *Repo) FindById(orderId string) (*Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, ok := r.cache[orderId]
	if !ok {
		return nil, fmt.Errorf("cant find order in cache")
	}
	return order, nil
}

// Fill cache from db
func (r *Repo) Fill() error {
	rows, err := r.db.Queryx("select  * from orders")
	if err != nil {
		return fmt.Errorf("cant process query to fill cache: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		order := Order{}
		err := rows.StructScan(&order)
		if err != nil {
			return err
		}
		rowsPayment, err := r.db.Queryx("select transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee from payments where fk_payments_order=$1", &order.OrderUid)
		if err != nil {
			return err
		}
		defer rowsPayment.Close()
		payment := Payment{}
		for rowsPayment.Next() {
			err = rowsPayment.StructScan(&payment)
			if err != nil {
				return err
			}
		}

		rowsDelivery, err := r.db.Queryx("select name, phone, zip, city, address, region, email from delivery where fk_delivery_order=$1", &order.OrderUid)
		if err != nil {
			return err
		}
		defer rowsDelivery.Close()
		delivery := Delivery{}
		for rowsDelivery.Next() {
			err = rowsDelivery.StructScan(&delivery)
			if err != nil {
				return err
			}
		}

		rowsItems, err := r.db.Queryx("select chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status from items where fk_items_order=$1", &order.OrderUid)
		if err != nil {
			return err
		}
		defer rowsItems.Close()

		var items []Item
		for rowsItems.Next() {
			item := Item{}
			err = rowsItems.StructScan(&item)
			if err != nil {
				return err
			}
			items = append(items, item)
		}

		t := &Order{order.OrderUid, order.TrackNumber, order.Entry, delivery,
			payment, items, order.Locale, order.InternalSignature,
			order.CustomerId, order.DeliveryService, order.Shardkey,
			order.SmId, order.DateCreated, order.OofShard}
		r.mu.Lock()
		r.cache[order.OrderUid] = t
		r.mu.Unlock()
	}
	return nil
}

// Insert a new order into repository
func (r *Repo) Insert(order *Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	// if something goes wrong - just rollback.
	defer tx.Rollback()
	tx.Exec(`insert into orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		order.OrderUid, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerId,
		order.DeliveryService, order.Shardkey, order.SmId, order.DateCreated, order.OofShard)

	tx.Exec(`insert into payments (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, fk_payments_order) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		order.Payment.Transaction, order.Payment.RequestId, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal,
		order.Payment.CustomFee, order.OrderUid)
	tx.Exec(`insert into delivery (name, phone, zip, city, address, region, email, fk_delivery_order) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email, order.OrderUid)
	for _, item := range order.Items {
		tx.Exec(`insert into items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status, fk_items_order) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			item.ChrtId, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId,
			item.Brand, item.Status, order.OrderUid)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	r.cache[order.OrderUid] = order
	return nil
}
