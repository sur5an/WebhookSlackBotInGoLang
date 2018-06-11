package main

import (
	"flag"
	"github.com/spf13/viper"
	"log"
)

const (
	SlackToken       = "slackToken"
	RabbitMQDetails  = "rabbitMQ"
	RabbitMQUserName = "username"
	RabbitMQPassword = "password"
	RabbitMQHost     = "host"
	RabbitMQPort     = "port"
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
	failOnError(err, "unable to read the bot config")
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

	var client slackClient
	client.connect(v.GetString(SlackToken))

	getMessage := make(chan SlackMessage)
	sendAck := make(chan bool)

	var mqClient rabbitMQClient
	mqClient.Init(v.GetStringMapString(RabbitMQDetails))
	mqClient.connect()

	defer mqClient.ch.Close()
	defer mqClient.connection.Close()
	defer close(sendAck)

	go mqClient.listen("publish_to_unknown", getMessage, sendAck)

	for {
		messageToSlack := <-getMessage
		log.Printf("Got a message to %s (%f)", messageToSlack.Consumer, messageToSlack.Time)
		err := client.sendMessage(messageToSlack.MessageToSend, messageToSlack.Consumer)
		ack := false
		if err != nil {
			ack = true
		}
		sendAck <- ack
	}
}
