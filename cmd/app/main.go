package main

import (
	"flag"
	"log"
	"micros/internal/config"
	"micros/internal/server"

	"github.com/BurntSushi/toml"
)

var (
	configServPath string //Глобальная переменная, здесь мы будем хранить конфигурацию
)

func init() {
	//Парсер конфига
	flag.StringVar(&configServPath, "configserv-path", "configs/servconfig.toml", "path to config file")
	config.ConfigSetup() //Запуск конфигурации
}

func main() {
	/*
		Запуск сервера и всех программ
	*/

	flag.Parse() //Распарсиваем наши флаги и записывем в переменные

	//Кофиг сервера
	servconfig := server.NewServConfig()

	//Используем библиотеку Берд суши томл, для чтения, парсинга файла и записи в нашу переменную servconfig
	_, err := toml.DecodeFile(configServPath, servconfig)
	if err != nil {
		log.Fatal(err)
	}

	//Точка входа в запуск сервера
	serv := server.New(servconfig)
	//Инициализация натс стриминг(брокер сообщений)

	//Проверка на ощшибки при запуске
	serv.Start()
	serv.Exit()
}
