package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/evio"
)

func main() {
	var port int
	var loops int
	var udp bool
	var trace bool
	var reuseport bool
	var stdlib bool

	// Считываем значения-параметры, дефолтные если нету
	flag.IntVar(&port, "port", 5000, "server port")
	flag.BoolVar(&udp, "udp", false, "listen on udp")
	flag.BoolVar(&reuseport, "reuseport", false, "reuseport (SO_REUSEPORT)")
	flag.BoolVar(&trace, "trace", false, "print packets to console")
	flag.IntVar(&loops, "loops", 0, "num loops")
	flag.BoolVar(&stdlib, "stdlib", false, "use stdlib")
	flag.Parse()

	// Функция, которая вызывается при старте сервера когда он уже может работать с входящими соединениями
	servingFunc := func(srv evio.Server) (action evio.Action) {
		log.Printf("Echo server started on port %d (loops: %d)", port, srv.NumLoops)
		if reuseport {
			log.Printf("reuseport mode")
		}
		if stdlib {
			log.Printf("stdlib mode")
		}
		return action
	}

	// Вызывается при создании нового подключения к серверу
	newConnection := func(c evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
		contextString := "Current connection context string"
		c.SetContext(&contextString)
		return out, opts, action
	}

	// Функция, которая обрабатывает наличие новых данных
	dataFunc := func(c evio.Conn, in []byte) ([]byte, evio.Action) {
		contextString := c.Context().(*string)
		log.Printf("Get data called with context: %s", *contextString)
		if trace {
			log.Printf("%s", strings.TrimSpace(string(in)))
		}
		return in, evio.None
	}

	// Функция вызывается когда соединение было закрыто
	closedFunc := func(c evio.Conn, err error) evio.Action {
		return evio.None
	}

	var events evio.Events
	events.NumLoops = loops // Устанавливаем количество потоков с циклами
	events.Serving = servingFunc
	events.Opened = newConnection
	events.Data = dataFunc
	events.Closed = closedFunc
	scheme := "tcp"
	if udp {
		scheme = "udp"
	}
	if stdlib {
		scheme += "-net"
	}
	log.Fatal(evio.Serve(events, fmt.Sprintf("%s://:%d?reuseport=%t", scheme, port, reuseport)))
}
