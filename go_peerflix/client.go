package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/iplist"
	"github.com/dustin/go-humanize"
)

// Специальная команда очистки экрана
const clearScreen = "\033[H\033[2J"

// TODO: ???
const torrentBlockListURL = "http://john.bitsurge.net/public/biglist.p2p.gz"

// Регулярка проверки HTTPS
var isHTTP = regexp.MustCompile(`^https?:\/\/`)

////////////////////////////////////////////////////////////////////////////////////////////////////

func downloadFile(URL string) (fileName string, err error) {
	var file *os.File
	if file, err = ioutil.TempFile(os.TempDir(), "go-peerflix"); err != nil {
		return
	}

	defer func() {
		if ferr := file.Close(); ferr != nil {
			log.Printf("Error closing torrent file: %s", ferr)
		}
	}()

	// #nosec
	// We are downloading the url the user passed to us, we trust it is a torrent file.
	response, err := http.Get(URL)
	if err != nil {
		return
	}

	defer func() {
		if ferr := response.Body.Close(); ferr != nil {
			log.Printf("Error closing torrent file: %s", ferr)
		}
	}()

	_, err = io.Copy(file, response.Body)

	return file.Name(), err
}

func downloadBlockList(blocklistPath string) (err error) {
	log.Printf("Downloading blocklist")
	fileName, err := downloadFile(torrentBlockListURL)
	if err != nil {
		log.Printf("Error downloading blocklist: %s\n", err)
		return
	}

	return os.Rename(fileName, blocklistPath)
}

// Download and add the blocklist.
func getBlocklist() iplist.Ranger {
	var err error
	blocklistPath := os.TempDir() + "/go-peerflix-blocklist.gz"

	if _, err = os.Stat(blocklistPath); os.IsNotExist(err) {
		err = downloadBlockList(blocklistPath)
	}

	if err != nil {
		log.Printf("Error downloading blocklist: %s", err)
		return nil
	}

	// Load blocklist.
	// #nosec
	// We trust our temporary directory as we just wrote the file there ourselves.
	blocklistReader, err := os.Open(blocklistPath)
	if err != nil {
		log.Printf("Error opening blocklist: %s", err)
		return nil
	}

	// Extract file.
	gzipReader, err := gzip.NewReader(blocklistReader)
	if err != nil {
		log.Printf("Error extracting blocklist: %s", err)
		return nil
	}

	// Read as iplist.
	blocklist, err := iplist.NewFromReader(gzipReader)
	if err != nil {
		log.Printf("Error reading blocklist: %s", err)
		return nil
	}

	log.Printf("Loading blocklist.\nFound %d ranges\n", blocklist.NumRanges())
	return blocklist
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// ClientError formats errors coming from the client.
type ClientError struct {
	Type   string
	Origin error
}

func (clientError ClientError) Error() string {
	return fmt.Sprintf("Error %s: %s\n", clientError.Type, clientError.Origin)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// ClientConfig specifies the behaviour of a client.
type ClientConfig struct {
	TorrentPath    string
	Port           int
	TorrentPort    int
	WithUploading  bool
	TCP            bool
	MaxConnections int
}

// NewClientConfig creates a new default configuration.
func NewClientConfig() ClientConfig {
	return ClientConfig{
		Port:           8080,
		TorrentPort:    50007,
		WithUploading:  false,
		TCP:            true,
		MaxConnections: 25,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Client manages the torrent downloading.
type Client struct {
	TorrentClient *torrent.Client
	Torrent       *torrent.Torrent
	Progress      int64
	Uploaded      int64
	Config        ClientConfig
}

// NewClient creates a new torrent client based on a magnet or a torrent file.
// If the torrent file is on http, we try downloading it.
func NewClient(cfg ClientConfig) (client Client, err error) {
	// Сохряняем конфигурацию клиента
	client.Config = cfg

	// TODO: Получаем список адресов
	blocklist := getBlocklist()

	// Создаем конфигурацию для торрент-клиента
	torrentConfig := torrent.NewDefaultClientConfig()
	torrentConfig.DataDir = os.TempDir()
	torrentConfig.NoUpload = !cfg.WithUploading
	torrentConfig.DisableTCP = !cfg.TCP
	torrentConfig.ListenPort = cfg.TorrentPort
	torrentConfig.IPBlocklist = blocklist
	//torrentConfig.EstablishedConnsPerTorrent = 1
	//torrentConfig.DownloadRateLimiter = rate.NewLimiter(50, 50)
	//torrentConfig.UploadRateLimiter = rate.NewLimiter(50, 50)

	// Создаем торрент-клиент
	var c *torrent.Client
	c, err = torrent.NewClient(torrentConfig)
	if err != nil {
		return client, ClientError{Type: "creating torrent client", Origin: err}
	}

	// Сохраняем его
	client.TorrentClient = c

	// Работа с торрентом

	// Обрабатываем ссылку для создания торрента
	var t *torrent.Torrent
	if strings.HasPrefix(cfg.TorrentPath, "magnet:") {
		// Добавляем Magnet ссылку
		if t, err = client.TorrentClient.AddMagnet(cfg.TorrentPath); err != nil {
			return client, ClientError{Type: "adding torrent", Origin: err}
		}
	} else {
		// Иначе добавляем торрент файлик

		// Если это ссылка на фалик - начинаем его качать
		if isHTTP.MatchString(cfg.TorrentPath) {
			if cfg.TorrentPath, err = downloadFile(cfg.TorrentPath); err != nil {
				return client, ClientError{Type: "downloading torrent file", Origin: err}
			}
		}

		// Добавляем торрент файл на обработку клиенту
		if t, err = client.TorrentClient.AddTorrentFromFile(cfg.TorrentPath); err != nil {
			return client, ClientError{Type: "adding torrent to the client", Origin: err}
		}
	}

	// Сохряняем торрент объект и устанавливаем нужное количество подключений
	client.Torrent = t
	client.Torrent.SetMaxEstablishedConns(cfg.MaxConnections)
	//client.Torrent.SetMaxEstablishedConns(20)

	// Запускаем корутину, которая отслеживает поступление новых данных
	go func() {
		// TODO: ???
		// Ждем получения новой информации
		<-client.Torrent.GotInfo()
		// Выкачиваем эту информацию
		client.Torrent.DownloadAll()

		//client.TorrentClient.WaitAll()

		// TODO: ???
		// Prioritize first 5% of the file.
		largestFile := client.getLargestFile()
		firstPieceIndex := largestFile.Offset() * int64(t.NumPieces()) / t.Length()
		endPieceIndex := (largestFile.Offset() + largestFile.Length()) * int64(t.NumPieces()) / t.Length()
		for idx := firstPieceIndex; idx <= endPieceIndex*5/100; idx++ {
			t.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
		}
	}()

	return
}

// Close закрывает работу с торрентом
func (c *Client) Close() {
	c.Torrent.Drop()
	c.TorrentClient.Close()
}

// RenderInfoToCLI выводит в коммандную строку информацию о прогрессе и тд
func (c *Client) RenderInfoToCLI() {
	t := c.Torrent

	if t.Info() == nil {
		return
	}

	currentProgress := t.BytesCompleted()
	downloadSpeed := humanize.Bytes(uint64(currentProgress-c.Progress)) + "/s"
	c.Progress = currentProgress

	complete := humanize.Bytes(uint64(currentProgress))
	size := humanize.Bytes(uint64(t.Info().TotalLength()))

	bytesWrittenData := t.Stats().BytesWrittenData
	uploadProgress := (&bytesWrittenData).Int64() - c.Uploaded
	uploadSpeed := humanize.Bytes(uint64(uploadProgress)) + "/s"
	c.Uploaded = uploadProgress
	percentage := c.percentage()
	totalLength := t.Info().TotalLength()

	output := bufio.NewWriter(os.Stdout)

	fmt.Fprint(output, clearScreen)
	fmt.Fprint(output, t.Info().Name+"\n")
	fmt.Fprint(output, strings.Repeat("=", len(t.Info().Name))+"\n")
	if c.ReadyForPlayback() {
		fmt.Fprintf(output, "Stream: \thttp://localhost:%d\n", c.Config.Port)
	}
	if currentProgress > 0 {
		fmt.Fprintf(output, "Progress: \t%s / %s  %.2f%%\n", complete, size, percentage)
	}
	if currentProgress < totalLength {
		fmt.Fprintf(output, "Download speed: %s\n", downloadSpeed)
	}
	if c.Config.WithUploading {
		fmt.Fprintf(output, "Upload speed: \t%s", uploadSpeed)
	}

	output.Flush()
}

/*
func (c Client) RenderPieces() (output string) {
	pieces := c.Torrent.PieceStateRuns()
	for i := range pieces {
		piece := pieces[i]

		if piece.Priority == torrent.PiecePriorityReadahead {
			output += "!"
		}

		if piece.Partial {
			output += "P"
		} else if piece.Checking {
			output += "c"
		} else if piece.Complete {
			output += "d"
		} else {
			output += "_"
		}
	}

	return
}
*/

// ReadyForPlayback checks if the torrent is ready for playback or not.
// We wait until 5% of the torrent to start playing.
func (c Client) ReadyForPlayback() bool {
	return c.percentage() > 5
}

// GetFile это http обработчик, обрабатывающий HTTP запрос на отдачу файла видео
func (c Client) GetFile(w http.ResponseWriter, r *http.Request) {
	target := c.getLargestFile()
	entry, err := NewFileReader(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := entry.Close(); err != nil {
			log.Printf("Error closing file reader: %s\n", err)
		}
	}()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+c.Torrent.Info().Name+"\"")
	http.ServeContent(w, r, target.DisplayPath(), time.Now(), entry)
}

func (c Client) getLargestFile() *torrent.File {
	var target *torrent.File
	var maxSize int64

	for _, file := range c.Torrent.Files() {
		if maxSize < file.Length() {
			maxSize = file.Length()
			target = file
		}
	}

	return target
}

func (c Client) percentage() float64 {
	info := c.Torrent.Info()

	if info == nil {
		return 0
	}

	return float64(c.Torrent.BytesCompleted()) / float64(info.TotalLength()) * 100
}
