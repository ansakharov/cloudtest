## Сервер-клиент для записи логов

#### Как использовать

Запустить сервер: 

`go run cmd/server/main.go`

После запустить клиента для записи в файл новых данных: 

`go run cmd/client/main.go -l=100000`

#### Опции

`-l` - количество сообщений, которые сгенерирует клиент для отправки серверу. Флаг обязателен.

`-rand` - при наличии флага генерируются сообщения от случайных байт, занимает много времени (секунды). 
Без флага генерирует последовательность байта 'a'.

#### Ограничения
Нет обработки сетевых ошибок, ошибок файловой системы.