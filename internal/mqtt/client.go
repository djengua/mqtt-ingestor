package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	inner  mqtt.Client
	logger *slog.Logger
	topics []string
	qos    byte
	h      Handler
}

func NewClient(broker, clientID, username, password string, topics []string, qos byte, logger *slog.Logger, handler Handler) *Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetOrderMatters(false)

	c := &Client{logger: logger, topics: topics, qos: qos, h: handler}

	opts.OnConnect = func(mc mqtt.Client) {
		logger.Info("mqtt connected")
		for _, topic := range topics {
			token := mc.Subscribe(topic, qos, func(_ mqtt.Client, msg mqtt.Message) {
				if err := handler(context.Background(), msg.Topic(), msg.Payload()); err != nil {
					logger.Error("failed to process mqtt message",
						slog.String("topic", msg.Topic()),
						slog.String("error", err.Error()),
					)
				}
			})
			token.Wait()
			if err := token.Error(); err != nil {
				logger.Error("failed to subscribe",
					slog.String("topic", topic),
					slog.String("error", err.Error()),
				)
				continue
			}
			logger.Info("mqtt subscribed", slog.String("topic", topic))
		}
	}

	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		logger.Warn("mqtt connection lost", slog.String("error", err.Error()))
	}

	c.inner = mqtt.NewClient(opts)
	return c
}

func (c *Client) Start(ctx context.Context) error {
	token := c.inner.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("connect mqtt: %w", err)
	}
	return nil
}

func (c *Client) Stop() {
	if c.inner.IsConnected() {
		c.inner.Disconnect(250)
	}
}

func (c *Client) IsConnected() bool {
	return c.inner.IsConnected()
}
