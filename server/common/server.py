import socket
import logging
import signal
import sys
from common import utils
from common import protocol
from common.connection import Connection
import os
import threading



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

        self._client_threads = []   
        self._done_lock = threading.Lock()
        self._all_done = threading.Event()
        self._done_clients = 0

        self._agency_lock = threading.Lock()
        self._clients_lock = threading.Lock()
        self._bets_lock = threading.Lock()

        self._expected_clients = int(os.getenv("CLIENTS"))

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
        while self.shutdown == False:
            try:
                conn = self.__accept_new_connection()

                # Almaceno el Connection del cliente
                with self._clients_lock:
                    self._clients.append(conn)

                t = threading.Thread(target=self._client_thread_wrapper, args=(conn,))
                t.start()
                self._client_threads.append(t)

                # if self._done_clients >= self._expected_clients:
                self._all_done.wait()
                logging.info("action: received_all_clients | result: success | waiting for done")

                logging.info("action: join_clients | result: in_progress | ")

                self.join_client_threads()
                logging.info("action: join_clients | result: success | ")

                # logging.info("RECIBI TODO ARRANCA LA LOTERIA")
                self.beginLottery()
                for c in self._clients:
                    #logging.info("action: Cierro cliente")
                    self.__close_client(c)

                # no deberia recibir mas clientes, salgo
                break 
            except socket.timeout:
                # vuelvo a intentar obtener una conexion
                continue
            except OSError as e:
                logging.error(f"Error en loop principal: {e}")
                break
        self.clean_resourses()


    def __relate_conn_with_agency(self, conn, bets):
        agency_number = bets[0].agency  

        with self._agency_lock:
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
                with self._bets_lock:
                    utils.store_bets([bet])
            except Exception:
                batch_ok = False
                break
                
        if batch_ok:
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            conn.send_message(protocol.success_message())
        else:
            logging.info(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            conn.send_message(protocol.error_message())

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
            with self._clients_lock:
                if conn in self._clients:
                    self._clients.remove(conn)
            with self._agency_lock:
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

        with self._clients_lock:
            for client_sock in self._clients:
                try:
                    logging.info('Closing client connection')
                    client_sock.close()
                except OSError as e:
                    logging.error(f"Error cerrando client socket: {e}")
            
            self._clients.clear()
        
        with self._agency_lock:
            self._clients_agancy.clear()
        
        try:
            self._server_socket.close()
        except OSError as e:
           pass
        
        logging.info('Resources closed successfully')

    def beginLottery(self):
        # PONGO LOCK SOLO POR LAS DUDAS
        logging.info(f'action: sorteo | result: success')
        
        # notifico que comienza la loteria
        with self._clients_lock:
            for conn in self._clients:
                conn.send_message(protocol.beginLottery())

        with self._bets_lock:
            bets = utils.load_bets()
    
        with self._agency_lock:
            for b in bets:
                if utils.has_won(b):
                    # logging.info(f"Bet agency: {b.agency}")
                    winner = self._clients_agancy[b.agency]
                    winner.send_message(protocol.parseWinner(b))

        with self._clients_lock:
            for conn in self._clients:
                conn.send_message(protocol.noMoreWinners())
                conn.send_message(protocol.end_message())

    def _client_thread_wrapper(self, conn):
        """
        Wrapper que ejecuta el handler del cliente y marca cuando el cliente terminó.
        Así tenemos control sobre cuándo hacer join() de cada hilo.
        """
        try:
            self.__handle_client_connection(conn)
        except Exception as e:
            logging.error(f"action: _client_thread_wrapper | result: fail | error: {e}")
        finally:
            # actualizo que termino un cliente
            with self._done_lock:
                self._done_clients += 1
                logging.info(f"action: client_finished | result: in_progress | done_clients: {self._done_clients}")
                if self._done_clients >= self._expected_clients:
                    logging.info("action: all_expected_clients_done | result: success")
                    self._all_done.set()

    def join_client_threads(self, timeout=None):
        """
        Hace join() de todos los threads de clientes que estén en self._client_threads.
        Si timeout != None, pasa ese timeout a cada join individual.
        """
        logging.info("action: join_client_threads | result: in_progress")
        for t in self._client_threads:
            try:
                t.join(timeout)
                logging.info(f"action: join_client_threads | thread {t.name} joined")
            except RuntimeError as e:
                logging.error(f"action: join_client_threads | result: fail | error: {e}")
        # limpiar la lista de threads una vez se hizo join
        self._client_threads.clear()
        logging.info("action: join_client_threads | result: success")