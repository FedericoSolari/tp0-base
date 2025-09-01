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
        # self.clean_resourses()
        
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        contador = 0
        clientes = 3
        while self.shutdown == False:
            try:
                client_sock = self.__accept_new_connection()
                contador +=1 # nuevo cliente

                # Almaceno el socket del cliente
                self._client_skts.append(client_sock)
                self.__handle_client_connection(client_sock)
                if contador == clientes:
                    logging.info("RCIBI TODO ARRANCA LA LOTERIA")
                    self.beginLottery()
                    for skt in self._client_skts:
                        logging.info("action: Cierro cliente")
                        self.__close_client_socket(skt)
            except socket.timeout:
                # vuelvo a intentar obtener una conexion
                continue
            except OSError as e:
                logging.error(f"Error en loop principal: {e}")
                break
        self.clean_resourses()


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
            # logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            self.send_all(client_sock, protocol.success_message())
        else:
            logging.info(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            self.send_all(client_sock, protocol.error_message())

    def __process_client_messages(self, client_sock):

        leftover = None
        session_active = False
        while True:
            msg, leftover = self.recv_all(client_sock, leftover)
            if not msg:
                break

            if protocol.isStartMessage(msg):
                session_active = True
            elif protocol.isAllBetsDoneMessage(msg):
                session_active = False
                break
            elif session_active:
                self.__handle_bets(client_sock, msg)
            else:
                logging.error(f"action: __process_client_messages | Message not identificate")
                break

    # Envia el mensaje de finalizacion y cierra el socket del cliente.
    def __close_client_socket(self, client_sock):
        try:
            self.send_all(client_sock, protocol.end_message())
            client_sock.close()
        except OSError as e:
            logging.error(f"Error cerrando socket cliente: {e}")
        finally:
            if client_sock in self._client_skts:
                self._client_skts.remove(client_sock)
            
    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
             self.__process_client_messages(client_sock)
        except OSError as e:
            logging.error("action: __handle_client_connection | result: fail | error: {e}")
        # finally:
            # self.__close_client_socket(client_sock)

    def recv_all(self, client_sock, leftover=None):
        buffer = bytearray()

        # si quedo algo de la llamada anterior, lo usamos de arranque
        if leftover:
            buffer.extend(leftover)

        while True:
            # hay un \n en el buffer?
            idx = buffer.find(b'\n')
            if idx != -1:
                # Devuelvo hasta el \n y lo que sobra ( o nada)
                line = buffer[:idx + 1]
                rest = buffer[idx + 1:] or None
                return bytes(line), rest

            chunk = client_sock.recv(256)
            if not chunk:
                if buffer:
                    return bytes(buffer), None
                logging.info("action: EOF recibido (socket cerrado)")
                return None, None

            buffer.extend(chunk)
    
    def send_all(self,skt, data: bytes):
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
            try:
                logging.info('Closing client connection')
                client_sock.close()
            except OSError as e:
                logging.error(f"Error cerrando client socket: {e}")
        self._client_skts.clear()
        
        try:
            self._server_socket.close()
        except OSError as e:
           pass
        
        logging.info('Resources closed successfully')

    def beginLottery(self):
        logging.info(f'action: sorteo | result: success')
        # notifico que comienza la loteria
        for client in self._client_skts:
            self.send_all(client, protocol.beginLottery())

        bets = utils.load_bets()
        ganadores = 0
        for b in bets:
            if utils.has_won(b):
                ganadores+=1
                self.send_all(self._client_skts[b.agency -1], protocol.parseWinner(b))
                logging.info(f"Bet agemcy{b.agency} dni:{b.document} has won!")

        for client in self._client_skts:
            self.send_all(client, protocol.noMoreWinners())
        # logging.info(f'action: sorteo | result: success')