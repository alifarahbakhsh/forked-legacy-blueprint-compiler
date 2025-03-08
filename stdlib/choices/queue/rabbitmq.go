package queue

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/stdlib/components"
)

type RabbitMQ struct {
	name  string
	queue amqp.Queue
	ch    *amqp.Channel
	conn  *amqp.Connection
}

func NewRabbitMQ(queue_name string, addr string, port string) *RabbitMQ {
	conn, err := amqp.Dial("amqp://guest:guest@" + addr + ":" + port + "/")
	if err != nil {
		log.Fatal(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	q, err := ch.QueueDeclare(queue_name, false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &RabbitMQ{name: queue_name, conn: conn, ch: ch, queue: q}
}

func (q *RabbitMQ) Send(ctx context.Context, msg []byte) error {
	publish_msg := amqp.Publishing{ContentType: "text/plain", Body: msg}
	return q.ch.Publish("", q.queue.Name, false, false, publish_msg)
}

func (q *RabbitMQ) Recv(fn components.Callback_fn) {
	msgs, err := q.ch.Consume(q.queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			fn(d.Body)
		}
	}()
	<-forever
}
