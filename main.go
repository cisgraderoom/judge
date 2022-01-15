package main

func main() {
	// // Start Report Service
	// logger.LogInfo("starter", fmt.Sprintf("🐳 GO_ENV: %s, APP_ENV: %s", os.Getenv("GO_ENV"), os.Getenv("APP_ENV")))

	// // Connect RabbitMQ Server
	// rabbitmqConfig := configs.RabbitMQ_URL()
	// conn, err := amqp.Dial(rabbitmqConfig)

	// if err != nil {
	// 	logger.LogFatal("main connect rabbitmq", err)
	// }

	// defer conn.Close()

	// // Coonect RabbitMQ Success Log
	// logger.LogInfo("rabbitmq", "Successfully Connected to RabbitMQ")

	// // get service queue config from service
	// serviceConfig := configs.QueueConfig[configs.SERVICE].(map[string]interface{})["queues"].(map[string]interface{})

	// for key, _ := range serviceConfig {
	// 	go helper.GetChannelCB(configs.SERVICE, key, conn)
	// }

	// select {}
}
