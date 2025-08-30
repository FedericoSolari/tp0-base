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

// Cierra el archivo al terminar
func (bl *BetLoader) Close() error {
	return bl.file.Close()
}

func (bl *BetLoader) NextBatch() ([]*Bet, error) {
	var bets []*Bet
	currentBatchSize := 0

	// Si hay una bet pendiente del batch anterior, la agrego
	if bl.pendingBet != nil {
		betSize := len(FormatBetMessage(bl.pendingBet))
		if betSize <= MaxBatchBytes {
			bets = append(bets, bl.pendingBet)
			currentBatchSize += betSize + len(BetSeparator) // consideramos separador
			bl.pendingBet = nil
		}
	}

	for len(bets) < bl.batch {
		data, err := bl.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		num, err := strconv.Atoi(data[5])
		if err != nil {
			return nil, fmt.Errorf("valor invalido en Number: %v", err)
		}

		bet := NewBet(data[0], data[1], data[2], data[3], data[4], num)

		betSize := len(FormatBetMessage(bet))
		if currentBatchSize+betSize+len(BetSeparator)+len(BatchEnd) > MaxBatchBytes {
			// No entra en el batch actual, la guardo para la proxima llamada
			bl.pendingBet = bet
			break
		}

		bets = append(bets, bet)
		currentBatchSize += betSize + len(BetSeparator)
	}

	if len(bets) == 0 {
		return nil, io.EOF
	}

	return bets, nil
}
