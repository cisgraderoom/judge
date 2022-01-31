package config

import "os"

// SERVICE_NAME is the name of the service
var SERVICE = os.Getenv("APP_SERVICE")

// QueueConfig is the configuration for the queue
var QueueConfig = map[string]interface{}{
	"judge": map[string]interface{}{
		"exchange": map[string]interface{}{
			"name": "cisgraderoom.judge",
			"type": "topic",
			"options": map[string]interface{}{
				"durable": true,
			},
		},
		"queues": map[string]interface{}{
			"sender": map[string]interface{}{
				"name":         "cisgraderoom.judge.sender",
				"routing_keys": []string{"cisgraderoom.judge.sender.*"},
				"options": map[string]interface{}{
					"durable": true,
				},
			},
			"result": map[string]interface{}{
				"name":         "cisgraderoom.judge.result",
				"routing_keys": []string{"cisgraderoom.judge.result.*"},
				"options": map[string]interface{}{
					"durable": true,
				},
			},
		},
	},
}
