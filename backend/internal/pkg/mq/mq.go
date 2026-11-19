package mq

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func New(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	// 声明一个持久化队列（生产者/消费者共用）
	if _, err := ch.QueueDeclare("notifications", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	slog.Info("rabbitmq connected")
	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) Publish(queue string, body []byte) error {
	return c.ch.PublishWithContext(context.Background(), "", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (c *Client) Consume(queue string) (<-chan amqp.Delivery, error) {
	return c.ch.Consume(queue, "", false, false, false, false, nil)
}

func (c *Client) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
