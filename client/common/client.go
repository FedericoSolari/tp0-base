package common

import (
	"bytes"
	"fmt"
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
	// log.Println("Inicio sendall")
	if c.conn == nil {
		return fmt.Errorf("No hay una conexion abierta")
	}

	total := len(data)
	sent := 0

	for sent < total {
		// log.Println("enviando mensaje")
		n, err := c.conn.Write(data[sent:])
		if err != nil {
			return err
		}
		//log.Printf("enviado %d bytes: %q\n", n, data[sent:sent+n])
		sent += n
	}
	// log.Println("mensaje enviado")

	return nil
}

func (c *Client) recvAll() (string, error) {
	// log.Println("inicio recv all")
	if c.conn == nil {
		return "", fmt.Errorf("No hay una conexion abierta")
	}

	buffer := make([]byte, 0, 1024)
	tmp := make([]byte, 256)
	found := false
	// log.Println("variables seteadas")

	for !found {
		//
		n, err := c.conn.Read(tmp)
		if err != nil {
			return "", err
			// log.Println("NO LEI NADA") err // EOF o error
		}

		buffer = append(buffer, tmp[:n]...)

		//log.Println("leídos %d bytes: %q\n", n, tmp[:n])

		// verifico si ya lei el '\n'
		if bytes.Contains(tmp[:n], []byte{'\n'}) {
			found = true
		}
	}

	return string(buffer), nil
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

	bet, err := NewBet()
	if err != nil {
		log.Errorf("action: create_bet | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	err = c.createClientSocket()
	if err != nil {
		log.Errorf("action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}
	defer c.conn.Close() // Al salir de la func cierro el skt

	msg := bet.FormatMessage()

	err = c.sendall([]byte(msg))
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	response, err := c.recvAll()
	if err != nil {
		log.Errorf("action: recv_response | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	if bet.IsExpectedResponse(response) {
		log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %s",
			bet.Document, bet.Number)
	}
	//else {
	//log.Warnf("action: check_response | result: rejected | client_id: %v | response: %s",
	//	c.config.ID, response)
	//}
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
