from flask import Blueprint, request, current_app
from flask_socketio import emit, disconnect
from flask_jwt_extended import decode_token
from app import socketio, db
from app.models.session import SSHSession
from app.models.connection import SSHConnection
from app.ssh_handler.ssh_client import SSHClient
from app.utils.crypto import Encryptor
from config import Config
import logging
import uuid

bp = Blueprint('ssh', __name__, url_prefix='/api/v1/ssh')

logger = logging.getLogger(__name__)

# Store active SSH sessions
active_sessions = {}

@socketio.on('connect', namespace='/ssh')
def handle_connect():
    """Handle WebSocket connection"""
    try:
        # Get token from query params
        token = request.args.get('token')
        if not token:
            logger.warning("No token provided")
            disconnect()
            return False
        
        # Decode JWT token
        decoded = decode_token(token)
        user_id = decoded['sub']
        
        logger.info(f"User {user_id} connected to SSH WebSocket")
        
        return True
    except Exception as e:
        logger.error(f"Connection error: {e}")
        disconnect()
        return False

@socketio.on('start_session', namespace='/ssh')
def handle_start_session(data):
    """Start SSH session"""
    try:
        # Get user from token
        token = request.args.get('token')
        decoded = decode_token(token)
        user_id = decoded['sub']
        
        logger.info(f"Starting SSH session for user {user_id}")
        
        # Get connection details
        connection_id = data.get('connection_id')
        host = data.get('host')
        port = data.get('port', 22)
        username = data.get('username')
        auth_type = data.get('auth_type')
        password = data.get('password')
        private_key = data.get('private_key')
        
        # If connection_id provided, load from database
        if connection_id:
            connection = SSHConnection.query.filter_by(id=connection_id, user_id=user_id).first()
            if not connection:
                emit('error', {'message': 'Connection not found'})
                return
            
            host = connection.host
            port = connection.port
            username = connection.username
            auth_type = connection.auth_type
            
            # Decrypt credentials
            encryptor = Encryptor(Config.ENCRYPTION_KEY)
            if connection.auth_type == 'password' and connection.encrypted_password:
                password = encryptor.decrypt(
                    connection.encrypted_password,
                    connection.encryption_salt,
                    connection.encryption_nonce
                )
            elif connection.auth_type == 'privatekey' and connection.encrypted_private_key:
                private_key = encryptor.decrypt(
                    connection.encrypted_private_key,
                    connection.encryption_salt,
                    connection.encryption_nonce
                )
        
        # Create SSH session in database
        session_token = str(uuid.uuid4())
        ssh_session = SSHSession(
            user_id=user_id,
            connection_id=connection_id,
            session_token=session_token,
            client_ip=request.remote_addr,
            host=host,
            port=port,
            username=username,
            status='active'
        )
        db.session.add(ssh_session)
        db.session.commit()
        
        # Create SSH client
        ssh_client = SSHClient(
            host=host,
            port=port,
            username=username,
            password=password,
            private_key=private_key,
            session_id=ssh_session.id
        )
        
        # Connect to SSH server
        success, error = ssh_client.connect()
        
        if not success:
            emit('error', {'message': f'SSH connection failed: {error}'})
            ssh_session.status = 'failed'
            ssh_session.disconnect_reason = error
            db.session.commit()
            return
        
        # Store session
        active_sessions[request.sid] = {
            'ssh_client': ssh_client,
            'session_id': ssh_session.id
        }
        
        emit('connected', {'message': 'SSH connection established'})
        
        # Start reading from SSH
        ssh_client.start_read_loop(lambda data: emit('data', {'data': data}))
        
    except Exception as e:
        logger.error(f"Error starting session: {e}")
        emit('error', {'message': str(e)})

@socketio.on('data', namespace='/ssh')
def handle_data(data):
    """Handle terminal input"""
    try:
        session = active_sessions.get(request.sid)
        if not session:
            emit('error', {'message': 'No active session'})
            return
        
        ssh_client = session['ssh_client']
        ssh_client.write(data.get('data', ''))
        
    except Exception as e:
        logger.error(f"Error sending data: {e}")
        emit('error', {'message': str(e)})

@socketio.on('resize', namespace='/ssh')
def handle_resize(data):
    """Handle terminal resize"""
    try:
        session = active_sessions.get(request.sid)
        if not session:
            return
        
        ssh_client = session['ssh_client']
        rows = data.get('rows', 24)
        cols = data.get('cols', 80)
        ssh_client.resize(rows, cols)
        
    except Exception as e:
        logger.error(f"Error resizing terminal: {e}")

@socketio.on('disconnect', namespace='/ssh')
def handle_disconnect():
    """Handle WebSocket disconnect"""
    try:
        session = active_sessions.get(request.sid)
        if session:
            ssh_client = session['ssh_client']
            ssh_client.close()
            
            # Update database
            ssh_session = SSHSession.query.get(session['session_id'])
            if ssh_session:
                ssh_session.status = 'closed'
                ssh_session.disconnect_reason = 'User disconnected'
                db.session.commit()
            
            del active_sessions[request.sid]
            logger.info(f"Session {session['session_id']} closed")
    except Exception as e:
        logger.error(f"Error on disconnect: {e}")
