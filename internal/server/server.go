package server

import (
	"context"
	"fmt"
	"log"
	"micros/internal/database"
	"micros/internal/transport"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// АПИ Сервер
type APIServer struct {
	configserv         *ConfigServ            //Указатель на структуру ConfigServ
	logger             *logrus.Logger         //Логи
	router             *mux.Router            //Маршрутизатор
	database           *database.DB           //Указатель на структуру DB
	cache              *database.Cache        //Указатель на кеш в структуре базы данных
	natsStreaming      *transport.NatsHandler //Указатель nats
	httpServerExitDone *sync.WaitGroup        //Для грамотного завершения работы сервера
}

// пльзовательский тип для приведения к общему интерфейсу
type key string

// Ключ для обработки хендлера
const orderName key = "order"

func New(configserv *ConfigServ) *APIServer {
	s := &APIServer{
		configserv: configserv,      //Конфигурация
		logger:     logrus.New(),    //Логи
		router:     mux.NewRouter(), //Новый маршрут
	}
	s.configRouter()    //Запуск маршрутов
	s.configureLogger() //Запуск логов
	s.ConfigDatabase()  //Запуск базы
	s.ConfigCache()     //Загрузка кеша
	s.ConfigNats()      //Запуск натс стриминг
	return s
}

// Метод запуска сервера, где также происходит соединение с бд
func (s *APIServer) Start() {

	s.httpServerExitDone = &sync.WaitGroup{}
	s.httpServerExitDone.Add(1)

	//Теперь, сначала проверяем не произошло ли какой-то беды
	if err := s.configureLogger(); err != nil {
		log.Printf("Ошибка конфигуранции логов %v", err)
	}
	/*
		//Запуск конфигурации БД
		if err := s.ConfigDatabase(); err != nil {
			log.Printf("Ошибка конфигурации БД %v", err)
		}
		//Инициализация кеша
		if err := s.ConfigCache(); err != nil {
			log.Printf("Ошибка конфигурации кеша %v", err)
		}

		if err := s.ConfigNats(); err != nil {
			log.Printf("Ошибка настс %v", err)
		}
	*/
	go func() {
		defer s.httpServerExitDone.Done() // let main know we are done cleaning up

		if err := http.ListenAndServe(s.configserv.BindAddr, s.router); err != http.ErrServerClosed {
			s.logger.Info("Сервер стартует, все прошло отлично!") //Если все ок, оповещаем, что все ОК
			//return
		}

	}()

}

// Для конфигурации логгера
func (s *APIServer) configureLogger() error {
	//Парсим нашу строку, которая хранит конфиг лог левел
	level, err := logrus.ParseLevel(s.configserv.LogLevel)
	if err != nil {
		return err
	}
	//Если все ок, ставим логгеру соответствующий уровень
	s.logger.SetLevel(level)
	return nil
}

// Собираем конфигурацию БД
func (s *APIServer) ConfigDatabase() error {

	//Новое подключение
	st := database.NewBD(s.configserv.DataBase)
	if err := st.Open(); err != nil {
		return err
	}
	//Записываем в переменную database
	s.database = st
	return nil
}

// Собираем конфигурацию Кеша
func (s *APIServer) ConfigCache() error {

	//Новое подключение
	ch := database.NewCache(s.database)

	//Записываем в переменную cache
	s.cache = ch
	return nil
}

// Собираем конфигурацию natsStreaming
func (s *APIServer) ConfigNats() error {

	//Новое подключение
	sn := transport.NewNatsHandler(s.database)
	if err := sn.Connect(); err != nil {
		return err
	}
	//Записываем в переменную natsStreaming
	s.natsStreaming = sn
	return nil
}

// Что-то вроде мидлвар, здесь мы сохраняем Заказ в контекст
func (s *APIServer) orderCont(next http.Handler) http.Handler {
	//s.Exit()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Записываем переменные маршрута для текущего запроса, если таковые имеются.
		order_id_string := mux.Vars(r)["orderID"] // SetURLVars(r, "orderID")
		//Пасим строку с айди и заносим в массив
		orderID, err := strconv.ParseInt(order_id_string, 10, 64)
		if err != nil {
			log.Printf("Ошибка конвертации %s в число: %v\n", order_id_string, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		log.Printf("Запрос Заказа из кеша/бд, Заказ id: %v\n", order_id_string)
		orderOut, err := s.cache.GetOrderCache(orderID)
		if err != nil {
			log.Printf("Ошибка получения Заказ из базы данных: %v\n", err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound) // 404
			return
		}
		ctx := context.WithValue(r.Context(), orderName, orderOut)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *APIServer) Exit() {
	// По нажатию CTRL-C
	// Запуск очистки
	signalChan := make(chan os.Signal, 1)
	idleConnsClosed := make(chan bool) //Флаг хорошего завершения работы
	signal.Notify(signalChan, os.Interrupt)
	//Отслеживаем действие горутиной после запуска сервера
	go func() {

		for range signalChan {
			fmt.Printf("\nПолучено прерывание, отмена подписки и закрытие соединения...\n\n")
			s.cache.Finish()         //Очистка кеша
			s.natsStreaming.Finish() //Завершение работы NATS-Streaming
			s.ShutdownServer()       //Корректное завершение работы сервера
			idleConnsClosed <- true
		}
	}()
	<-idleConnsClosed
}

// Корректное завершение работы
func (s *APIServer) ShutdownServer() {
	log.Printf("завершение работы сервера...")
	s.httpServerExitDone.Wait()
	log.Println("Сервер завершил свою работу!")
}
