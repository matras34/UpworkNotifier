import paramiko
import threading
import logging
from io import StringIO

logger = logging.getLogger(__name__)

class SSHClient:
    """SSH client wrapper using paramiko"""
    
    def __init__(self, host, port, username, password=None, private_key=None, session_id=None):
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.private_key = private_key
        self.session_id = session_id
        
        self.client = None
        self.channel = None
        self.read_thread = None
        self.running = False
    
    def connect(self):
        """Connect to SSH server"""
        try:
            self.client = paramiko.SSHClient()
            self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
            
            # Prepare authentication
            connect_kwargs = {
                'hostname': self.host,
                'port': self.port,
                'username': self.username,
                'timeout': 10
            }
            
            if self.password:
                connect_kwargs['password'] = self.password
            elif self.private_key:
                try:
                    key_file = StringIO(self.private_key)
                    pkey = paramiko.RSAKey.from_private_key(key_file)
                    connect_kwargs['pkey'] = pkey
                except Exception:
                    # Try other key types
                    try:
                        key_file = StringIO(self.private_key)
                        pkey = paramiko.Ed25519Key.from_private_key(key_file)
                        connect_kwargs['pkey'] = pkey
                    except Exception as e:
                        return False, f"Invalid private key: {str(e)}"
            
            # Connect
            self.client.connect(**connect_kwargs)
            
            # Open shell channel
            self.channel = self.client.invoke_shell(term='xterm-256color', width=80, height=24)
            self.running = True
            
            logger.info(f"SSH connection established to {self.host}:{self.port}")
            return True, None
            
        except paramiko.AuthenticationException:
            return False, "Authentication failed"
        except paramiko.SSHException as e:
            return False, f"SSH error: {str(e)}"
        except Exception as e:
            return False, f"Connection error: {str(e)}"
    
    def start_read_loop(self, callback):
        """Start reading from SSH channel in background thread"""
        def read_loop():
            try:
                while self.running and self.channel:
                    if self.channel.recv_ready():
                        data = self.channel.recv(4096)
                        if data:
                            callback(data.decode('utf-8', errors='replace'))
            except Exception as e:
                logger.error(f"Error in read loop: {e}")
        
        self.read_thread = threading.Thread(target=read_loop, daemon=True)
        self.read_thread.start()
    
    def write(self, data):
        """Write data to SSH channel"""
        if self.channel and self.running:
            try:
                self.channel.send(data)
            except Exception as e:
                logger.error(f"Error writing to SSH: {e}")
    
    def resize(self, rows, cols):
        """Resize terminal"""
        if self.channel and self.running:
            try:
                self.channel.resize_pty(width=cols, height=rows)
            except Exception as e:
                logger.error(f"Error resizing terminal: {e}")
    
    def close(self):
        """Close SSH connection"""
        self.running = False
        
        if self.channel:
            try:
                self.channel.close()
            except Exception:
                pass
        
        if self.client:
            try:
                self.client.close()
            except Exception:
                pass
        
        logger.info(f"SSH connection to {self.host}:{self.port} closed")
