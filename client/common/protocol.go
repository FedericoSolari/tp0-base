package common

import (
	"fmt"
	"strings"
)

const (
	BetSeparator = "|"
	EndDelimiter = "\n"
)

func FormatBetMessage(b *Bet) string {
	// Formato que espera el server: "Agency,FirstName,LastName,DNI,Birthdate,Number\n"
	return fmt.Sprintf("%s,%s,%s,%s,%s,%d", b.Agency, b.FirstName, b.LastName, b.Document, b.Birthdate, b.Number)
}

func IsSuccessResponse(response string) bool {
	return strings.TrimSpace(response) == "OK"
}

func IsERRORResponse(response string) bool {
	return strings.TrimSpace(response) == "ERROR"
}

func startMessage() string {
	return "START" + EndDelimiter
}

func AllBetsDone() string {
	return "END" + EndDelimiter
}

func FormatBatchMessage(bets []*Bet) string {
	var msgs []string
	for _, bet := range bets {
		msgs = append(msgs, FormatBetMessage(bet))
	}
	return strings.Join(msgs, BetSeparator) + EndDelimiter
}
