package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Bet representa una apuesta de usuario
type Bet struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    int
}

func NewBet() *Bet {
	return &Bet{}
}

func NewBetFromEnvs() (*Bet, error) {
	agency := os.Getenv("CLI_ID")
	firstName := os.Getenv("NOMBRE")
	lastName := os.Getenv("APELLIDO")
	document := os.Getenv("DOCUMENTO")
	birth := os.Getenv("NACIMIENTO")
	numberStr := os.Getenv("NUMERO")

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return nil, fmt.Errorf("NUMERO invalido: %v", err)
	}

	return &Bet{
		Agency:    agency,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birth,
		Number:    number,
	}, nil
}

func (b *Bet) LoadBets(path string) ([]*Bet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	var bets []*Bet

	for {
		data, err := reader.Read()
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

		bet := &Bet{
			Agency:    data[0],
			FirstName: data[1],
			LastName:  data[2],
			Document:  data[3],
			Birthdate: data[4],
			Number:    num,
		}
		bets = append(bets, bet)
	}

	return bets, nil
}
