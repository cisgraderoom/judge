package config

import (
	"fmt"
	"os"
)

// RabbitMQDefaultConfig is the default RabbitMQ configuration
type RabbitMQDefaultConfig struct {
	protocol string
	hostname string
	port     string
	vhost    string
	username string
	password string
}

// GetDefaultConfig returns a default RabbitMQ configuration
func RabbitMQ_URL() string {
	rabbitmq := RabbitMQDefaultConfig{
		protocol: os.Getenv("CIS_RABBITMQ_PROTOCAL"),
		hostname: os.Getenv("CIS_RABBITMQ_HOST"),
		port:     os.Getenv("CIS_RABBITMQ_PORT"),
		vhost:    os.Getenv("APP_RABBITMQ_VHOST"),
		username: os.Getenv("RABBITMQ_DEFAULT_USER"),
		password: os.Getenv("RABBITMQ_DEFAULT_PASS"),
	}
	return fmt.Sprintf("%s://%s:%s@%s:%s/", rabbitmq.protocol, rabbitmq.username, rabbitmq.password, rabbitmq.hostname, rabbitmq.port)
}
