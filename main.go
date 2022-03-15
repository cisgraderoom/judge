package main

import (
	"encoding/json"
	"fmt"
	"os"

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
		logs.Info(fmt.Sprintf("data : %v", string(m.Body.Data[:])))
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

	return nil
}

// Judge - function for judge code, call from submission.controller.go
// func Judge(compileFile, fileC, fileOut, numberOfTestCase, testCase, OutDir, fileSol, submissionID string, score float64, timeout int, teacher string, problemID string, username string) error {
// 	var result string
// 	var rScore float64
// 	runningDir, _ := os.Getwd()
// 	submission := models.Submission{}

// 	// compile

// 	compile := exec.Command(compileFile, fileC, fileOut)
// 	err := compile.Run()
// 	if err != nil {
// 		if err := database.GetDB().Where("id = ?", submissionID).First(&submission); err.Error != nil {
// 			return err.Error
// 		}
// 		submission.Err = err.Error()
// 		submission.Status = "error"
// 		submission.Result = "compile error"
// 		if err := database.GetDB().Save(&submission); err.Error != nil {
// 			return err.Error
// 		}
// 		return err
// 	}

// 	// runtime
// 	num, err := strconv.Atoi(numberOfTestCase)
// 	for i := 1; i <= num; i++ {
// 		var t = time.Now()
// 		runtime := exec.Command(runningDir+"/runtime.sh", fmt.Sprint(timeout), testCase, OutDir+"/"+fmt.Sprint(i)+".in", fileSol+"/"+fmt.Sprint(i)+".out")
// 		err := runtime.Run()
// 		if err != nil {
// 			println(err.Error())
// 			return err
// 		}
// 		elapsed := time.Since(t)
// 		println(elapsed)
// 		// check
// 		fileTeacher, err := ioutil.ReadFile(fmt.Sprintf("%s/%d.out", OutDir, i))
// 		if err != nil {
// 			return err
// 		}
// 		fileStudent, err := ioutil.ReadFile(fmt.Sprintf("%s/%d.out", fileSol, i))
// 		if err != nil {
// 			return err
// 		}
// 		if bytes.Equal(fileTeacher, fileStudent) {
// 			result += "P"
// 			rScore += score / float64(num)
// 		} else if elapsed > (time.Duration(timeout) * 1e9) {
// 			result += "T"
// 		} else {
// 			result += "-"
// 		}
// 	}
// 	if err := database.GetDB().Where("id = ?", submissionID).First(&submission); err.Error != nil {
// 		return err.Error
// 	}
// 	submission.Status = "successful"
// 	submission.Result = result
// 	submission.Score = rScore

// 	if err := database.GetDB().Save(&submission); err.Error != nil {
// 		return err.Error
// 	}

// 	sql := fmt.Sprintf("UPDATE section_%s SET %s=%v WHERE username='%s';", teacher, problemID, rScore, username)
// 	fmt.Print(sql)
// 	if err := database.GetDB().Exec(sql); err.Error != nil {
// 		return err.Error
// 	}

// 	return err
// }
