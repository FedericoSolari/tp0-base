package common

import (
	"fmt"
	"strings"
)

func FormatBetMessage(b *Bet) string {
	// Formato que espera el server: "FirstName,LastName,DNI,Birthdate,Number\n"
	return fmt.Sprintf("%s,%s,%s,%s,%d\n", b.FirstName, b.LastName, b.Document, b.Birthdate, b.Number)
}

func IsSuccessResponse(response string) bool {
	return strings.TrimSpace(response) == "OK"
}
