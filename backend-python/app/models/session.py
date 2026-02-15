from app import db
from datetime import datetime
import uuid

class SSHSession(db.Model):
    __tablename__ = 'ssh_sessions'
    
    id = db.Column(db.String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = db.Column(db.String(36), db.ForeignKey('users.id'), nullable=False)
    connection_id = db.Column(db.String(36), db.ForeignKey('ssh_connections.id'))
    session_token = db.Column(db.String(255), unique=True, nullable=False)
    client_ip = db.Column(db.String(45), nullable=False)
    host = db.Column(db.String(255), nullable=False)
    port = db.Column(db.Integer, nullable=False)
    username = db.Column(db.String(255), nullable=False)
    status = db.Column(db.String(50), nullable=False, default='active')  # active, closed, failed, timeout
    started_at = db.Column(db.DateTime, default=datetime.utcnow)
    last_activity_at = db.Column(db.DateTime, default=datetime.utcnow)
    ended_at = db.Column(db.DateTime)
    disconnect_reason = db.Column(db.String(255))
    bytes_sent = db.Column(db.BigInteger, default=0)
    bytes_received = db.Column(db.BigInteger, default=0)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    
    def to_dict(self):
        """Convert to dictionary"""
        return {
            'id': self.id,
            'user_id': self.user_id,
            'connection_id': self.connection_id,
            'host': self.host,
            'port': self.port,
            'username': self.username,
            'status': self.status,
            'started_at': self.started_at.isoformat() if self.started_at else None,
            'ended_at': self.ended_at.isoformat() if self.ended_at else None
        }
