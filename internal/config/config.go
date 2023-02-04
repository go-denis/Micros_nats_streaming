package config

import (
	"os"
)

// Конфигурация занесенна в переменные среды
func ConfigSetup() {

	//Настройки сервера
	os.Setenv("DB_BINN_ADDR", ":8080")
	os.Setenv("DB_LOG_LEVEL", "debug") //Уровень ошибок

	// Настройки базы
	os.Setenv("DB_USERNAME", "postgres")
	os.Setenv("DB_PASSWORD", "IrzsmYTa4X")
	os.Setenv("DB_HOST", "5432")
	os.Setenv("DB_NAME", "postgres")
	os.Setenv("DB_SSL_MODE", "disable")
	os.Setenv("DB_POOL_MAXCONN", "5")
	os.Setenv("DB_POOL_MAXCONN_LIFETIME", "300")

	os.Setenv("DELIVERY", "meest")

	// Настройки НАСТ стриминг
	os.Setenv("NATS_URL", "nats://localhost:4222")
	os.Setenv("NATS_CLUSTER_ID", "test-cluster")
	os.Setenv("NATS_CLIENT_ID", "WB")
	os.Setenv("NATS_GETTER", "denis") //Тема сообщения
	os.Setenv("NATS_DURABLE_NAME", "nats")
	os.Setenv("NATS_WAIT_SECONDS", "30")

	// Настройки кеша
	os.Setenv("CACHE_SIZE", "10")
	os.Setenv("APP_KEY", "APP_1")

}
