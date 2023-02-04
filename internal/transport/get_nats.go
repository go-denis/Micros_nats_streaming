package transport

import (
	"encoding/json"
	"log"
	"micros/internal/database"
	"micros/internal/models"
	"os"
	"strconv"
	"time"

	stan "github.com/nats-io/stan.go"
)

type Getter struct {
	getter stan.Subscription //Подписка
	dbO    *database.DB      //База данных
	sc     *stan.Conn        //Соединеие натс
	name   string            //Название
}

func NewSubscriber(db *database.DB, conn *stan.Conn) *Getter {
	return &Getter{
		name: "Gettet", //Название
		dbO:  db,       //База
		sc:   conn,     //Ссылка
	}
}

// Создаем подписку на тему
func (g Getter) Getter() {
	// Ошибки
	var err error
	//Сонвертируем в int время ожидания
	ackWait, err := strconv.Atoi(os.Getenv("NATS_WAIT_SECONDS"))
	if err != nil {
		log.Printf("%s: Сообщение получено!\n", g.name)
		return
	}
	//Подписчмк
	g.getter, err = (*g.sc).Subscribe(os.Getenv("NATS_GETTER"),
		//Метод подтверждения с проверкой на сохранение
		func(m *stan.Msg) {
			log.Printf("%s: Сообщение получено!\n", g.name)
			//var mdata := &m.Data{}
			if g.messageHandler(m.Data) {
				err := m.Ack() // в случае успешного сохранения msg уведомляем NATS.
				if err != nil {
					log.Printf("%s ack() err: %s", g.name, err)
				}
			}
		},
		stan.AckWait(time.Duration(ackWait)*time.Second), // Интервал тайм-аута - AckWait (30 сек default) - ожидание уведомления NATS о чтении сообщения
		//stan.DeliverAllAvailable(),                       // DeliverAllAvailable доставит все доступные сообщения
		stan.DurableName(os.Getenv("NATS_DURABLE_NAME")), // долговечные подписки позволяют клиентам назначить постоянное имя подписке

		stan.SetManualAckMode(), // ручной режим подтверждения приема сообщения для подписки
		stan.MaxInflight(5))     // указывает максимальное количество ожидающих подтверждения (сообщений, которые были доставлены, но не подтверждены),

	if err != nil {
		log.Printf("%s: Ошибка: %v\n", g.name, err)
	}
	log.Printf("%s: Подписался на  %s\n", g.name, os.Getenv("NATS_GETTER"))
}

// Работа с сообщением, записть в базу заказа
func (g *Getter) messageHandler(data []byte) bool {
	order := models.Orders{} //Полученный заказ
	err := json.Unmarshal(data, &order)
	if err != nil {
		log.Printf("%s: messageHandler() error, %v\n", g.name, err)
		// ошибка формата присланных данных. Пропускаем, сообщив серверу, что сообщение получили
		return true
	}
	log.Printf("%s: Распарил json: %v\n", g.name, order)

	_, err = g.
		dbO.
		InsertOrder(order) //Записываем заказ в базу
	//Проверка на корректность записи в базу
	if err != nil {
		log.Printf("%s: ошибка, повторите попытку, ну удается добавить заказ: %v\n", g.name, err)
		return false
	}
	return true
}

// Закрытие получения
func (s *Getter) UnGetter() {
	if s.getter != nil {
		s.getter.Unsubscribe()
	}
}
