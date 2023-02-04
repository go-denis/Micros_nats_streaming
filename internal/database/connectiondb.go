package database

import (
	"database/sql"

	_ "github.com/lib/pq" //анонимный инпорт, чтобы все методы не импортировались в наш код
)

type DB struct {
	config *Config
	db     *sql.DB
	cache  *Cache
	name   string
}

// Вспомогательный метод, для возврата указателя на наше хранилище
func NewBD(config *Config) *DB {
	return &DB{
		config: config,
	}
}

// Для обратных вызовов в кеш, сохраняем инстанс *Cache
func (db *DB) SetCahceInstance(cache *Cache) {
	db.cache = cache
}

// Открытие соединения
func (d *DB) Open() error {

	d.name = "PostgreSQL"
	//Открываем соединение с бд
	db, err := sql.Open("postgres", d.config.DatabaseURL)
	//Проверка на ошибки
	if err != nil {
		return err
	}
	//defer db.Close()
	//Пингуем Бд, чтобы проверить дейстивтельно ли соединение установленно
	//по документации, при вызове sql.Open, реальное соединение не создается, а созадется оно лениво
	//Поэтому явно проверяем, что конфиг верен
	if err := db.Ping(); err != nil {
		return err
	}
	//Если все ок, записывем db
	d.db = db

	return nil
}

// Закрытие соединения
func (d *DB) Close() {
	d.db.Close()
}

// Для удобной работы с табличей Заказов
/*func (s *DB) Orders() *OrderReposytory {
	if s.orderRepository != nil {
		return s.orderRepository
	}

	s.orderRepository = &OrderReposytory{
		database: s,
	}
	return s.orderRepository
}
*/
