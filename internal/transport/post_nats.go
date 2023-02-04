package transport

import (
	//"micros/internal/database"
	//"github.com/nats-io/stan.go"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"micros/internal/models"
	"os"

	//"github.com/nats-io/stan.go"
	stan "github.com/nats-io/stan.go"
)

const path string = "./internal/transport/data" //Директория для поиска сообщений

var orders models.Orders //orders...

// Структура отправителя
type Publish struct {
	sc   *stan.Conn
	name string
}

// Возвращаем ссылку на отправителя (метод для дальнейших взаиможействий) чтоб было удобно работать
func NewSend(conn *stan.Conn) *Publish {
	return &Publish{
		sc:   conn,
		name: "Sender",
	}
}

func (pub *Publish) SendNats() {
	//Так как может быть много сообщений и все их надо отправить, то будем искать их в папке data
	//Не особо разобрался как и в каком формате приходят сообщения, поэтому предпологаю
	//Что другие сервисы закидывают в папку в формате json
	fmt.Println("Ищу сообщения ...")
	//Записывем нашу папку
	lst, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	//Проходимся по всем файлам и отправляем их
	for _, val := range lst {
		//Проверка на директорию
		if val.IsDir() {
			fmt.Printf("Директория [%s] не будет считана, кладите файл json в корень папки data\n", val.Name())
		} else {
			//Открываем json файл
			jsonFile, err := os.Open("./internal/transport/data/" + val.Name())
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Файл успешно открыт " + val.Name())
			defer jsonFile.Close()
			//Считываем json файл
			byteValue, _ := ioutil.ReadAll(jsonFile)
			//Парсим данные в структуру
			json.Unmarshal(byteValue, &orders)

		}
	}
	//Заносим в json
	orderData, err := json.Marshal(orders)

	if err != nil {
		log.Printf("%s: json.Marshal error: %v\n", pub.name, err)
	}
	/*
		Базовый API публикации (Publish(subject, payload)) является синхронным; он не возвращает
		управление вызывающей стороне до тех пор, пока потоковый сервер NATS не подтвердит получение сообщения.
		Для выполнения этого генерируется NUID для сообщения о создании, и клиентская библиотека ожидает подтверждения
		публикации от сервера с соответствующим NUID, прежде чем она вернет управление вызывающей стороне, возможно, с ошибкой,
		указывающей, что операция не была успешной из-за какой-либо проблемы с сервером или ошибки авторизации.
	*/
	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			log.Printf("Предупреждение: ошибка публикации идентификатора msg %s: %v\n", ackedNuid, err.Error())
		} else {
			log.Printf("Получено подтверждение для идентификатора msg %s\n", ackedNuid)
		}
	}

	//Публикуем данные соединение через структуру
	nuid, err := (*pub.sc).PublishAsync(os.Getenv("NATS_GETTER"), orderData, ackHandler)
	if err != nil {
		log.Printf("Ошибка публикации сообщения об ошибке %s: %v\n", nuid, err.Error())
	}

}
