import logging
from common import utils
import datetime

def parse_bet_message(msg: str):
    try:
        parts = msg.strip().split(",")
        if len(parts) != 5:
            raise ValueError("Mensaje invalido")
        
        first_name, last_name, document, birthdate_str, number_str = parts

        birthdate = datetime.datetime.strptime(birthdate_str, "%d/%m/%Y").date()
        number = int(number_str) 

        # return Bet(agency, first_name, last_name, document, birthdate, number)
        return utils.Bet("1", first_name, last_name, document, birthdate.isoformat(), number)
    except Exception as e:
        logging.error("action: parse_bet | result: fail | error: %s", e)
        return None

def format_response(ok: bool, msg: str = "") -> bytes:
    if ok:
        return f"OK|{msg}\n".encode()
    else:
        return f"ERROR|{msg}\n".encode()
