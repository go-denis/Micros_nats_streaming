# Micros_nats_streaming
Микросервис

# Golang + Nats-straming + Postgress 

Логированеи ошибок осуществялется с помощью пакета https://github.com/sirupsen/logrus
Работа с томл файлами, парсинг и тд с помощью пакета https://github.com/BurntSushi/toml
Для работы с http был выбран https://github.com/gorilla/mux
Для тестов использовалась библиотека github.com/stretchr/testify

Сервис реализует:
- Подключение и подписка на канал в Nats-streaming
- Сохранение полученных данных в Postgres
- Хранение и выдача данных по id из кеша
- В случае падения сервиса Кеш восстанваливается из Postgres
- Сделан простейший интерфейс отображения полученных данных

Не получилось решить проблему с парсингом toml, по какой-то причине, не парсятся данные, не смог выяснить
Схема базы данныйх:
![image](https://user-images.githubusercontent.com/97671717/216770355-dad83f7d-bc2d-46b0-8013-d52daec76d41.png)

Работа программы терминал:

![image](https://user-images.githubusercontent.com/97671717/216770548-0b45d79c-e908-4912-84f4-1a7b14ef85a3.png)

Работа фронденд программы:

![image](https://user-images.githubusercontent.com/97671717/216770468-a4487f96-9c31-4312-ae42-435ec3998c96.png)

## Требования
Установите компилятор для Golang 
Настроить локально БД (Создать таблицы/связи)

Установить Nats или развернуть в докер двумя командами:
//Создание изображения натс
docker run -d --name=nats-main -p 4222:4222 -p 6222:6222 -p 8222:8222 nats
//Добавление изображения натс стриминг и объединение 
docker run -d --link nats-main nats-streaming -store file -dir datastore -ns nats://nats-main:4222

Запусить командой go run `.\cmd\app\main.go`

Настроить под себя в файле `internal\config\config.go`

### Завершение работы
Для завешения работы просто нажмите `Ctrl+C` в консоли
