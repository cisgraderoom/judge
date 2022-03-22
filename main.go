package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

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
		logs.Info(fmt.Sprintf("🐳 Received message from %s", m.Queue))
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
		logs.Fatal(err)
		return err
	}
	if payload.Mode == "success" {
		err := successJob(payload)
		if err != nil {
			return err
		}
		return nil
	}
	err := writeTestCaseFile(payload)
	if err != nil {
		logs.Fatal(err)
		return err
	}
	err = compile(payload)
	if err != nil {
		if err.Error() == "CF" {
			return nil
		} else {
			logs.Fatal(err)
			return err
		}
	}
	err = grade(payload)
	if err != nil {
		logs.Fatal(err)
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
	os.MkdirAll(fmt.Sprintf("%s/out/%d", dir, payload.SubmissionId), 0777)
	for i := 1; i <= payload.Testcase; i++ {
		index := i - 1
		input := []byte(payload.Input[index])
		if err := ioutil.WriteFile(fmt.Sprintf("%s/out/%d/%d.in", dir, payload.SubmissionId, i), input, 0775); err != nil {
			return err
		}
		output := []byte(payload.Output[index])
		if err := ioutil.WriteFile(fmt.Sprintf("%s/out/%d/%d.out", dir, payload.SubmissionId, i), output, 0775); err != nil {
			return err
		}
	}
	code := []byte(payload.Code)
	if err := os.WriteFile(fmt.Sprintf("%s/out/%d/Main.%s", dir, payload.SubmissionId, payload.Language), code, 0775); err != nil {
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
	fileCompile := fmt.Sprintf("%s/out/%d/Main.%s", dir, payload.SubmissionId, payload.Language)
	fileOutput := fmt.Sprintf("%s/out/%d/Main", dir, payload.SubmissionId)
	fileError := fmt.Sprintf("%s/out/%d/error", dir, payload.SubmissionId)
	res := exec.Command("/bin/sh", script, fileCompile, payload.Language, fileOutput, fileError)
	if res.Run() != nil {
		return res.Run()
	}
	fe, err := os.ReadFile(fileError)
	if err != nil {
		return err
	}
	if string(fe) != "" {
		insertCompileError(string(fe), payload)
		return fmt.Errorf("%s", "CF")
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
	fileStudent := fmt.Sprintf("%s/out/%d/Main", dir, payload.SubmissionId)
	result := fmt.Sprintf("%s/out/%d/result", dir, payload.SubmissionId)
	for i := 1; i <= payload.Testcase; i++ {
		compare := fmt.Sprintf("%s/out/%d/%d.out", dir, payload.SubmissionId, i)
		input := fmt.Sprintf("%s/out/%d/%d.in", dir, payload.SubmissionId, i)
		outStudent := fmt.Sprintf("%s/out/%d/student%d.out", dir, payload.SubmissionId, i)
		timeout := fmt.Sprintf("%d", payload.TimeLimit)
		mem := fmt.Sprintf("%d", payload.MemLimit)
		res := exec.Command("/bin/sh", script, fileLimit, timeout, mem, fileStudent, input, outStudent, compare, result)
		if res.Run() != nil {
			return res.Run()
		}
	}
	data, err := os.ReadFile(result)
	if err != nil {
		return err
	}
	data = data[0 : len(data)-1]
	resultData := strings.Split(string(data), "")

	err = insertResult(resultData, payload)
	if err != nil {
		return err
	}
	return nil
}

func insertCompileError(fe string, payload schemas.Payload) error {
	if payload.Mode == "new" {
		err := db.Table("submission").Where("submission_id = ?", payload.SubmissionId).Updates(map[string]interface{}{
			"result": fe,
			"score":  0,
		})
		if err.Error != nil {
			return err.Error
		}
	} else {
		err := db.Table("submission").Create(map[string]interface{}{
			"username":   payload.Username,
			"classcode":  payload.Classcode,
			"problem_id": payload.ProblemId,
			"code":       payload.Code,
			"result":     fe,
			"score":      0.00,
			"lang":       payload.Language,
			"created_at": time.Now(),
		})
		if err.Error != nil {
			return err.Error
		}
	}
	if err := db.Table("score").Where("username = ?", payload.Username).Where("problem_id = ?", payload.ProblemId).Updates(map[string]interface{}{
		"score": 0,
	}).Error; err != nil {
		return err
	}
	return nil
}

func insertResult(data []string, payload schemas.Payload) error {
	var score float64
	for _, val := range data {
		if val == "P" {
			score += payload.MaxScore / float64(payload.Testcase)
		}
	}
	if payload.Mode == "new" {
		err := db.Table("submission").Where("submission_id = ?", payload.SubmissionId).Updates(map[string]interface{}{
			"result": strings.Join(data, ""),
			"score":  score,
		})
		if err.Error != nil {
			return err.Error
		}
	} else {
		err := db.Table("submission").Create(map[string]interface{}{
			"username":   payload.Username,
			"classcode":  payload.Classcode,
			"problem_id": payload.ProblemId,
			"code":       payload.Code,
			"result":     strings.Join(data, ""),
			"score":      score,
			"lang":       payload.Language,
			"created_at": time.Now(),
		})
		if err.Error != nil {
			return err.Error
		}
	}

	if err := db.Table("score").Where("username = ?", payload.Username).Where("problem_id = ?", payload.ProblemId).Updates(map[string]interface{}{
		"score": score,
	}).Error; err != nil {
		return err
	}

	return nil
}

func removeAllfile(payload schemas.Payload) error {
	dir, _ := os.Getwd()
	err := os.RemoveAll(fmt.Sprintf("%s/out/%d", dir, payload.SubmissionId))
	if err != nil {
		return err
	}
	return nil
}

func successJob(payload schemas.Payload) error {
	if err := db.Table("jobs").Where("id = ?", payload.JobId).Updates(map[string]interface{}{
		"status": true,
	}).Error; err != nil {
		return err
	}
	return nil
}
