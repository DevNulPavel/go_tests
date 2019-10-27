package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
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
	// Профилирование кода
	// https: //golang.org/pkg/runtime/pprof/
	// https://blog.golang.org/profiling-go-programs
	cpuprofile := flag.String("cpuprofile", "", "Write cpu profile to `file`")
	memprofile := flag.String("memprofile", "", "Write memory profile to `file`")

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

	// Профилирование кода
	// https://golang.org/pkg/runtime/pprof/
	// https://blog.golang.org/profiling-go-programs
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

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
	mainLoopExitChannel := make(chan bool)
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
			mainLoopExitChannel <- true
		}
	}(interruptChannel)

	// Главный цикл работы приложения
	stopLoop := false
	for !stopLoop {
		client.RenderInfoToCLI()
		time.Sleep(time.Second)

		select {
		case stopLoop = <-mainLoopExitChannel:
		default:
		}
	}

	log.Println("Exit success")
}
