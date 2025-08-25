package common

import (
	"fmt"
	"strconv"
	"strings"
)

// Bet representa una apuesta de usuario
type Bet struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string // ahora es string
	Number    int
}

func NewBet() (*Bet, error) {
	firstName := "fede"
	lastName := "solari"
	document := "42819254"
	birth := "22/11/2000" // formato string
	numberStr := "1234"
	// Descomentar para leer desde variables de entorno:
	// firstName := os.Getenv("NOMBRE")     
	// lastName := os.Getenv("APELLIDO")
	// document := os.Getenv("DOCUMENTO")
	// birth := os.Getenv("NACIMIENTO")
	// numberStr := os.Getenv("NUMERO")

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

func (b *Bet) FormatMessage() string {
	return fmt.Sprintf(
		"%s,%s,%s,%s,%d\n",
		b.FirstName,
		b.LastName,
		b.Document,
		b.Birthdate,
		b.Number,
	)
}

func (b *Bet) IsExpectedResponse(response string) bool {
	return strings.TrimSpace(response) == "OK"
}
