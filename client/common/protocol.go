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

func IsEndResponse(response string) bool {
	return strings.TrimSpace(response) == "END"
}

func IsERRORResponse(response string) bool {
	return strings.TrimSpace(response) == "ERROR"
}

func IsBetWinnerResponse(response string) bool {
	return strings.TrimSpace(response) == "Winner"
}

func IsNoMoreWinnerResponse(response string) bool {
	return strings.TrimSpace(response) == "NoMoreWinners"
}

func IsBeginLotteryResponse(response string) bool {
	return strings.TrimSpace(response) == "beginLottery"
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

func IsWinnerResponse(response string) bool {
	resp := strings.TrimSpace(response)
	return strings.HasPrefix(resp, "Winner:")
}

func ParseWinnerDocument(response string) string {
	if !IsWinnerResponse(response) {
		return ""
	}

	resp := strings.TrimSpace(response)
	// separar despues de "Winner: "
	parts := strings.SplitN(resp, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	// parts[1] contiene el doc
	return strings.TrimSpace(parts[1])
}
