import socket
import logging
import signal
import sys
import time



class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        # Agrego time out para que no se qude esperando por siempre una conxion
        self._server_socket.settimeout(5.0)
        self._client_skts = []
        self.shutdown = False

        # Capturo el sigterm para hacer el handeleo 
        signal.signal(signal.SIGTERM, self.handle_sigterm_signal)

    def handle_sigterm_signal(self, signum, frame):

        # logging.info("action: handle_sigterm_signal | result: in progress")
        logging.info("action: handle_sigterm_signal | result: success")
        self.shutdown = True
        self.clean_resourses()
        
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while self.shutdown == False:
            try:
                logging.info(f'Server: action: connecting')
                client_sock = self.__accept_new_connection()
            except socket.timeout:
                # vuelvo a intentar obtener una conexion
                logging.info(f'Server: action: Socket_timeOut | reintentando')
                continue
            # Almaceno el socket del cliente
            self._client_skts.append(client_sock)

            self.__handle_client_connection(client_sock)



    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            time.sleep(0.5)
            # recibo todo y decodifico el mensaje
            msg = self.recv_all(client_sock).rstrip().decode('utf-8')
            # addr = client_sock.getpeername()

            bet_data = parse_bet_message(msg)
            if bet_data:
                logging.info(f'Server: action: apuesta_almacenada | result: success | dni: {bet_data["document"]} | numero: {bet_data["number"]}')
                # store_bet(
                #     bet_data["first_name"],
                #     bet_data["last_name"],
                #     bet_data["document"],
                #     bet_data["birthdate"],
                #     bet_data["number"]
                # )

            client_sock.send(b"OK\n")
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
        # Elimino el socket almacenado
        self._client_skts.remove(client_sock)

    def recv_all(self, client_sock):

        logging.info(f'Server: action: inicio recv_all')
        buffer = bytearray()
        found = False
        while found == False:
            logging.info(f'Server: action: leyendo mensaje')
            chunk = client_sock.recv(256)
            if not chunk:
                logging.error(f'action: Error en la lectura del mensaje')
                break
            buffer.extend(chunk)
            
            # busco el \n que significa el fin segun el protoolo definido
            if b'\n' in chunk:
                logging.info(f'Server: action: mensaje leido')
                found = True
            
        return bytes(buffer)
    
    def send_all(skt, data: bytes):
        """
        Envía todos los bytes del mensaje por el socket.
        Se asegura que todo se envíe, evitando short-write.
        data debe contener el '\n' al final para indicar fin de mensaje.
        """
        logging.info(f'Server: action: comienza send_all')
        total_sent = 0
        total_len = len(data)

        while total_sent < total_len:
            logging.info(f'Server: action: enviando mensaje')
            try:
                sent = skt.send(data[total_sent:])
                if sent == 0:
                    raise RuntimeError("socket connection broken")
                total_sent += sent
            except OSError as e:
                logging.error(f"Server: action: send_all | result: fail | error: {e}")
                raise

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('Server: action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'Server: action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def clean_resourses(self):
        logging.info('Server: Received SIGTERM signal')

        for client_sock in self._client_skts:
            logging.info('Server: Closing client connection')
            client_sock.close()
        
        self._server_socket.close()
        logging.info('Server: Server connection closed')
        
        logging.info('Server: Resources closed successfully')
        sys.exit(0)


    def parse_bet_message(message: str):
        """
        Recibe un mensaje de apuesta separado por comas:
        "FirstName,LastName,DNI,Birthdate,Number"
        y devuelve un diccionario con los campos.
        """
        try:
            parts = message.strip().split(',')
            if len(parts) != 5:
                raise ValueError("Mensaje con cantidad de campos incorrecta")

            first_name, last_name, document, birthdate, number_str = parts
            number = int(number_str)  # convertir el número a entero

            return {
                "first_name": first_name,
                "last_name": last_name,
                "document": document,
                "birthdate": birthdate,
                "number": number
            }
        except Exception as e:
            logging.error(f"Server: action: parse_bet | result: fail | error: {e}")
            return None