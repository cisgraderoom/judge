package utils

import (
	"fmt"

	"github.com/streadway/amqp"
	"gitlab.com/cisclassroom/compiler/config"
	"gitlab.com/cisclassroom/compiler/logs"
)

func GetChannelCB(service, module string, conn *amqp.Connection) {

	// Declare Channel
	ch, err := conn.Channel()

	if err != nil {
		logs.Error("create channel")
		return
	}

	defer ch.Close()

	exchange := config.QueueConfig[service].(map[string]interface{})["exchange"]
	queue := config.QueueConfig[service].(map[string]interface{})["queues"].(map[string]interface{})[module]

	err = ch.ExchangeDeclare(
		exchange.(map[string]interface{})["name"].(string),                                      // name
		exchange.(map[string]interface{})["type"].(string),                                      // type
		exchange.(map[string]interface{})["options"].(map[string]interface{})["durable"].(bool), // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	)

	if err != nil {
		logs.Error("ExchangeDeclare error")
		return
	}

	q, err := ch.QueueDeclare(
		queue.(map[string]interface{})["name"].(string),                                      // name
		queue.(map[string]interface{})["options"].(map[string]interface{})["durable"].(bool), // durable
		false, // delete when used
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		logs.Error("QueueDeclare error")
		return
	}

	logs.Info(fmt.Sprintf("Waiting - queue : %s", q.Name))

	// bind queue with all routing keys
	for _, s := range queue.(map[string]interface{})["routing_keys"].([]string) {
		err = ch.QueueBind(
			q.Name, // queue name
			s,      // routing key
			exchange.(map[string]interface{})["name"].(string), // exchange
			false, // noWait
			nil,   // args
		)
		if err != nil {
			logs.Error("Failed to bind a queue")
			return
		}
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)

	if err != nil {
		logs.Error("Failed to bind a queue")
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Println(string(d.Body[:]))
			// curlUpdateDailyReport(d)
		}
	}()
	<-forever
}
