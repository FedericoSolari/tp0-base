package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

const MaxBatchBytes = 8 * 1024 // 8 KB

type BetLoader struct {
	file       *os.File
	reader     *csv.Reader
	batch      int
	pendingBet *Bet
}

func NewBetLoader(path string, batchSize int) (*BetLoader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &BetLoader{
		file:       f,
		reader:     csv.NewReader(f),
		batch:      batchSize,
		pendingBet: nil,
	}, nil
}

func validateField(value string, fieldName string) int {
	if value == "" {
		return 0
	}
	fieldvalue, err := strconv.Atoi(value)
	if err != nil {
		fmt.Printf("valor invalido en %s: %v", fieldName, err)
		fieldvalue = 0
	}
	return fieldvalue
}

func (bl *BetLoader) flushPendingBet(bets []*Bet, currentBatchSize *int) []*Bet {
	if bl.pendingBet != nil {
		betSize := len(FormatBetMessage(bl.pendingBet))
		if betSize <= MaxBatchBytes {
			bets = append(bets, bl.pendingBet)
			*currentBatchSize += betSize + len(BetSeparator)
			bl.pendingBet = nil
		}
	}
	return bets
}

// Cierra el archivo al terminar
func (bl *BetLoader) Close() error {
	return bl.file.Close()
}

func (bl *BetLoader) NextBatch() ([]*Bet, error) {
	var bets []*Bet
	ok := true
	currentBatchSize := 0
	agency := os.Getenv("CLI_ID")

	// Si hay una bet pendiente del batch anterior, la agrego
	bets = bl.flushPendingBet(bets, &currentBatchSize)

	for len(bets) < bl.batch {
		data, err := bl.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Si la fila tiene menos datos, los completo con vacio
		for len(data) < 6 {
			data = append(data, "")
		}

		num := validateField(data[0], "num")

		bet := NewBet(agency, data[1], data[2], data[3], data[4], num)

		bets, ok = bl.tryAddBet(bet, bets, &currentBatchSize)
		if !ok {
			break
		}
	}

	if len(bets) == 0 {
		return nil, io.EOF
	}

	return bets, nil
}

func (bl *BetLoader) tryAddBet(bet *Bet, bets []*Bet, currentBatchSize *int) ([]*Bet, bool) {
	betSize := len(FormatBetMessage(bet))
	if *currentBatchSize+betSize+len(BetSeparator)+len(BatchEnd) > MaxBatchBytes {
		// No entra en el batch actual, lo guardo para la proxima llamada
		bl.pendingBet = bet
		return bets, false
	}

	bets = append(bets, bet)
	*currentBatchSize += betSize + len(BetSeparator)
	return bets, true
}
