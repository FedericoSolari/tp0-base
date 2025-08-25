import socket
import logging
import signal
import sys



class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
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
            client_sock = self.__accept_new_connection()

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
            # TODO: Modify the receive to avoid short-reads
            msg = client_sock.recv(1024).rstrip().decode('utf-8')
            addr = client_sock.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
            # TODO: Modify the send to avoid short-writes
            client_sock.send("{}\n".format(msg).encode('utf-8'))
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
        # Elimino el socket almacenado
        self._client_skts.remove(client_sock)

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
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