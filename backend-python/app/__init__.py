from flask import Flask
from flask_socketio import SocketIO
from flask_jwt_extended import JWTManager
from flask_cors import CORS
from flask_sqlalchemy import SQLAlchemy
import redis
import logging

# Initialize extensions
db = SQLAlchemy()
socketio = SocketIO()
jwt = JWTManager()

def create_app():
    """Application factory"""
    app = Flask(__name__)
    
    # Load configuration
    from config import Config
    app.config.from_object(Config)
    
    # Setup logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # Initialize extensions
    db.init_app(app)
    jwt.init_app(app)
    
    # CORS
    CORS(app, resources={r"/api/*": {"origins": Config.CORS_ORIGINS}})
    
    # SocketIO
    socketio.init_app(app, cors_allowed_origins=Config.CORS_ORIGINS, async_mode='eventlet')
    
    # Redis
    app.redis = redis.Redis(
        host=Config.REDIS_HOST,
        port=Config.REDIS_PORT,
        password=Config.REDIS_PASSWORD if Config.REDIS_PASSWORD else None,
        db=Config.REDIS_DB,
        decode_responses=True
    )
    
    # Create tables
    with app.app_context():
        from app.models import user, connection, session
        db.create_all()
        
        # Create default admin user if not exists
        from app.models.user import User
        if not User.query.filter_by(email='admin@example.com').first():
            admin = User(
                email='admin@example.com',
                role='admin'
            )
            admin.set_password('admin123')
            db.session.add(admin)
            db.session.commit()
            logging.info("Created default admin user: admin@example.com / admin123")
    
    # Register blueprints
    from app.api import auth, connections, ssh
    app.register_blueprint(auth.bp)
    app.register_blueprint(connections.bp)
    app.register_blueprint(ssh.bp)
    
    @app.route('/health')
    def health():
        return {'status': 'ok'}, 200
    
    return app
