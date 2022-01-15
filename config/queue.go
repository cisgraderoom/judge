package config

import "os"



// SERVICE_NAME is the name of the service
var SERVICE = os.Getenv("APP_SERVICE")

// QueueConfig is the configuration for the queue
var QueueConfig = map[string]interface{}{
	"writer": map[string]interface{}{
		"exchange": map[string]interface{}{
			"name": "report.writer",
			"type": "topic",
			"options": map[string]interface{}{
				"durable": true,
			},
		},
		"queues": map[string]interface{}{
			"daily": map[string]interface{}{
				"name":         "report.writer.daily",
				"routing_keys": []string{"report.writer.daily.*"},
				"options": map[string]interface{}{
					"durable": true,
				},
			},
		},
	},
}
