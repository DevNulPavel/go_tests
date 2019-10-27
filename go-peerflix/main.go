package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Exit statuses.
const (
	_ = iota
	exitNoTorrentProvided
	exitErrorInClient
)

func main() {
	// Парсим имя плеера для воспроизведения
	player := flag.String("player", "", "Open the stream with a video player ("+joinPlayerNames()+")")

	// Создаем конфигурацию для плеера
	cfg := NewClientConfig()
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Port to stream the video on")
	flag.IntVar(&cfg.TorrentPort, "torrent-port", cfg.TorrentPort, "Port to listen for incoming torrent connections")
	flag.BoolVar(&cfg.Seed, "seed", cfg.Seed, "Seed after finished downloading")
	flag.IntVar(&cfg.MaxConnections, "conn", cfg.MaxConnections, "Maximum number of connections")
	flag.BoolVar(&cfg.TCP, "tcp", cfg.TCP, "Allow connections via TCP")
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(exitNoTorrentProvided)
	}
	cfg.TorrentPath = flag.Arg(0)

	// Создаем новый торрент-клиент
	client, err := NewClient(cfg)
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(exitErrorInClient)
	}

	// Запускаем HTTP сервер для отдачи данных плееру
	go func() {
		http.HandleFunc("/", client.GetFile)
		httpErr := http.ListenAndServe(":"+strconv.Itoa(cfg.Port), nil)
		log.Fatal(httpErr)
	}()

	// Запускаем нужный нам плеер
	if *player != "" {
		go func() {
			for !client.ReadyForPlayback() {
				time.Sleep(time.Second)
			}
			openPlayer(*player, cfg.Port)
		}()
	}

	// Обработка закрытия приложения
	// Создаем канал, куда будут прилетать сигналы
	interruptChannel := make(chan os.Signal, 1)
	// Описываем нужные нам сигналы
	signal.Notify(interruptChannel,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	// Запускаем корутину, которая будет обрабатывать наши сигналы
	go func(interruptChannel chan os.Signal) {
		for range interruptChannel {
			log.Println("Exiting...")
			client.Close()
			os.Exit(0)
		}
	}(interruptChannel)

	// Главный цикл работы приложения
	for {
		client.RenderInfoToCLI()
		time.Sleep(time.Second)
	}
}
