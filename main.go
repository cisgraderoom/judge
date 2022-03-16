package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/streadway/amqp"
	"gitlab.com/cisclassroom/compiler/comms"
	"gitlab.com/cisclassroom/compiler/config"
	"gitlab.com/cisclassroom/compiler/logs"
	"gitlab.com/cisclassroom/compiler/schemas"
)

func main() {
	// Start Report Service
	logs.Info(fmt.Sprintf("🐳 GO_ENV: %s, APP_ENV: %s", os.Getenv("GO_ENV"), os.Getenv("APP_ENV")))

	forever := make(chan bool)
	refers := []string{"result"}

	conn := comms.NewConnection(
		config.RabbitMQ_URL(),
		config.SERVICE,
		refers,
		config.QueueConfig[config.SERVICE].(map[string]interface{})["exchange"].(map[string]interface{}),
		config.QueueConfig[config.SERVICE].(map[string]interface{})["queues"].(map[string]interface{}),
		config.QueueConfig[config.SERVICE].(map[string]interface{})["consumer"].(map[string]interface{}),
	)
	if err := conn.Connect(); err != nil {
		panic(err)
	}
	if err := conn.BindQueue(); err != nil {
		panic(err)
	}
	deliveries, err := conn.Consume()
	if err != nil {
		panic(err)
	}
	for q, d := range deliveries {
		go conn.HandleConsumedDeliveries(q, d, messageHandler)
	}
	<-forever
}

func messageHandler(_ comms.Connection, q string, deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		m := comms.Message{
			Queue:         q,
			Body:          comms.MessageBody{Data: d.Body, Type: ""},
			ContentType:   d.ContentType,
			Priority:      d.Priority,
			CorrelationID: d.CorrelationId,
		}
		if err := judge(m.Body); err != nil {
			logs.Error(err)
		} else {
			d.Ack(false)
		}
	}
}

func judge(message comms.MessageBody) error {
	var payload schemas.Payload
	if err := json.Unmarshal(message.Data, &payload); err != nil {
		logs.Error(err)
		return err
	}
	err := writeTestCaseFile(payload)
	if err != nil {
		logs.Error(err)
		return err
	}
	err = compile(payload)
	if err != nil {
		logs.Error(err)
		return err
	}
	return nil
}

func writeTestCaseFile(payload schemas.Payload) error {
	dir, _ := os.Getwd()
	for i := 1; i <= payload.Testcase; i++ {
		index := i - 1
		input := []byte(payload.Input[index])
		if err := os.WriteFile(fmt.Sprintf("%s/out/%d.in", dir, i), input, 0775); err != nil {
			return err
		}
		output := []byte(payload.Output[index])
		if err := os.WriteFile(fmt.Sprintf("%s/out/%d.out", dir, i), output, 0775); err != nil {
			return err
		}
	}
	code := []byte(payload.Code)
	if err := os.WriteFile(fmt.Sprintf("%s/out/Main.%s", dir, payload.Language), code, 0775); err != nil {
		return err
	}
	return nil
}

func compile(payload schemas.Payload) error {
	dir, _ := os.Getwd()
	script := fmt.Sprintf("%s/script/%s", dir, "compile.sh")
	err := os.Chmod(script, 0775)
	if err != nil {
		return err
	}
	fileCompile := fmt.Sprintf("%s/out/Main.%s", dir, payload.Language)
	fileOutput := fmt.Sprintf("%s/out/output", dir)
	errOutput := fmt.Sprintf("%s/out/err", dir)
	res := exec.Command("/bin/sh", script, fileCompile, payload.Language, fileOutput, errOutput)
	if res.Run() != nil {
		return res.Run()
	}
	return nil
}
