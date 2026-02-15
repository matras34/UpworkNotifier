from app import db
from datetime import datetime
import uuid

class SSHConnection(db.Model):
    __tablename__ = 'ssh_connections'
    
    id = db.Column(db.String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = db.Column(db.String(36), db.ForeignKey('users.id'), nullable=False)
    name = db.Column(db.String(255), nullable=False)
    host = db.Column(db.String(255), nullable=False)
    port = db.Column(db.Integer, default=22)
    username = db.Column(db.String(255), nullable=False)
    auth_type = db.Column(db.String(50), nullable=False)  # password or privatekey
    encrypted_password = db.Column(db.Text)
    encrypted_private_key = db.Column(db.Text)
    encryption_salt = db.Column(db.String(255))
    encryption_nonce = db.Column(db.String(255))
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    def to_dict(self):
        """Convert to dictionary (without sensitive data)"""
        return {
            'id': self.id,
            'user_id': self.user_id,
            'name': self.name,
            'host': self.host,
            'port': self.port,
            'username': self.username,
            'auth_type': self.auth_type,
            'created_at': self.created_at.isoformat() if self.created_at else None,
            'updated_at': self.updated_at.isoformat() if self.updated_at else None
        }
