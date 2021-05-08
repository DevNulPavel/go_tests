package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func setup_logs() {
	// Log as JSON instead of the default ASCII formatter.
	// logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: false,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.TraceLevel)
}

func main() {
	setup_logs()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		logrus.Fatalf("AMQP connection failed: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		logrus.Fatalf("Channel create failed: %s", err)
	}

	q, err := ch.QueueDeclare("my_queue", true, true, false, false, nil)
	if err != nil {
		logrus.Fatalf("Queue create failed: %s", err)
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("Test text"),
	})
	if err != nil {
		logrus.Fatalf("Data publish failed: %s", err)
	}
}
