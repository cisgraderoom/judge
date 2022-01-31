package main

import (
	"fmt"
	"os"

	"github.com/streadway/amqp"
	"gitlab.com/cisclassroom/compiler/config"
	"gitlab.com/cisclassroom/compiler/logs"
	"gitlab.com/cisclassroom/compiler/utils"
)

func main() {
	// Start Report Service
	logs.Info(fmt.Sprintf("🐳 GO_ENV: %s, APP_ENV: %s", os.Getenv("GO_ENV"), os.Getenv("APP_ENV")))

	// Connect RabbitMQ Server
	rabbitmqConfig := config.RabbitMQ_URL()
	conn, err := amqp.Dial(rabbitmqConfig)

	logs.Info(rabbitmqConfig)

	if err != nil {
		logs.Fatal("main connect rabbitmq")
	}

	defer conn.Close()

	// Conect RabbitMQ Success Log
	logs.Info("Successfully Connected to RabbitMQ")

	// get service queue config from service
	serviceConfig := config.QueueConfig[config.SERVICE].(map[string]interface{})["queues"].(map[string]interface{})

	for key, _ := range serviceConfig {
		go utils.GetChannelCB(config.SERVICE, key, conn)
	}

	select {}
}
