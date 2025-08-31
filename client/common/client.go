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
		return err
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

	// fmt.Printf(">>> Enviando (%d bytes): %q\n", len(data), data)

	for sent < total {
		n, err := c.conn.Write(data[sent:])
		if err != nil {
			return err
		}
		sent += n
	}

	return nil
}

func (c *Client) recvall(buffer []byte) (string, []byte, error) {
	if c.conn == nil {
		return "", nil, fmt.Errorf("no hay una conexion abierta")
	}

	// si buffer es nil, inicializamos uno nuevo
	if buffer == nil {
		buffer = make([]byte, 0, 1024)
	}

	tmp := make([]byte, 256)

	for {
		// busco un \n en el buffer acumulado
		if idx := bytes.IndexByte(buffer, '\n'); idx != -1 {
			// devuelvo hasta el '\n' incluido
			line := buffer[:idx+1]
			// guardo lo que sobra después del '\n'
			rest := buffer[idx+1:]
			return string(line), rest, nil
		}

		// leo datos del socket
		n, err := c.conn.Read(tmp)
		if err != nil {
			if err == io.EOF {
				return "", buffer, io.EOF
			}
			return "", buffer, fmt.Errorf("error leyendo del socket: %w", err)
		}

		if n == 0 {
			// socket cerrado
			if len(buffer) > 0 {
				return string(buffer), nil, nil
			}
			return "", nil, fmt.Errorf("socket cerrado")
		}

		buffer = append(buffer, tmp[:n]...)
	}
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

func (c *Client) ProcessBets(bets []*Bet) error {
	var leftover []byte

	msg := FormatBatchMessage(bets)

	err := c.sendall([]byte(msg))
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	response, leftover, err := c.recvall(leftover)
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

func (c *Client) ProcessAllBets() error {
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

		err = c.ProcessBets(batch)
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

	err := c.createClientSocket()
	if err != nil {
		log.Errorf("action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	//  NO importa como termine la funcion al final libero todo
	defer c.close_connections()

	if err := c.SendStart(); err != nil {
		log.Infof("ERROR EN EL SENDSTART")
		return
	}

	if err := c.ProcessAllBets(); err != nil {
		log.Infof("ERROR EN EL PROCESSALLBETS")
		return
	}

	if err := c.SendFinish(); err != nil {
		log.Infof("ERROR EN EL SENDFINISH")
		return
	}

	// espero que el server haya recibido nuestro final para poder cerrar
	response, _, err := c.recvall(nil)
	if !IsEndResponse(response) {
		log.Infof("action: recv_finish_mesagge | result: fail_response")
	}
	if err != nil {
		log.Errorf("action: recv_finish_mesagge | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
	}

	time.Sleep(100 * time.Millisecond)
}

func (c *Client) handle_SIGTERM_signal(sigs chan os.Signal) {
	if c.conn != nil {
		err := c.conn.Close()
		if err == nil {
			log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
		}
	}
	log.Infof("action: close_client | result: success | client_id: %v", c.config.ID)
}
