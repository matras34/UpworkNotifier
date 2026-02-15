from flask import Blueprint, request, jsonify, current_app
from flask_jwt_extended import jwt_required, get_jwt_identity
from app import db
from app.models.connection import SSHConnection
from app.utils.crypto import Encryptor
from config import Config

bp = Blueprint('connections', __name__, url_prefix='/api/v1/connections')

@bp.route('', methods=['GET'])
@jwt_required()
def list_connections():
    """List all connections for current user"""
    user_id = get_jwt_identity()
    connections = SSHConnection.query.filter_by(user_id=user_id).all()
    
    return jsonify([conn.to_dict() for conn in connections]), 200

@bp.route('', methods=['POST'])
@jwt_required()
def create_connection():
    """Create a new SSH connection"""
    user_id = get_jwt_identity()
    data = request.get_json()
    
    # Validate
    if not all(k in data for k in ['name', 'host', 'username', 'auth_type']):
        return jsonify({'error': 'Missing required fields'}), 400
    
    if data['auth_type'] not in ['password', 'privatekey']:
        return jsonify({'error': 'Invalid auth_type'}), 400
    
    # Encrypt credentials
    encryptor = Encryptor(Config.ENCRYPTION_KEY)
    encrypted_password = None
    encrypted_key = None
    salt = None
    nonce = None
    
    if data['auth_type'] == 'password' and data.get('password'):
        encrypted_password, salt, nonce = encryptor.encrypt(data['password'])
    elif data['auth_type'] == 'privatekey' and data.get('private_key'):
        encrypted_key, salt, nonce = encryptor.encrypt(data['private_key'])
    
    # Create connection
    connection = SSHConnection(
        user_id=user_id,
        name=data['name'],
        host=data['host'],
        port=data.get('port', 22),
        username=data['username'],
        auth_type=data['auth_type'],
        encrypted_password=encrypted_password,
        encrypted_private_key=encrypted_key,
        encryption_salt=salt,
        encryption_nonce=nonce
    )
    
    db.session.add(connection)
    db.session.commit()
    
    return jsonify({
        'id': connection.id,
        'message': 'Connection saved successfully'
    }), 201

@bp.route('', methods=['DELETE'])
@jwt_required()
def delete_connection():
    """Delete a connection"""
    user_id = get_jwt_identity()
    connection_id = request.args.get('id')
    
    if not connection_id:
        return jsonify({'error': 'Connection ID required'}), 400
    
    connection = SSHConnection.query.filter_by(id=connection_id, user_id=user_id).first()
    
    if not connection:
        return jsonify({'error': 'Connection not found'}), 404
    
    db.session.delete(connection)
    db.session.commit()
    
    return '', 204
