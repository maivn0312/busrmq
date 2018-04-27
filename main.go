package busrmq

import (
	"fmt"

	"github.com/streadway/amqp"
	"github.com/tkanos/gonfig"
)

type Config struct {
	USER string
	PASS string
	HOST string
	PORT string
}

func (c *Config) StringConnection() string {
	return "amqp://" + c.USER + ":" + c.PASS + "@" + c.HOST + ":" + c.PORT
}

type BusRabbitMQ struct {
	Config Config
}

func Init(config string) BusRabbitMQ {
	Config := Config{}
	err := gonfig.GetConf(config, &Config)
	if err != nil {
		panic(err)
	}
	Bus := BusRabbitMQ{Config: Config}
	return Bus
}
func (bus *BusRabbitMQ) Consumer(exchangeName string, routingKeys []string, callback func([] byte)) {

	fmt.Println(bus)
	conn, err := amqp.Dial(bus.Config.StringConnection()) //создаем подключение
	failOnError(err, "Failed connection to RabbitMQ")
	defer conn.Close()

	fmt.Println("tyty2")
	ch, err := conn.Channel() //создаем канал
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	fmt.Println("tyty")
	ch.ExchangeDeclare( //создаем точку доступа
		exchangeName,
		"direct",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to open a exchange")

	queue, err := ch.QueueDeclare( //создаем очередь
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for _, routingKey := range routingKeys { //биндим точку доступа с очередью с route key
		ch.QueueBind(queue.Name, routingKey, exchangeName, false, nil)
	}

	msgs, err := ch.Consume( // подписываемся на сообщения из очереди
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failOnError(err, "Failed to register a consumer")

	fmt.Println("Queue " + queue.Name + " run")
	for d := range msgs { // обрабатываем сообщения
		fmt.Printf("Received a message: %s", d.Body)
		callback(d.Body)
	}
	fmt.Println("Queue " + queue.Name + " end")
}

func (bus *BusRabbitMQ) ConsumerAck(exchangeName string, routingKeys []string, callback func([] byte) bool) {
	conn, err := amqp.Dial(bus.Config.StringConnection()) //создаем подключение
	failOnError(err, "Failed connection to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel() //создаем канал
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	ch.ExchangeDeclare( //создаем точку доступа
		exchangeName,
		"direct",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to open a exchange")

	queue, err := ch.QueueDeclare( //создаем очередь
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for _, routingKey := range routingKeys { //биндим точку доступа с очередью с route key
		ch.QueueBind(queue.Name, routingKey, exchangeName, false, nil)
	}

	msgs, err := ch.Consume( // подписываемся на сообщения из очереди
		queue.Name, // queue
		"",         // consumer
		false,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failOnError(err, "Failed to register a consumer")

	fmt.Println("Queue " + queue.Name + " run")
	for d := range msgs { // обрабатываем сообщения
		fmt.Printf("Received a message: %s", d.Body)
		if callback(d.Body) {
			ch.Ack(d.DeliveryTag, false)
		}
	}
	fmt.Println("Queue " + queue.Name + " end")
}

func (bus *BusRabbitMQ) Producer(exchangeName string, routingKey string, msg string) {
	conn, err := amqp.Dial(bus.Config.StringConnection())
	failOnError(err, "Failed connection to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName,
		"direct",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to open a exchange")

	err = ch.Publish(
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	failOnError(err, "Failed to publish a message")
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
		panic(err)
	}
}
