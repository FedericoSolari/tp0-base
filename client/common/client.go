package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
// func (c *Client) createClientSocket() *ConnectionHandler ,error {
// 	conn, err := net.Dial("tcp", c.config.ServerAddress)
// 	if err != nil {
// 		log.Criticalf(
// 			"action: connect | result: fail | client_id: %v | error: %v",
// 			c.config.ID,
// 			err,
// 		)
// 		return nil, err
// 	}
// 	c.conn = conn
// 	return nil
// }

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		c.handle_SIGTERM_signal(sigs)
		os.Exit(0)
	}()

	bet, err := NewBet()
	if err != nil {
		log.Errorf("action: create_bet | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Errorf("action: connect | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	defer func() {
		if conn != nil {
			err := conn.Close()
			if err != nil {
				log.Errorf("Error cerrando la conexión: %v", err)
			}
		}
	}()

	handler := NewConnectionHandler(conn)

	msg := FormatBetMessage(bet)

	err = handler.SendAll([]byte(msg))
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	response, err := handler.RecvAll()
	if err != nil {
		log.Errorf("action: recv_response | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	if IsSuccessResponse(response) {
		log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %s",
			bet.Document, bet.Number)
	}

}

func (c *Client) handle_SIGTERM_signal(sigs chan os.Signal) {
	if c.conn != nil {
		err := c.conn.Close()
		if err == nil {
			log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
		}
	}

	if sigs != nil {
		close(sigs)
		log.Infof("action: close_client | result: success | client_id: %v", c.config.ID)
	}
}
