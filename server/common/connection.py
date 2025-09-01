# common/connection.py
import socket
import logging
from typing import Optional

class Connection:
    def __init__(self, sock: socket.socket, addr=None, recv_chunk_size: int = 256):
        self.sock = sock
        self.addr = addr
        self._buffer = bytearray()
        self.recv_chunk_size = recv_chunk_size
        self.closed = False

    # Lee hasta un '\n' (incluido). Devuelve:
    # bytes (incluye '\n') si se leyo una linea,
    # bytes (resto sin '\n') si socket cerro pero habia datos,
    # None si socket cerro y no había datos.
    def recv_all(self) -> Optional[bytes]:
        
        while True:
            idx = self._buffer.find(b'\n')
            if idx != -1:
                line = bytes(self._buffer[:idx+1])
                del self._buffer[:idx+1]
                return line

            try:
                chunk = self.sock.recv(self.recv_chunk_size)
            except OSError as e:
                logging.error(f"recv error from {self.addr}: {e}")
                raise

            if not chunk:
                # EOF
                if self._buffer:
                    data = bytes(self._buffer)
                    self._buffer.clear()
                    return data
                logging.info(f"action: EOF recibido (socket cerrado) addr={self.addr}")
                return None

            self._buffer.extend(chunk)

    def send_all(self, data: bytes) -> None:
        total_sent = 0
        total_len = len(data)
        while total_sent < total_len:
            try:
                sent = self.sock.send(data[total_sent:])
                if sent == 0:
                    raise RuntimeError("socket connection broken")
                total_sent += sent
            except OSError as e:
                logging.error(f"action: send_all | result: fail | error: {e} addr={self.addr}")
                raise

    def send_message(self, msg: bytes) -> None:
        if not msg.endswith(b'\n'):
            msg = msg + b'\n'
        self.send_all(msg)

    def shutdown_wr(self) -> None:
        try:
            self.sock.shutdown(socket.SHUT_WR)
        except OSError:
            pass

    def close(self) -> None:
        try:
            self.sock.close()
        finally:
            self.closed = True

