package server

// Конфигурация маршрутов ошибки не возврашает, потому что просто описывает какой-то маршрут
func (s *APIServer) configRouter() {
	//Новый роутер создался в New() APiServer, поэтому не обьявляем
	//s.router.NewRoute()
	//Главная страница приветсвия
	s.router.HandleFunc("/", s.IndexHandler).Methods("GET")

	sub := s.router.Methods("GET").Subrouter()
	sub.HandleFunc("/orders/{orderID}", s.HandleOrder)
	sub.Use(s.orderCont)
	//s.router.HandleFunc("/{id: [0-9] +}", s.HandleOrder).Methods("GET")
	//s.router.HandleFunc("/{orderID}", s.HandleOrder) //.Subrouter().Use(s.orderCont)
	//s.orderRoute()
	//s.router.NewRoute()
	//страница заказов с выгрузкой
	//s.router.HandleFunc("/orders/{orderID}", s.HandleOrder).Subrouter().Use(s.orderCont)
	//s.router.Use(s.orderCont)
	//s.router.HandleFunc("/{orderID}", s.HandleOrder).Subrouter().Use() //.Methods("GET", "POST")
}
