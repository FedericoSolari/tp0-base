import logging
from utils import Bet

def parse_bet_message(msg: str):
    try:
        parts = msg.strip().split(",")
        if len(parts) != 5:
            raise ValueError("Mensaje inválido")
        
        first_name, last_name, document, birthdate, number = parts
        # return Bet(agency, first_name, last_name, document, birthdate, number)
        return Bet("1", first_name, last_name, document, birthdate, number)
    except Exception as e:
        logging.error("action: parse_bet | result: fail | error: %s", e)
        return None

def format_response(ok: bool, msg: str = "") -> bytes:
    if ok:
        return f"OK|{msg}\n".encode()
    else:
        return f"ERROR|{msg}\n".encode()
