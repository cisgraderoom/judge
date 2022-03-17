package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/streadway/amqp"
	"gitlab.com/cisclassroom/compiler/comms"
	"gitlab.com/cisclassroom/compiler/config"
	"gitlab.com/cisclassroom/compiler/conn"
	"gitlab.com/cisclassroom/compiler/logs"
	"gitlab.com/cisclassroom/compiler/schemas"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// Start Report Service
	logs.Info(fmt.Sprintf("🐳 GO_ENV: %s, APP_ENV: %s", os.Getenv("GO_ENV"), os.Getenv("APP_ENV")))

	db = conn.Connection()

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
	err = grade(payload)
	if err != nil {
		logs.Error(err)
		return err
	}
	err = removeAllfile(payload)
	if err != nil {
		logs.Error(err)
		return err
	}
	return nil
}

func writeTestCaseFile(payload schemas.Payload) error {
	dir, _ := os.Getwd()
	os.MkdirAll(fmt.Sprintf("%s/out/%s", dir, payload.Username), 0777)
	for i := 1; i <= payload.Testcase; i++ {
		index := i - 1
		input := []byte(payload.Input[index])
		if err := ioutil.WriteFile(fmt.Sprintf("%s/out/%s/%d.in", dir, payload.Username, i), input, 0775); err != nil {
			return err
		}
		output := []byte(payload.Output[index])
		if err := ioutil.WriteFile(fmt.Sprintf("%s/out/%s/%d.out", dir, payload.Username, i), output, 0775); err != nil {
			return err
		}
	}
	code := []byte(payload.Code)
	if err := os.WriteFile(fmt.Sprintf("%s/out/%s/Main.%s", dir, payload.Username, payload.Language), code, 0775); err != nil {
		return err
	}
	return nil
}

func compile(payload schemas.Payload) error {
	dir, _ := os.Getwd()
	script := fmt.Sprintf("%s/script/%s", dir, "compile.sh")
	err := os.Chmod(script, 0777)
	if err != nil {
		return err
	}
	fileCompile := fmt.Sprintf("%s/out/%s/Main.%s", dir, payload.Username, payload.Language)
	fileOutput := fmt.Sprintf("%s/out/%s/Main", dir, payload.Username)
	fileError := fmt.Sprintf("%s/out/%s/error", dir, payload.Username)
	res := exec.Command("/bin/sh", script, fileCompile, payload.Language, fileOutput, fileError)
	if res.Run() != nil {
		return res.Run()
	}
	return nil
}

func grade(payload schemas.Payload) error {
	dir, _ := os.Getwd()
	script := fmt.Sprintf("%s/script/%s", dir, "autograder.sh")
	err := os.Chmod(script, 0777)
	if err != nil {
		return err
	}
	fileLimit := fmt.Sprintf("%s/script/limit", dir)
	err = os.Chmod(fileLimit, 0777)
	if err != nil {
		return err
	}
	fileStudent := fmt.Sprintf("%s/out/%s/Main", dir, payload.Username)
	for i := 1; i <= payload.Testcase; i++ {
		compare := fmt.Sprintf("%s/out/%s/%d.out", dir, payload.Username, i)
		input := fmt.Sprintf("%s/out/%s/%d.in", dir, payload.Username, i)
		outStudent := fmt.Sprintf("%s/out/%s/student%d.out", dir, payload.Username, i)
		timeout := fmt.Sprintf("%d", payload.TimeLimit)
		mem := fmt.Sprintf("%d", payload.MemLimit)
		result := fmt.Sprintf("%s/out/%s/result", dir, payload.Username)
		res := exec.Command("/bin/sh", script, fileLimit, timeout, mem, fileStudent, input, outStudent, compare, result)
		if res.Run() != nil {
			return res.Run()
		}
	}
	return nil
}

func removeAllfile(payload schemas.Payload) error {
	dir, _ := os.Getwd()
	err := os.RemoveAll(fmt.Sprintf("%s/out/%s", dir, payload.Username))
	if err != nil {
		return err
	}
	return nil
}
