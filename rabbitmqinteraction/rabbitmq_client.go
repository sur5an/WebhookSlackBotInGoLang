package rabbitmqinteraction

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"github.com/sur5an/WebhookSlackBotInGoLang/utils"
)

const (
	RabbitMQUserName	= "username"
	RabbitMQPassword 	= "password"
	RabbitMQHost     	= "host"
	RabbitMQPort     	= "port"
	ForClient			 	= "ForClient"
	ChannelCloseEvent 	= "ChannelClosed"
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

type RabbitMQClient struct {
	rabbitMQHost string
	username     string
	password     string
	port         string
	Channel      *amqp.Channel
	Connection   *amqp.Connection
}

var AMQPConnect = amqp.Dial

func (client *RabbitMQClient) Init(rabbitMQDetails map[string]string) {
	client.rabbitMQHost = rabbitMQDetails[RabbitMQHost]
	client.username = rabbitMQDetails[RabbitMQUserName]
	client.password = rabbitMQDetails[RabbitMQPassword]
	client.port = rabbitMQDetails[RabbitMQPort]
	return
}

func (client *RabbitMQClient) Connect() {
	conn, err := AMQPConnect("amqp://" + client.username + ":" +
		client.password + "@" + client.rabbitMQHost + ":" + client.port)
	utils.FailOnError(err, "failed to open rabbitmq connection with "+client.rabbitMQHost)

	ch, err := conn.Channel()
	utils.FailOnError(err, "failed to open channel")
	client.Channel = ch
	client.Connection = conn
	return
}

func (client RabbitMQClient) Listen(queueName string, messageChannel chan SlackMessage, responseChannel chan bool) {

	q, err := client.Channel.QueueDeclare(queueName, false, false,
		false, false, nil)
	utils.FailOnError(err, "unable to declare queue")

	messages, err := client.Channel.Consume(q.Name, "goClient", false,
		false, false, false, nil)

	utils.FailOnError(err, "consume failed")

	for message := range messages {
		var messageToSend SlackMessage
		json.Unmarshal(message.Body, &messageToSend)
		log.Printf("Sending a message to %s (%f)", messageToSend.Consumer, messageToSend.Time)
		messageChannel <- messageToSend
		if <-responseChannel {
			err := message.Ack(false)
			utils.FailOnError(err, "unable to ack")
		} else {
			err := message.Nack(false, true)
			utils.FailOnError(err, "unable to ack")
		}
		log.Printf("Waiting for messages")
	}
	var messageToSend SlackMessage
	messageToSend.Consumer = ForClient
	messageToSend.MessageToSend = ChannelCloseEvent
	messageChannel <- messageToSend
}
