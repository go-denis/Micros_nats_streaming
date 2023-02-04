package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_handleOrder(t *testing.T) {
	s := New(NewServConfig())                                 //Наш сервер
	rec := httptest.NewRecorder()                             //Рекордер из пакета httptest
	req, _ := http.NewRequest(http.MethodGet, "/orders", nil) //Объект реквест
	s.HandleOrder(rec, req)                                   //.ServeHTTP(rec, req)
	assert.Equal(t, rec.Body.String(), "Hello")               //Проверяем что тело равно hello
}
