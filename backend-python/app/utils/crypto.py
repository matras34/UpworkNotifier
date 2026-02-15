from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2
from cryptography.hazmat.backends import default_backend
import base64
import os

class Encryptor:
    """Simple AES-GCM encryption for credentials"""
    
    def __init__(self, master_key):
        if len(master_key) < 32:
            raise ValueError("Master key must be at least 32 characters")
        
        # Derive key from master key
        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,
            salt=b'webssh-salt',
            iterations=100000,
            backend=default_backend()
        )
        self.master_key = kdf.derive(master_key.encode())
    
    def encrypt(self, plaintext):
        """Encrypt data and return encrypted text, salt, and nonce"""
        if not plaintext:
            return None, None, None
        
        # Generate random salt and nonce
        salt = os.urandom(16)
        nonce = os.urandom(12)
        
        # Derive encryption key
        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,
            salt=salt,
            iterations=100000,
            backend=default_backend()
        )
        key = kdf.derive(self.master_key)
        
        # Encrypt
        aesgcm = AESGCM(key)
        ciphertext = aesgcm.encrypt(nonce, plaintext.encode(), None)
        
        return (
            base64.b64encode(ciphertext).decode(),
            base64.b64encode(salt).decode(),
            base64.b64encode(nonce).decode()
        )
    
    def decrypt(self, encrypted_text, salt, nonce):
        """Decrypt data using salt and nonce"""
        if not encrypted_text:
            return None
        
        # Decode from base64
        ciphertext = base64.b64decode(encrypted_text)
        salt_bytes = base64.b64decode(salt)
        nonce_bytes = base64.b64decode(nonce)
        
        # Derive encryption key
        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,
            salt=salt_bytes,
            iterations=100000,
            backend=default_backend()
        )
        key = kdf.derive(self.master_key)
        
        # Decrypt
        aesgcm = AESGCM(key)
        plaintext = aesgcm.decrypt(nonce_bytes, ciphertext, None)
        
        return plaintext.decode()
