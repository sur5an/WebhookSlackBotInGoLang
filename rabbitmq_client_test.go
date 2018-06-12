package main

import "testing"

func TestRabbitMQClient_Init(t *testing.T) {
	rmq := make(map[string]string)
	host := "localhost"
	username := "username"
	password := "password"
	port := "port"
	rmq["host"] = host
	rmq["username"] = username
	rmq["password"] = password
	rmq["port"] = port
	rmqClient := rabbitMQClient{}
	rmqClient.Init(rmq)
	if rmqClient.rabbitMQHost != host || rmqClient.password != password ||
		rmqClient.username != username || rmqClient.port != port {
		t.Errorf("Unit test failed for Init - missing RabbitMQClinet configs")
	}
}
