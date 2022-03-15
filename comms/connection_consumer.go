package comms

import (
	"errors"
	"fmt"

	"github.com/streadway/amqp"
)

// MessageBody is the struct for the body passed in the AMQP message. The type will be set on the Request header
type MessageBody struct {
	Data []byte
	Type string // message type name
}

// Message is the amqp request to publish
type Message struct {
	Queue         string
	ReplyTo       string
	ContentType   string
	CorrelationID string
	Priority      uint8
	Body          MessageBody
}

// Connection is the connection created
type Connection struct {
	uri      string
	name     string
	conn     *amqp.Connection
	channel  *amqp.Channel
	refers   []string
	exchange map[string]interface{}
	queues   map[string]interface{}
	consumer map[string]interface{}
	err      chan error
}

var (
	connectionPool = make(map[string]*Connection)
)

// NewConnection returns the new connection object
func NewConnection(uri string, name string, refers []string, exchange map[string]interface{}, queues map[string]interface{}, consumer map[string]interface{}) *Connection {
	if c, ok := connectionPool[name]; ok {
		return c
	}

	c := &Connection{
		uri:      uri,
		refers:   refers,
		exchange: exchange,
		queues:   queues,
		consumer: consumer,
		err:      make(chan error),
	}
	connectionPool[name] = c
	return c
}

// GetConnection returns the connection which was instantiated
func GetConnection(name string) *Connection {
	return connectionPool[name]
}

func (c *Connection) Connect() error {
	var err error
	c.conn, err = amqp.Dial(c.uri)
	if err != nil {
		return fmt.Errorf("error in creating rabbitmq connection with %s : %s", c.uri, err.Error())
	}
	go func() {
		<-c.conn.NotifyClose(make(chan *amqp.Error)) // Listen to NotifyClose
		c.err <- errors.New("Connection Closed")
	}()
	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %s", err)
	}
	if err := c.channel.ExchangeDeclare(
		c.exchange["name"].(string),                                         // name
		c.exchange["type"].(string),                                         // type
		c.exchange["options"].(map[string]interface{})["durable"].(bool),    // durable
		c.exchange["options"].(map[string]interface{})["autoDelete"].(bool), // auto-deleted
		c.exchange["options"].(map[string]interface{})["internal"].(bool),   // internal
		c.exchange["options"].(map[string]interface{})["noWait"].(bool),     // no-wait
		nil, // arguments
	); err != nil {
		return fmt.Errorf("error in Exchange Declare: %s", err)
	}
	return nil
}

func (c *Connection) BindQueue() error {
	for _, k := range c.refers {
		_, err := c.channel.QueueDeclare(
			c.queues[k].(map[string]interface{})["name"].(string),                                         // queue name
			c.queues[k].(map[string]interface{})["options"].(map[string]interface{})["durable"].(bool),    // durable
			c.queues[k].(map[string]interface{})["options"].(map[string]interface{})["autoDelete"].(bool), // delete when unused
			c.queues[k].(map[string]interface{})["options"].(map[string]interface{})["exclusive"].(bool),  // exclusive
			c.queues[k].(map[string]interface{})["options"].(map[string]interface{})["noWait"].(bool),     // no-wait
			nil, // arguments
		)
		if err != nil {
			return fmt.Errorf("error in declaring the queue %s", err)
		}

		for _, r := range c.queues[k].(map[string]interface{})["routing_keys"].([]string) {
			if err := c.channel.QueueBind(
				c.queues[k].(map[string]interface{})["name"].(string), // queue name
				r,                           // routing key
				c.exchange["name"].(string), // exchange
				c.queues[k].(map[string]interface{})["options"].(map[string]interface{})["noWait"].(bool), // no-wait
				nil, // arguments
			); err != nil {
				return fmt.Errorf("queue  Bind error: %s", err)
			}
		}
	}
	return nil
}

// Reconnect reconnects the connection
func (c *Connection) Reconnect() error {
	if err := c.Connect(); err != nil {
		return err
	}
	if err := c.BindQueue(); err != nil {
		return err
	}
	return nil
}

// Consume consumes the messages from the queues and passes it as map of chan of amqp.Delivery
func (c *Connection) Consume() (map[string]<-chan amqp.Delivery, error) {
	m := make(map[string]<-chan amqp.Delivery)
	for _, k := range c.refers {
		deliveries, err := c.channel.Consume(
			c.queues[k].(map[string]interface{})["name"].(string), // queue name
			c.consumer["tag"].(string),                            // consumer-tag
			c.consumer["autoAck"].(bool),                          // auto-ack
			c.consumer["exclusive"].(bool),                        // exclusive
			c.consumer["noLocal"].(bool),                          // no-local
			c.consumer["noWait"].(bool),                           // no-wait
			nil,                                                   // arguments
		)
		if err != nil {
			return nil, err
		}
		m[c.queues[k].(map[string]interface{})["name"].(string)] = deliveries
	}
	return m, nil
}

// HandleConsumedDeliveries handles the consumed deliveries from the queues. Should be called only for a consumer connection
func (c *Connection) HandleConsumedDeliveries(q string, delivery <-chan amqp.Delivery, fn func(Connection, string, <-chan amqp.Delivery)) {
	fmt.Println(q)
	for {
		go fn(*c, q, delivery)
		if err := <-c.err; err != nil {
			c.Reconnect()
			deliveries, err := c.Consume()
			if err != nil {
				panic(err) // raising panic if consume fails even after reconnecting
			}
			delivery = deliveries[q]
		}
	}
}
