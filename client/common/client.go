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
// func (c *Client) createClientSocket() error {
// 	conn, err := net.Dial("tcp", c.config.ServerAddress)
// 	if err != nil {
// 		log.Criticalf(
// 			"action: connect | result: fail | client_id: %v | error: %v",
// 			c.config.ID,
// 			err,
// 		)
// 		return err
// 	}
// 	c.conn = conn
// 	return nil
// }

// func (c *Client) SendStart() error {
// 	err := c.sendall([]byte(startMessage()))
// 	if err != nil {
// 		log.Errorf("action: Send_start | result: fail | client_id: %v | error: %v",
// 			c.config.ID, err)
// 		return err
// 	}
// 	return nil
// }

// func (c *Client) SendFinish() error {
// 	err := c.sendall([]byte(AllBetsDone()))
// 	if err != nil {
// 		log.Errorf("action: SendFinish | result: fail | client_id: %v | error: %v",
// 			c.config.ID, err)
// 		return err
// 	}
// 	return nil
// }

func (c *Client) waitBeginLottery(handler *ConnectionHandler) (bool, error) {
	response, err := handler.RecvAll()
	if err != nil {
		log.Errorf("action: waitBeginLottery | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return false, err
	}

	if !IsBeginLotteryResponse(response) {
		log.Infof("action: getWinners | result: Fail ")
		return false, fmt.Errorf("el servidor respondio con fallo: %q", response)
	}

	return true, nil
}

func (c *Client) receiveWinners(handler *ConnectionHandler) ([]string, error) {
	var winners []string

	for {
		response, err := handler.RecvAll()
		if err != nil {
			log.Errorf("action: receiveWinners | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return winners, err
		}

		if IsWinnerResponse(response) {
			doc := ParseWinnerDocument(response)
			winners = append(winners, doc)
			log.Infof("action: Lottery | winner:%v", doc)
		} else if IsNoMoreWinnerResponse(response) {
			log.Infof("action: NO MORE WINNERS")
			break
		} else {
			log.Infof("action: RECEIVE_WINNERS | result: Fail ")
			return winners, fmt.Errorf("el servidor respondio con fallo: %q", response)
		}
	}

	return winners, nil
}

func (c *Client) getWinners(handler *ConnectionHandler) error {
	beginLottery, err := c.waitBeginLottery(handler)
	if err != nil {
		return err
	}

	if !beginLottery {
		return fmt.Errorf("La loteria no comenzo correctamente")
	}

	winners, err := c.receiveWinners(handler)
	if err != nil {
		return err
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: ${%d}", len(winners))
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
			break // no quedan mas batches
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

	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Errorf("action: connect | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	//encapsular
	defer func() {
		if conn != nil {
			err := conn.Close()
			if err != nil {
				log.Errorf("Error cerrando la conexión: %v", err)
			} else {
				log.Infof("Conexión cerrada correctamente")
			}
		}
	}()

	handler := NewConnectionHandler(conn)

	//  NO importa como termine la funcion al final libero todo
	// defer c.close_connections()

	c.runClientSession(handler)

	time.Sleep(500 * time.Millisecond)
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

func (c *Client) runClientSession(handler *ConnectionHandler) {
	//encapsular
	if err := handler.SendAll([]byte(startMessage())); err != nil {
		log.Infof("ERROR EN EL SENDSTART: %v", err)
		return
	}

	//encapsular
	if err := c.ProcessAllBets(handler); err != nil {
		log.Infof("ERROR EN EL PROCESSALLBETS: %v", err)
		return
	}

	if err := handler.SendAll([]byte(AllBetsDone())); err != nil {
		log.Infof("ERROR EN EL SENDFINISH: %v", err)
		return
	}

	if err := c.getWinners(handler); err != nil {
		log.Infof("ERROR EN EL getWinners: %v", err)
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
