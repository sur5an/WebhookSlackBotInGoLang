package main

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
)

type SlackMessage struct {
	EventType     string  `json:"event_type"`
	EventName     string  `json:"event_name"`
	RawData       bool    `json:"raw_data"`
	Actor         string  `json:"actor"`
	Time          float64 `json:"time"`
	MessageToSend string  `json:"message"`
	Consumer      string  `json:"consumer"`
	Channel       string  `json:"channel"`
}

type rabbitMQClient struct {
	rabbitMQHost string
	username     string
	password     string
	port         string
	ch           *amqp.Channel
	connection   *amqp.Connection
}

func (client *rabbitMQClient) Init(rabbitMQDetails map[string]string) {
	client.rabbitMQHost = rabbitMQDetails[RabbitMQHost]
	client.username = rabbitMQDetails[RabbitMQUserName]
	client.password = rabbitMQDetails[RabbitMQPassword]
	client.port = rabbitMQDetails[RabbitMQPort]
	return
}

func (client *rabbitMQClient) connect() {
	conn, err := amqp.Dial("amqp://" + client.username + ":" +
		client.password + "@" + client.rabbitMQHost + ":" + client.port)
	failOnError(err, "failed to open rabbitmq connection with "+client.rabbitMQHost)

	ch, err := conn.Channel()
	failOnError(err, "failed to open channel")
	client.ch = ch
	client.connection = conn
	return
}

func (client rabbitMQClient) listen(queueName string, goChannel chan SlackMessage) {

	defer close(goChannel)

	q, err := client.ch.QueueDeclare(queueName, false, false,
		false, false, nil)
	failOnError(err, "unable to declare queue")

	messages, err := client.ch.Consume(q.Name, "goClient", false,
		false, false, false, nil)

	failOnError(err, "consume failed")

	for message := range messages {
		var messageToSend SlackMessage
		json.Unmarshal(message.Body, &messageToSend)
		log.Printf("Sending a message to %s (%f)", messageToSend.Consumer, messageToSend.Time)
		goChannel <- messageToSend
		err := message.Ack(false)
		failOnError(err, "unable to ack")
		log.Printf("Waiting for messages")
	}
}
