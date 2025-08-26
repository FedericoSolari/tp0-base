package common

import (
	"fmt"
	"os"
	"strconv"
)

// Bet representa una apuesta de usuario
type Bet struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    int
}

func NewBet() (*Bet, error) {
	// firstName := "fede"
	// lastName := "solari"
	// document := "42819254"
	// birth := "22/11/2000" // formato string
	// numberStr := "1234"
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
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birth,
		Number:    number,
	}, nil
}
