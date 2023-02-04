package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"micros/internal/models"
	"os"
	"strconv"
	"sync"
)

type Cache struct {
	buffer  map[int64]models.Orders
	queue   []int64
	bufSize int
	pos     int
	DBInst  *DB
	name    string
	mutex   *sync.RWMutex
}

func NewCache(db *DB) *Cache {
	csh := Cache{}
	csh.InitCache(db)
	return &csh
}

// Инициализация кеша...
func (ch *Cache) InitCache(db *DB) {

	ch.DBInst = db
	db.SetCahceInstance(ch)
	ch.name = "Кеш"
	ch.mutex = &sync.RWMutex{}

	// Установка размера кеша
	bufSize, err := strconv.Atoi(os.Getenv("CACHE_SIZE"))
	if err != nil {
		log.Printf("%s: Пожалуйста, установите размер кеша по умолчанию 10\n", ch.name)
		bufSize = 10
	}

	ch.bufSize = bufSize
	ch.buffer = make(map[int64]models.Orders, ch.bufSize)
	ch.queue = make([]int64, ch.bufSize)

	// Восстанавление кеша из базы данных, если он есть в бд
	ch.RecoveCacheInDatabase()
}

// Восстанавливаем кеш из базы данных: читаем из файла содержимое кеша
func (ch *Cache) RecoveCacheInDatabase() {
	log.Printf("%v: Проверка и загрузка кеша из базы данных\n", ch.name)
	buf, queue, pos, err := ch.DBInst.LoadCacheRecover(ch.bufSize)
	if err != nil {
		log.Printf("%s: Не удается загрузить из базы данных кеш, либо он пустой: %v\n", ch.name, err)
		return
	}

	// Проверяем, не заполнили ли буфер полностью. Если да - сбрасываем указатель на начало циклической очереди
	if pos == ch.bufSize {
		pos = 0
	}

	ch.mutex.Lock() //Блокируем rw для записи, до тех пор пока не станет доступно для записи
	ch.buffer = buf
	ch.queue = queue
	ch.pos = pos
	ch.mutex.Unlock()
	log.Printf("%s: Загружен в базу: следующий в очереди: %v, далее в очереди: %v", ch.name, ch.queue, ch.pos)
}

// Сохранение в кеш после успешного добавления Order в БД
func (ch *Cache) SetOrder(orderInID int64, o models.Orders) {
	if ch.bufSize > 0 {
		ch.mutex.Lock()
		// сохраняем в циклическую очередь новый orderId (если на позиции pos будет Order, он будет перезаписан)
		ch.queue[ch.pos] = orderInID
		ch.pos++
		if ch.pos == ch.bufSize {
			ch.pos = 0
		}

		// сохраняем в буфер новый Order
		ch.buffer[orderInID] = o
		ch.mutex.Unlock()

		// сохраняем в таблицу Cache в БД новый OrderID - для восстановления кеша после сбоя
		ch.DBInst.SendOrderIDToCache(orderInID)
		log.Printf("%s: Заказ успешно добавлен в Кеш, позиция заказа в очереди равна: %v\n", ch.name, ch.pos)
	} else {
		log.Printf("Кеш отключили, bufSize = 0 (поправьте файл config.go)\n")
	}

	log.Printf("%s: В очереди: %v, следующий в очереди: %v", ch.name, ch.queue, ch.pos)
}

// Получаем Order по ID из кеша. Преобразование в модели для выдачи
func (ch *Cache) GetOrderCache(orderInID int64) (*models.Orders_issuance, error) {
	var ord_issu *models.Orders_issuance = &models.Orders_issuance{}
	var o models.Orders
	var err error

	ch.mutex.RLock()
	// проверка в кеше. Если нет - идем в базу
	o, isExist := ch.buffer[orderInID]
	ch.mutex.RUnlock()

	if isExist {
		log.Printf("%s: Orders_issuance (id:%d) взят из кеша!\n", ch.name, orderInID)
	} else {
		// запрос Order к базе данных
		o, err = ch.DBInst.GetOrderInDB(orderInID)
		if err != nil {
			log.Printf("%s: GetOrderCache(): ошибка получения Order: %v\n", ch.name, err)
			return ord_issu, err
		}
		// Сохранение в кеш
		ch.SetOrder(orderInID, o)
		log.Printf("%s: Заказ %d, взят из бд и сохранен в кеш!\n", ch.name, orderInID)
	}

	// Преобразование к модели для выдачи
	ord_issu.OrderUID = o.OrderUID
	ord_issu.TrackNumber = o.TrackNumber
	ord_issu.Entry = o.Entry
	ord_issu.Locale = o.Locale
	ord_issu.InternalSignature = o.InternalSignature
	ord_issu.CustomerID = o.CustomerID
	ord_issu.DeliveryService = o.DeliveryService
	ord_issu.Shardkey = o.Shardkey
	ord_issu.SmID = o.SmID
	ord_issu.DateCreated = o.DateCreated
	ord_issu.OofShard = o.OofShard

	//ord_issu.TotalPrice = o.GetTotalPrice()

	return ord_issu, nil
}

func (ch *Cache) Finish() {
	log.Printf("Закрываем кеш")
	ch.DBInst.ClearCache()
	log.Printf("Кеш закрыт")
}

// Загрузка объектов Orders (кеша) при его восстановлении
func (db *DB) LoadCacheRecover(bufSize int) (map[int64]models.Orders, []int64, int, error) {

	buffer := make(map[int64]models.Orders, bufSize)
	queue := make([]int64, bufSize)
	var queueInd int

	// Выбираем все OrderID для нашей программы (APP_KEY) из таблицы кеша
	query := fmt.Sprintf("SELECT order_id FROM cache WHERE app_key = '%s' ORDER BY id DESC LIMIT %d", os.Getenv("APP_KEY"), bufSize)
	rows, err := db.db.QueryContext(context.Background(), query)
	if err != nil {
		log.Printf("Кеш не удается получить id Заказов из базы данных: %v\n", err)
	}
	defer rows.Close()

	// Цикл по списку OrderID
	var orderInID int64
	for rows.Next() {
		if err := rows.Scan(&orderInID); err != nil {
			log.Printf("Не удается получить id Заказа(ов) из базы данных: %v\n", err)
			return buffer, queue, queueInd, errors.New("Не удается получить id Заказа(ов) из базы данных")
		}
		// сохраняем в очередь в порядке добавления в кеш (перед тем, как программа некоректно завершилась)
		queue[queueInd] = orderInID
		queueInd++

		o, err := db.GetOrderInDB(orderInID) //Достаем заказы
		if err != nil {
			log.Printf("Кеш: Не могу получить Заказы из базы данных: %v\n", err)
			continue
		}
		buffer[orderInID] = o //Заносим в буфер id заказов
	}

	if queueInd == 0 {
		return buffer, queue, queueInd, errors.New("Кеш пуст!!!")
	}

	// переиндексация - в начале queue - "старый" кеш, в конце очереди - "новый". После запроса (самого первого в этой функции) - наоборот
	// Пример: после выполнения кода выше очередь содержит список Order ID: queue = [109 108 107 106 105 104 0 0 0 0],
	// Поскольку id=109 - это более "свежие" данные, то правильный порядок в очереди должен быть такой:
	// queue = [104 105 106 107 108 109 0 0 0 0]
	for i := 0; i < int(queueInd/2); i++ {
		queue[i], queue[queueInd-i-1] = queue[queueInd-i-1], queue[i]
	}

	return buffer, queue, queueInd, nil
}

// Записываем кеш в базу данных, для тех случаев, когда программа не корретно завершает свою работу
func (db *DB) SendOrderIDToCache(orderInID int64) {
	db.db.QueryRowContext(context.Background(), `INSERT INTO cache (order_id, app_key) VALUES ($1, $2)`, orderInID, os.Getenv("APP_KEY"))
	log.Printf("id Заказа успешно добавлен в кеш!\n")
}

// Если программа корректно завершила свою работу, то удаляем кеш из базы данных
func (db *DB) ClearCache() {
	//Удаление без позврата каких-либо значений
	_, err := db.db.ExecContext(context.Background(), `DELETE FROM cache WHERE app_key = $1`, os.Getenv("APP_KEY"))
	if err != nil {
		log.Printf("Ошибка очистки кеша в базе данных: %s\n", err)
	}
	log.Printf("Кеш успешно очишен из базы данных!\n")
}
