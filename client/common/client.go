package common

import (
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
func (c *Client) createClientSocket() (*ConnectionHandler, error) {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Errorf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return nil, err
	}

	c.conn = conn
	return NewConnectionHandler(conn), nil
}

func (c *Client) SendStart(handler *ConnectionHandler) error {
	if err := handler.SendAll([]byte(startMessage())); err != nil {
		log.Errorf("action: Send_start | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}
	return nil
}

func (c *Client) SendFinish(handler *ConnectionHandler) error {

	if err := handler.SendAll([]byte(AllBetsDone())); err != nil {
		log.Errorf("action: SendFinish | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}
	return nil

}

func (c *Client) ProcessBets(handler *ConnectionHandler, bets []*Bet) error {

	msg := FormatBatchMessage(bets)

	err := handler.SendAll([]byte(msg))
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	response, err := handler.RecvAll()
	if err != nil {
		log.Errorf("action: recv_response | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	if !IsSuccessResponse(response) {
		log.Infof("action: batch_de_apuestas_NO_enviado | result: Fail ")
		return fmt.Errorf("el servidor respondio con fallo: %q", response)
	}
	return nil
}

func (c *Client) ProcessAllBets(handler *ConnectionHandler) error {
	loader, err := NewBetLoader("/data/agency.csv", c.config.BatchMaxAmount)
	if err != nil {
		log.Errorf("action: load_bets | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
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
			return err
		}

		err = c.ProcessBets(handler, batch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) close_connections() {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			log.Errorf("Error cerrando la conexión: %v", err)
		} else {
			log.Infof("Conexión cerrada correctamente")
		}
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

	handler, err := c.createClientSocket()
	if err != nil {
		return
	}
	defer c.close_connections()

	c.runClientSession(handler)
}

func (c *Client) handle_SIGTERM_signal() {
	if c.conn != nil {
		err := c.conn.Close()
		if err == nil {
			log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
		}
	}
	log.Infof("action: close_client | result: success | client_id: %v", c.config.ID)
}

func (c *Client) runClientSession(handler *ConnectionHandler) {
	if err := c.SendStart(handler); err != nil {
		log.Infof("ERROR EN EL SENDSTART")
		return
	}

	if err := c.ProcessAllBets(handler); err != nil {
		log.Infof("ERROR EN EL PROCESSALLBETS")
		return
	}

	if err := c.SendFinish(handler); err != nil {
		log.Infof("ERROR EN EL SENDFINISH")
		return
	}

	// Espero que el server haya recibido nuestro final para poder cerrar
	response, err := handler.RecvAll()
	if err != nil {
		log.Errorf("action: recv_finish_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if !IsEndResponse(response) {
		log.Infof("action: recv_finish_message | result: fail_response | response: %q", response)
	}
}
