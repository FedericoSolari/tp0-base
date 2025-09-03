import socket
import logging
import signal
import sys
from common import utils
from common import protocol
from common.connection import Connection
import os


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
            # Agrego time out para que no se qude esperando por siempre una conxion
        # self._server_socket.settimeout(5.0)
        self._clients = []
        self._clients_agancy = {}
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
        clientes = os.getenv("CLIENTS")
        if clientes is not None:
            clientes = int(clientes)
        else:
            print("No se encontró la variable CLIENTS")

        while self.shutdown == False:
            try:
                conn = self.__accept_new_connection()
                contador += 1

                # Almaceno el Connection del cliente
                self._clients.append(conn)

                self.__handle_client_connection(conn)

                if contador == clientes:
                    # logging.info("RECIBI TODO ARRANCA LA LOTERIA")
                    self.beginLottery()
                    for c in self._clients:
                        #logging.info("action: Cierro cliente")
                        self.__close_client(c)
            except socket.timeout:
                # vuelvo a intentar obtener una conexion
                continue
            except OSError as e:
                logging.error(f"Error en loop principal: {e}")
                break
        self.clean_resourses()


    def __relate_conn_with_agency(self, conn, bets):
        agency_number = bets[0].agency  

        # Solo asigno si no esta
        if conn not in self._clients_agancy:
            self._clients_agancy[agency_number] = conn
            # logging.info(f"action: registrar_agencia | conn: {conn} | agencia: {agency_number}")



    def __handle_bets(self, conn, msg):
        batch_ok = True

        bets = protocol.parse_bet_message(msg)
        self.__relate_conn_with_agency(conn, bets)
        for bet in bets:
            try:
                utils.store_bets([bet])
            except Exception:
                batch_ok = False
                break
                
        if batch_ok:
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            conn.send_message(protocol.success_message())
        else:
            logging.info(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            self.send_message(protocol.error_message())

    def __process_client_messages(self, conn):

        session_active = False
        while True:
            msg = conn.recv_all()
            if not msg:
                break

            if protocol.isStartMessage(msg):
                session_active = True
            elif protocol.isAllBetsDoneMessage(msg):
                session_active = False
                break
            elif session_active:
                self.__handle_bets(conn, msg)
            else:
                logging.error(f"action: __process_client_messages | Message not identificate")
                break

    # Envia el mensaje de finalizacion y cierra el socket del cliente.
    def __close_client(self, conn):
        try:
            # self.send_all(client_sock, protocol.end_message())
            conn.shutdown_wr()
            conn.close()
        except OSError as e:
            logging.error(f"Error cerrando socket cliente: {e}")
        finally:
            if conn in self._clients:
                self._clients.remove(conn)
            if conn in self._clients_agancy:
                del self._clients_agancy[conn]
            
    def __handle_client_connection(self, conn):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
             self.__process_client_messages(conn)
        except OSError as e:
            logging.error("action: __handle_client_connection | result: fail | error: {e}")
        
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

        for client_sock in self._client_skts:
            try:
                logging.info('Closing client connection')
                client_sock.close()
            except OSError as e:
                logging.error(f"Error cerrando client socket: {e}")
        
        self._client_skts.clear()
        self._clients_agancy.clear()
        
        try:
            self._server_socket.close()
        except OSError as e:
           pass
        
        logging.info('Resources closed successfully')

    def beginLottery(self):
        logging.info(f'action: sorteo | result: success')
        
        # notifico que comienza la loteria
        for conn in self._clients:
            conn.send_message(protocol.beginLottery())

        bets = utils.load_bets()
        for b in bets:
            if utils.has_won(b):
                # logging.info(f"Bet agency: {b.agency}")
                winner = self._clients_agancy[b.agency]
                winner.send_message(protocol.parseWinner(b))

        for conn in self._clients:
            conn.send_message(protocol.noMoreWinners())
            conn.send_message(protocol.end_message())