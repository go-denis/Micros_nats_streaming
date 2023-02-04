package server

import (
	"html/template"
	"log"
	"micros/internal/models"
	"net/http"
)

// Функция для обработки маршрута заказов вернет HandlerFunc это мы делаем, чтобы можно было определить какие-то переменные
// и код внутри хендлера выполнится один раз
func (s *APIServer) HandleOrder(w http.ResponseWriter, r *http.Request) {

	cont := r.Context()
	orderOut, flag := cont.Value(orderName).(*models.Orders_issuance)
	//var ord_string models.Orders_issuance := []intcont
	log.Printf("%v HandleOrder() 1лог: ошибка приведения интерфейса к типу *orderName\n", orderOut)
	if !flag {
		log.Printf("%v HandleOrder() 2лог: ошибка приведения интерфейса к типу *orderName\n", cont)
		//Пользовательская ошибка
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity) // 422
		return
	}

	// Установка ответа для браузера, что страница загрузилась
	t, err := template.ParseFiles("web/default.html")
	if err != nil {
		log.Printf("HandleOrder(): ошибка парсинга шаблона html: %s\n", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	//Статус 200
	w.WriteHeader(http.StatusOK)
	//Передаем модель ответа на форму html
	t.ExecuteTemplate(w, "default.html", orderOut)
	if err != nil {
		log.Printf("HandleOrder(): ошибка выполнения шаблона html: %s\n", err)
		return
	}
}

// Обработчик главной страницы
func (a *APIServer) IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Установка ответа для браузера, что страница загрузилась
	t, err := template.ParseFiles("web/default.html")
	if err != nil {
		log.Printf("IndexHandler(): ошибка парсинга шаблона html: %s\n", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = t.ExecuteTemplate(w, "default.html", nil)
	if err != nil {
		log.Printf(" IndexHandler(): ошибка выполнения шаблона html: %s\n", err)
		return
	}
}
