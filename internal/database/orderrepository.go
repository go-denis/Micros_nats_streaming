package database

import (
	"context"
	"errors"
	"log"
	"micros/internal/models"
)

/*
*
*
*Добавление заказа
*
*
 */
func (r *DB) InsertOrder(o models.Orders) (int64, error) {

	var lastInsertId int64 = 1

	var itemsIds []int64 = []int64{} //id Items Для загрузки в Заказы
	//var dbSql *DB
	//Запуск транзакции
	transact, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer transact.Rollback() //Прерываем/откатываем транзакцию

	// добавление в таблицу Item
	for _, item := range o.Items {
		err := transact.QueryRowContext(context.Background(), `INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand,status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status).Scan(&lastInsertId)
		if err != nil {
			log.Printf("Не удается вставить значения в таблицу Item: %v\n", err)
			return -1, err
		}
		itemsIds = append(itemsIds, lastInsertId)
	}

	// Добавление в таблицу Payment
	err = transact.QueryRowContext(context.Background(), `INSERT INTO payment (transaction, request_id, currency, provider, amount, 
		payment_dt, bank, delivery_cost, goods_total, custom_fee) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank,
		o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee).Scan(&lastInsertId)

	if err != nil {
		log.Printf("Не удается вставить значения в таблицу Payment: %v\n", err)
		return -1, err
	}
	paymentIdFk := lastInsertId

	// Добавление в таблицу Delivery
	err = transact.QueryRowContext(context.Background(), `INSERT INTO delivery (name, phone, zip, city, address, region, email) 
	values ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City,
		o.Delivery.Address, o.Delivery.Region, o.Delivery.Email).Scan(&lastInsertId)

	if err != nil {
		log.Printf("Не удается вставить значения в таблицу Payment: %v\n", err)
		return -1, err
	}
	DeliverydFk := lastInsertId

	// Добавление Order
	err = transact.QueryRowContext(context.Background(), `INSERT INTO orders (order_uid, track_number, entry, delivery_id_fk, payment_id_fk, Locale, 
		internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) values 
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,$12, $13)
			RETURNING id`,
		o.OrderUID, o.TrackNumber, o.Entry, paymentIdFk, DeliverydFk, o.Locale, o.InternalSignature, o.CustomerID, o.DeliveryService,
		o.Shardkey, o.SmID, o.DateCreated, o.OofShard).Scan(&lastInsertId)
	if err != nil {
		log.Printf("Не удается вставить значения в таблицу Заказов %v\n", err)
		return -1, err
	}
	orderIdFk := lastInsertId

	// Разрешение связей один-ко-многим для Order и Order.Items[]
	for _, itemId := range itemsIds {
		_, err := transact.ExecContext(context.Background(), `INSERT INTO orders_items (order_id_fk, item_id_fk) values ($1, $2)`,
			orderIdFk, itemId)
		if err != nil {
			log.Printf("Не удается вставить значения в таблицу Заказы-Айтемсы: %v\n", err)
			return -1, err
		}
	}

	err = transact.Commit() //Фиксируем транзакцию
	if err != nil {
		return 0, err
	}

	log.Printf("Ураааааа! Этот заказ наконеч-то добавился в базу!!!!\n")

	return lastInsertId, nil
}

/*
*
*
*Получение Order из базы...
*
*
 */
func (db *DB) GetOrderInDB(orderInID int64) (models.Orders, error) {
	var o models.Orders
	var (
		payment_id_fk  int64
		delivery_id_fk int64
	)
	// Сбор данных об Order
	err := db.db.QueryRowContext(context.Background(), `SELECT order_uid, track_number, entry, delivery_id_fk, payment_id_fk, locale, 
	internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE id = $1`,
		orderInID).Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &delivery_id_fk, &payment_id_fk, &o.Locale, &o.InternalSignature,
		&o.CustomerID, &o.DeliveryService, &o.Shardkey, &o.SmID, &o.DateCreated, &o.OofShard)
	if err != nil {
		return o, errors.New("не удается получить заказы из базы данных")
	}

	// Сбор данных о Payment
	err = db.db.QueryRowContext(context.Background(), `SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, 
	delivery_cost, goods_total, custom_fee FROM payment WHERE id = $1`, payment_id_fk).Scan(&o.Payment.Transaction, &o.Payment.RequestID,
		&o.Payment.Currency, &o.Payment.Provider, &o.Payment.Amount, &o.Payment.PaymentDt, &o.Payment.Bank, &o.Payment.DeliveryCost,
		&o.Payment.GoodsTotal, &o.Payment.CustomFee)
	if err != nil {
		log.Printf("%v: orderrepository GetOrderInDB() не удается получить информацию об оплате из бд: %v\n", db.name, err)
		return o, errors.New("orderrepository GetOrderInDB()не удается получить информацию об оплате из бд")
	}

	// Сбор данных о Delivery
	err = db.db.QueryRowContext(context.Background(), `SELECT name, phone, zip, city, address, region, email FROM delivery WHERE id = $1`,
		payment_id_fk).Scan(&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City, &o.Delivery.Address, &o.Delivery.Region,
		&o.Delivery.Email)
	if err != nil {
		log.Printf("%v: Не удается получить информацию о доставке из бд: %v\n", db.name, err)
		return o, errors.New("orderrepository GetOrderInDB() не удается получить информацию о доставке из бд")
	}

	// Сбор всех ItemsID для Order
	rowsItems, err := db.db.QueryContext(context.Background(), "SELECT item_id_fk FROM orders_items WHERE order_id_fk = $1", orderInID)
	if err != nil {
		return o, errors.New("orderrepository GetOrderInDB() не удается получить ПРЕДМЕТЫ(items) из базы данных")
	}
	defer rowsItems.Close()

	// Цикл по списку ItemsID
	var itemID int64
	for rowsItems.Next() {
		var item models.Items_DB
		if err := rowsItems.Scan(&itemID); err != nil {
			return o, errors.New("orderrepository GetOrderInDB() не удается взять item id из базы данных")
		}
		// Сбор данных об Items
		err = db.db.QueryRowContext(context.Background(), `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand,
		status
		FROM items WHERE id = $1`, itemID).Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			return o, errors.New("orderrepository GetOrderInDB() не удается взять itemsss из базы данных")
		}
		o.Items = append(o.Items, item)
	}
	return o, nil
}
