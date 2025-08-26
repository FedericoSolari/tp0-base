import logging
from common import utils
import datetime

def parse_bet_message(bytes_msg: bytes):
    try:
        msg = bytes_msg.rstrip().decode('utf-8')
        parts = msg.strip().split(",")

        if len(parts) != 5:
            raise ValueError("Mensaje invalido")
        
        agency_str,first_name, last_name, document, birthdate_str, number_str = parts

        agency = int(agency_str) 
        birthdate = datetime.datetime.strptime(birthdate_str, "%d/%m/%Y").date()
        number = int(number_str) 

        return utils.Bet(agency, first_name, last_name, document, birthdate.isoformat(), number)
    except Exception as e:
        logging.error("action: parse_bet | result: fail | error: %s", e)
        return None

def success_message() -> bytes:
    return b"OK\n"
