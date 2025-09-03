import socket
import logging
import signal
from common import utils
from common import protocol
from common.connection import Connection


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        # Agrego time out para que no se qude esperando por siempre una conxion
        self._server_socket.settimeout(5.0)
        self._clients = []
        self.shutdown = False

        # Capturo el sigterm para hacer el handeleo 
        signal.signal(signal.SIGTERM, self.handle_sigterm_signal)

    def handle_sigterm_signal(self, signum, frame):

        logging.info("action: handle_sigterm_signal | result: success")
        self.shutdown = True
        
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while self.shutdown == False:
            try:
                conn = self.__accept_new_connection()
                # Almaceno el socket del cliente
                self._clients.append(conn)

                self.__handle_client_connection(conn)
            except socket.timeout:
                # vuelvo a intentar obtener una conexion
                continue

        self.clean_resourses()



    def __handle_client_connection(self, conn):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # recibo todo y decodifico el mensaje
            msg = conn.recv_all()

            bet = protocol.parse_bet_message(msg)
            if bet:
                try:
                    utils.store_bets([bet])
                    logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
                    conn.send_message(protocol.success_message())
                except Exception as e:
                    logging.error(f"action: apuesta_almacenada | result: fail | error: {e}")
                    

        except OSError as e:
            logging.error("action: __handle_client_connection | result: fail | error: {e}")
        finally:
            conn.close()
        # Elimino el socket almacenado
        self._clients.remove(conn)

    # def recv_all(self, client_sock):

    #     buffer = bytearray()
    #     found = False
    #     while found == False:
    #         chunk = client_sock.recv(256)
    #         if not chunk:
    #             logging.error(f'action: Error en la lectura del mensaje')
    #             break
    #         buffer.extend(chunk)
            
    #         # busco el \n que significa el fin segun el protoolo definido
    #         if b'\n' in chunk:
    #             found = True
            
    #     return bytes(buffer)
    
    # def send_all(skt, data: bytes):
    #     """
    #     Envía todos los bytes del mensaje por el socket.
    #     Se asegura que todo se envíe, evitando short-write.
    #     data debe contener el '\n' al final para indicar fin de mensaje.
    #     """
    #     total_sent = 0
    #     total_len = len(data)

    #     while total_sent < total_len:
    #         try:
    #             sent = skt.send(data[total_sent:])
    #             if sent == 0:
    #                 raise RuntimeError("socket connection broken")
    #             total_sent += sent
    #         except OSError as e:
    #             logging.error(f"action: send_all | result: fail | error: {e}")
    #             raise

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        # Then connection created is printed and returned
        """

        logging.info('action: accept_connections | result: in_progress')

        sock, addr = self._server_socket.accept()

        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')

        conn = Connection(sock, addr)
        return conn

    def clean_resourses(self):
        logging.info('Received SIGTERM signal')

        for client_sock in self._clients:
            try:
                logging.info('Closing client connection')
                client_sock.close()
            except OSError as e:
                logging.error(f"Error cerrando client socket: {e}")
        
        self._clients.clear()
        try:
            self._server_socket.close()
        except OSError as e:
           pass
