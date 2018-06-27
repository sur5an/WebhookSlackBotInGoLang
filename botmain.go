package main

import (
	"flag"
	"github.com/spf13/viper"
	"log"
	"WebhookSlackBotInGoLang/slackintegration"
	"WebhookSlackBotInGoLang/utils"
	"WebhookSlackBotInGoLang/rabbitmqinteraction"
)

const (
	SlackToken       = "slackToken"
	RabbitMQDetails  = "rabbitMQ"
)

var rabbitMQDefaults = map[string]string{
	"username": "guest",
	"password": "guest",
	"host":     "localhost",
	"port":     "5672",
}

func readConfig(configFile string, defaults map[string]interface{}) *viper.Viper {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	utils.FailOnError(err, "unable to read the bot config")
	return v
}

func main() {
	inputConfig := flag.String("config", "./config/slackbot.yaml", "configuration file")
	flag.Parse()
	v := readConfig(*inputConfig,
		map[string]interface{}{
			"slackToken": nil,
			"rabbitMQ":   rabbitMQDefaults,
		})

	var client slackintegration.SlackClient
	client.Connect(v.GetString(SlackToken))

	getMessage := make(chan rabbitmqinteraction.SlackMessage)
	sendAck := make(chan bool)

	var mqClient rabbitmqinteraction.RabbitMQClient
	mqClient.Init(v.GetStringMapString(RabbitMQDetails))
	mqClient.Connect()

	defer mqClient.Channel.Close()
	defer mqClient.Connection.Close()
	defer close(sendAck)

	go mqClient.Listen("publish_to_unknown", getMessage, sendAck)

	for {
		messageToSlack := <-getMessage
		log.Printf("Got a message to %s (%f)", messageToSlack.Consumer, messageToSlack.Time)
		err := client.SendMessage(messageToSlack.MessageToSend, messageToSlack.Consumer)
		ack := false
		if err != nil {
			ack = true
		}
		sendAck <- ack
	}
}
