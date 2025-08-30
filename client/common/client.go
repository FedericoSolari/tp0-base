package common

import (
	"bytes"
	"fmt"
	"io"
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
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
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
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) sendall(data []byte) error {
	if c.conn == nil {
		return fmt.Errorf("no hay una conexion abierta")
	}

	total := len(data)
	sent := 0

	for sent < total {
		n, err := c.conn.Write(data[sent:])
		if err != nil {
			return err
		}
		sent += n
	}

	return nil
}

func (c *Client) recvall() (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("no hay una conexion abierta")
	}

	buffer := make([]byte, 0, 1024)
	tmp := make([]byte, 256)
	found := false
	for !found {
		//
		n, err := c.conn.Read(tmp)
		if err != nil {
			return "", err
		}

		buffer = append(buffer, tmp[:n]...)

		// verifico si ya lei el '\n'
		if bytes.Contains(tmp[:n], []byte{'\n'}) {
			found = true
		}
	}

	return string(buffer), nil
}

func (c *Client) SendStart() error {
	err := c.sendall([]byte(startMessage()))
	if err != nil {
		log.Errorf("action: Send_start | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}
	return nil
}

func (c *Client) SendFinish() error {
	err := c.sendall([]byte(AllBetsDone()))
	if err != nil {
		log.Errorf("action: SendFinish | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}
	return nil
}

func (c *Client) ProcessBets(bets []*Bet) {

	msg := FormatBatchMessage(bets)

	err := c.sendall([]byte(msg))
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	response, err := c.recvall()
	if err != nil {
		log.Errorf("action: recv_response | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	if IsSuccessResponse(response) {
		log.Infof("action: batch_de_apuestas_enviado | result: success | cantidad: %d", len(bets))
	} else {
		log.Infof("action: batch_de_apuestas_NO_enviado | result: Fail ")
	}
}

func (c *Client) ProcessAllBets() {
	loader, err := NewBetLoader("/data/agency.csv", c.config.BatchMaxAmount)
	if err != nil {
		log.Errorf("action: load_bets | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}
	defer loader.Close()

	for {
		batch, err := loader.NextBatch()
		if err == io.EOF {
			break // no quedan más batches
		}
		if err != nil {
			log.Errorf("action: load_batch | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return
		}

		c.ProcessBets(batch)
	}
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		c.handle_SIGTERM_signal(sigs)
		os.Exit(0)
	}()

	err := c.createClientSocket()
	if err != nil {
		log.Errorf("action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}
	defer c.conn.Close() // Al salir de la func cierro el skt

	c.SendStart()
	c.ProcessAllBets()
	c.SendFinish()

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
