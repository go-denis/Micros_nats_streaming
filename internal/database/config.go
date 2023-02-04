package database

import (
	"os"
)

type Config struct {
	DatabaseURL string `toml: "database_url"`
}

// Метод, который возвращает указатель на наш конфиг
func NewConfigDB() *Config {
	return &Config{
		DatabaseURL: "user=" + os.Getenv("DB_NAME") +
			" password=" + os.Getenv("DB_PASSWORD") +
			" dbname=" + os.Getenv("DB_NAME") +
			" sslmode=" + os.Getenv("DB_SSL_MODE") + "",
	}

}
