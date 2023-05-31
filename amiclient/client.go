package amiclient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/Arten331/observability/logger"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	ErrConnectionFailed         = errors.New("TCP connection failed")
	ErrClientDisabledBySettings = errors.New("ami client disabled by settings")
	ErrAuthTimeOut              = errors.New("auth timeout")
)

type Settings struct {
	ServiceName       string
	Host              string
	Port              int
	Username          string
	Password          string
	ConnectionTimeout time.Duration
	Disabled          bool
	ReadTimeOut       time.Duration
}

type Client struct {
	settings   *Settings
	conn       net.Conn
	msgChan    chan Message
	errChan    chan error
	stopReader chan interface{}
	metrics    *Metrics
}

func New(cfg *Settings) *Client {
	c := &Client{
		settings: cfg,
		metrics:  newMetrics(cfg.ServiceName),
	}

	c.msgChan = make(chan Message, 100)
	c.errChan = make(chan error, 1)
	c.stopReader = make(chan interface{}, 1)

	return c
}

func (c *Client) Connect(ctx context.Context, runReader bool) error {
	var err error

	if c.conn != nil {
		_ = c.conn.Close()
	}

	if c.Disabled() {
		return ErrClientDisabledBySettings
	}

	err = c.openConnection(ctx)
	if err != nil {
		return err
	}

	err = c.auth(ctx)
	if err != nil {
		return err
	}

	logger.L().Info("Client " + c.conn.LocalAddr().String() + " connected to " + c.conn.RemoteAddr().String())

	if runReader {
		go c.runReader(ctx)
		<-time.After(time.Millisecond * 50)
	}

	return nil
}

func (c *Client) Disconnect() {
	defer func() {
		err := recover()
		if err != nil {
			logger.L().Info("ignore panic on disconnect AMI client", zap.Any("recover", err))

			return
		}
	}()

	if c.conn != nil {
		_ = c.conn.Close()
		logger.L().Info("AMI client disconected", zap.String("address", c.conn.RemoteAddr().String()))
	}

	close(c.stopReader)
	close(c.msgChan)
	close(c.errChan)
}

func (c *Client) openConnection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.settings.ConnectionTimeout)
	defer cancel()

	c.metrics.StoreConnectionCount()

	dialer := net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp",
		net.JoinHostPort(c.settings.Host, strconv.Itoa(c.settings.Port)))
	if err != nil {
		logger.L().Error(
			"Failed to connect to tcp server",
			zap.Error(err),
		)

		return ErrConnectionFailed
	}

	logger.L().Info(fmt.Sprintf("open connection %s", conn.RemoteAddr()))

	c.conn = conn

	return nil
}

func (c *Client) auth(ctx context.Context) error {
	authCommand := Action{
		"Action":   "Login",
		"Username": c.settings.Username,
		"Secret":   c.settings.Password,
	}

	err := c.SendCommand(authCommand)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(c.conn)
	msgChan := make(chan Message, 100)
	errChan := make(chan error, 1)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			errChan <- ErrAuthTimeOut

			return ErrAuthTimeOut
		case err = <-errChan:
			return err
		case msg := <-msgChan:
			if msg["Response"] != "Success" && msg["Message"] != "Authentication accepted" {
				logger.S().Info("ami auth: receive message", msg)

				if msg["Response"] == "Error" {
					return fmt.Errorf("authenfication failed: %s", msg["Message"])
				}

				continue
			}

			logger.L().Info("Authentication accepted, User: %s", zap.String("username", c.settings.Username))
			cancel()

			return nil
		default:
			msg, errRead := ReadMessage(reader)
			if errRead != nil {
				errChan <- errRead
			}

			if len(msg) == 0 {
				continue
			}

			msgChan <- msg
		}
	}
}

func (c *Client) SendCommand(command Action) error {
	commandBytes := command.Serialize()
	_ = c.conn.SetWriteDeadline(time.Now().Add(time.Second))
	n, err := c.conn.Write(commandBytes)

	if n != len(commandBytes) {
		return fmt.Errorf("command not send, %d bytes writed", n)
	}

	c.metrics.StoreSentMessage()

	return err
}

func (c *Client) GetMetrics() []prometheus.Collector {
	return c.metrics.getMetrics()
}

func (c *Client) MsgChan() chan Message {
	return c.msgChan
}

func (c *Client) ErrChan() chan error {
	return c.errChan
}

func (c *Client) Addr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Client) Disabled() bool {
	return c.settings.Disabled
}
