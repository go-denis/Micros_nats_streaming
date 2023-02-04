package server

import (
	"micros/internal/database"
	"os"
)

/*
Конфигурация сервера, чтобы не хардкодить порты и тд.
Конфигурировать будем с помощью  toml
PS Я не понял в чем дело, но toml не  работает нормально, пришлось использовать os
*/

// Конфигурация

type ConfigServ struct {
	BindAddr string           `toml: "bind_addr"` //Порт
	LogLevel string           `toml:	"log_level"` //Логи
	DataBase *database.Config //База данных
}

// Для удобства, отдает инициализированный конфиг с дефольтыми параметрами, которые нас устраивают
func NewServConfig() *ConfigServ {
	return &ConfigServ{
		BindAddr: os.Getenv("DB_BINN_ADDR"),
		LogLevel: os.Getenv("DB_LOG_LEVEL"),
		DataBase: database.NewConfigDB(),
		//Nats: APIServer.ConfigNats(),
	}
}
