package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Handler func(ctx context.Context, topic string, payload []byte) error

type Client struct {
	inner  paho.Client
	logger *slog.Logger
	topics []string
	qos    byte
	h      Handler
}

func NewClient(
	broker string,
	clientID string,
	username string,
	password string,
	topics []string,
	qos byte,
	logger *slog.Logger,
	handler Handler,
) *Client {
	opts := paho.NewClientOptions()
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

	c := &Client{
		logger: logger,
		topics: topics,
		qos:    qos,
		h:      handler,
	}

	opts.OnConnect = func(mc paho.Client) {
		logger.Info("mqtt connected")

		for _, topic := range topics {
			token := mc.Subscribe(topic, qos, func(_ paho.Client, msg paho.Message) {
				if err := handler(context.Background(), msg.Topic(), msg.Payload()); err != nil {
					logger.Error(
						"failed to process mqtt message",
						slog.String("topic", msg.Topic()),
						slog.String("error", err.Error()),
					)
				}
			})

			token.Wait()
			if err := token.Error(); err != nil {
				logger.Error(
					"failed to subscribe",
					slog.String("topic", topic),
					slog.String("error", err.Error()),
				)
				continue
			}

			logger.Info("mqtt subscribed", slog.String("topic", topic))
		}
	}

	opts.OnConnectionLost = func(_ paho.Client, err error) {
		logger.Warn("mqtt connection lost", slog.String("error", err.Error()))
	}

	c.inner = paho.NewClient(opts)
	return c
}

func (c *Client) Start(ctx context.Context) error {
	_ = ctx

	token := c.inner.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("connect mqtt: %w", err)
	}

	return nil
}

func (c *Client) Stop() {
	if c.inner != nil && c.inner.IsConnected() {
		c.inner.Disconnect(250)
	}
}

func (c *Client) IsConnected() bool {
	if c.inner == nil {
		return false
	}
	return c.inner.IsConnected()
}
