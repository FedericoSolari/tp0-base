import logging
from common import utils
import datetime

BET_SEPARATOR = "|"
BATCH_END = "\n"

def parse_bet_message(bytes_msg: bytes):
    try:
        msg = bytes_msg.rstrip(BATCH_END.encode()).decode('utf-8')
        bet_strs = msg.split(BET_SEPARATOR)  # separar apuestas dentro del batch

        bets = []
        for bet_str in bet_strs:
            parts = bet_str.strip().split(",")
            if len(parts) != 6:
                raise ValueError(f"Mensaje invalido: {bet_str}")
            
            agency_str, first_name, last_name, document, birthdate_str, number_str = parts

            agency = int(agency_str)
            birthdate = datetime.datetime.strptime(birthdate_str, "%Y-%m-%d").date()
            number = int(number_str)

            bets.append(utils.Bet(agency, first_name, last_name, document, birthdate.isoformat(), number))

        return bets

    except Exception as e:
        logging.error("action: parse_bet | result: fail | error: %s", e)
        return None

def success_message() -> bytes:
    return b"OK\n"

def success_message() -> bytes:
    return b"ERROR\n"
