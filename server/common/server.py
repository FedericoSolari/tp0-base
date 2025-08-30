import socket
import logging
import signal
import sys
from common import utils
from common import protocol
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
                client_sock = self.__accept_new_connection()
                # Almaceno el socket del cliente
                self._client_skts.append(client_sock)
                self.__handle_client_connection(client_sock)
            except socket.timeout:
                # vuelvo a intentar obtener una conexion
                continue


    def __handle_bets(self, client_sock, msg):
        batch_ok = True

        bets = protocol.parse_bet_message(msg)
        for bet in bets:
            try:
                utils.store_bets([bet])
            except Exception:
                batch_ok = False
                break
                
        if batch_ok:
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            client_sock.send_all(protocol.success_message())
        else:
            logging.info(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            client_sock.send_all(protocol.error_message())

            
    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            while True:
                msg = self.recv_all(client_sock)
                if not msg:
                    break

                if protocol.startMessage(msg):
                    logging.info("Inicio de recepcion de bets")
                elif protocol.AllBetsDone(msg):
                    logging.info("Fin de recepcion de bets")
                    break
                else:
                    self.__handle_bets(client_sock, msg)


        except OSError as e:
            logging.error("action: __handle_client_connection | result: fail | error: {e}")
        finally:
            client_sock.close()
        # Elimino el socket almacenado
        self._client_skts.remove(client_sock)

    def recv_all(self, client_sock):

        buffer = bytearray()
        found = False
        while found == False:
            chunk = client_sock.recv(256)
            if not chunk:
                logging.error(f'action: Error en la lectura del mensaje')
                break
            buffer.extend(chunk)
            
            # busco el \n que significa el fin segun el protoolo definido
            if b'\n' in chunk:
                found = True
            
        return bytes(buffer)
    
    def send_all(skt, data: bytes):
        """
        Envía todos los bytes del mensaje por el socket.
        Se asegura que todo se envíe, evitando short-write.
        data debe contener el '\n' al final para indicar fin de mensaje.
        """
        total_sent = 0
        total_len = len(data)

        while total_sent < total_len:
            try:
                sent = skt.send(data[total_sent:])
                if sent == 0:
                    raise RuntimeError("socket connection broken")
                total_sent += sent
            except OSError as e:
                logging.error(f"action: send_all | result: fail | error: {e}")
                raise

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        # Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def clean_resourses(self):
        logging.info('Received SIGTERM signal')

        for client_sock in self._client_skts:
            logging.info('Closing client connection')
            client_sock.close()
        
        self._server_socket.close()
        logging.info('Server connection closed')
        
        logging.info('Resources closed successfully')
        sys.exit(0)
