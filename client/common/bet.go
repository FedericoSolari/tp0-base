package common

import (
	"fmt"
	"os"
	"strconv"
)

// Bet representa una apuesta de usuario
type Bet struct {
	Agency    int
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    int
}

func NewBet(agency int, firstName, lastName, document, birthdate string, number int) *Bet {
	return &Bet{
		Agency:    agency,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
}
func NewBetFromEnvs() (*Bet, error) {
	agencystr := os.Getenv("CLI_ID")
	firstName := os.Getenv("NOMBRE")
	lastName := os.Getenv("APELLIDO")
	document := os.Getenv("DOCUMENTO")
	birth := os.Getenv("NACIMIENTO")
	numberStr := os.Getenv("NUMERO")

	agency, err1 := strconv.Atoi(agencystr)
	if err1 != nil {
		return nil, fmt.Errorf("Agency invalido: %v", err1)
	}

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
