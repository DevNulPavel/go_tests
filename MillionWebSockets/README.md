# Usage
Этот репозиторий демонстрирует на сколько большое количество соединений может быть эффективно поддержано в линуксе

Каждая папка с примером показываем пример реализации сервера.

`setup.sh` - это враппер для запуска множества инстансов тестов с использованием Docker.
`destroy.sh` - останавливает тестовый враппер

Запуск отдельного инстанса может быть выполнен с помощью вызова `go run client_src/client.go -conn=<# connections to establish>`

Slides are available [here](https://speakerdeck.com/eranyanay/going-infinite-handling-1m-websockets-connections-in-go)
