package transport

import (
	"log"
	"micros/internal/database"
	"os"
	"time"

	stan "github.com/nats-io/stan.go"
)

type NatsHandler struct {
	conn   *stan.Conn //Соединение
	getter *Getter    //подписчик
	post   *Publish   //Публикатор
	name   string     //Назавние
	err    bool       //Ошибки
}

// Создаем новое подключение
func NewNatsHandler(db *database.DB) *NatsHandler {
	nh := NatsHandler{}
	nh.ConfigNatsInit(db)
	return &nh
}

// Инициализация Подписки и Отправки
func (nh *NatsHandler) ConfigNatsInit(db *database.DB) {
	nh.name = "Обработчик NATS" //Название
	err := nh.Connect()         //Подключение к натс

	if err != nil {
		nh.err = true
		log.Printf("%s: Обработчик NATS ошибка: %s", nh.name, err)
	} else {
		log.Printf("Nats-streaming запущен")
		//Если все ок, то запускаем Подписку и отправляем сообщение
		nh.getter = NewSubscriber(db, nh.conn)
		nh.getter.Getter()

		nh.post = NewSend(nh.conn)
		nh.post.SendNats()
	}
}

// Подключение к NATS
func (nh *NatsHandler) Connect() error {

	//Создаем обратный вызов чтобы соединение не отваливалось само по себе
	//conn, err := nats.Connect("test-cluster", "denis")

	conn, err := stan.Connect(os.Getenv("NATS_CLUSTER_ID"), os.Getenv("NATS_CLIENT_ID"),
		stan.ConnectWait(time.Second*4),
		stan.PubAckWait(time.Second*4),
		stan.NatsURL(os.Getenv("NATS_URL")),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}),
		stan.MaxPubAcksInflight(25))

	/*stan.Connect(os.Getenv("NATS_CLUSTER_ID"), os.Getenv("NATS_CLIENT_ID"),
		stan.NatsURL(os.Getenv("NATS_URL")),
		stan.NatsOptions(
			nats.ReconnectWait(time.Second*4),
			nats.Timeout(time.Second*4),
		),
		stan.Pings(10, 5), // Отправляйте пинги каждые 10 секунд и завершайте работу с ошибкой после 5 пингов без какого-либо ответа.
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("%s: Связь потеряна, из-за: %v", nh.name, reason)
		}),
	)*/
	if err != nil {
		log.Printf("%s: Не могу подключиться: %v.\n", nh.name, err)
		return err
	}
	nh.conn = &conn

	log.Printf("%s: Отлично, соединение установленно!", nh.name)
	return nil
}

// Завершение работы с NATS
func (nh *NatsHandler) Finish() {
	if !nh.err {
		log.Printf("%s: Nats завершает свою работу...", nh.name)
		nh.getter.UnGetter()
		(*nh.conn).Close()
		log.Printf("%s: Nats завершил свою работу!", nh.name)
	}
}
