import { useState, useEffect } from 'react';
import { authService, connectionService } from '../services/api';
import Terminal from './Terminal';
import './Dashboard.css';

function Dashboard({ user, onLogout }) {
  const [connections, setConnections] = useState([]);
  const [showNewConnectionForm, setShowNewConnectionForm] = useState(false);
  const [activeTerminal, setActiveTerminal] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const [formData, setFormData] = useState({
    name: '',
    host: '',
    port: 22,
    username: '',
    auth_type: 'password',
    password: '',
    private_key: '',
  });

  useEffect(() => {
    loadConnections();
  }, []);

  const loadConnections = async () => {
    try {
      const data = await connectionService.listConnections();
      setConnections(data || []);
    } catch (err) {
      setError('Failed to load connections');
    }
  };

  const handleSaveConnection = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await connectionService.saveConnection(formData);
      setShowNewConnectionForm(false);
      setFormData({
        name: '',
        host: '',
        port: 22,
        username: '',
        auth_type: 'password',
        password: '',
        private_key: '',
      });
      await loadConnections();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteConnection = async (id) => {
    if (!window.confirm('Are you sure you want to delete this connection?')) {
      return;
    }

    try {
      await connectionService.deleteConnection(id);
      await loadConnections();
    } catch (err) {
      setError('Failed to delete connection');
    }
  };

  const handleConnect = (connection) => {
    setActiveTerminal({
      connection_id: connection.id,
      host: connection.host,
      port: connection.port,
      username: connection.username,
    });
  };

  const handleQuickConnect = () => {
    setActiveTerminal({
      host: formData.host,
      port: formData.port || 22,
      username: formData.username,
      auth_type: formData.auth_type,
      password: formData.password || undefined,
      private_key: formData.private_key || undefined,
    });
  };

  if (activeTerminal) {
    return (
      <Terminal
        connection={activeTerminal}
        onClose={() => setActiveTerminal(null)}
      />
    );
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>Web SSH</h1>
        <div className="user-info">
          <span>{user.email}</span>
          <button onClick={onLogout}>Logout</button>
        </div>
      </div>

      <div className="dashboard-content">
        {error && <div className="error-message">{error}</div>}

        <div className="connections-section">
          <div className="section-header">
            <h2>Saved Connections</h2>
            <button onClick={() => setShowNewConnectionForm(!showNewConnectionForm)}>
              {showNewConnectionForm ? 'Cancel' : '+ New Connection'}
            </button>
          </div>

          {showNewConnectionForm && (
            <div className="connection-form">
              <h3>New Connection</h3>
              <form onSubmit={handleSaveConnection}>
                <div className="form-row">
                  <div className="form-group">
                    <label>Name *</label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      required
                      placeholder="My Server"
                    />
                  </div>
                  <div className="form-group">
                    <label>Host *</label>
                    <input
                      type="text"
                      value={formData.host}
                      onChange={(e) => setFormData({ ...formData, host: e.target.value })}
                      required
                      placeholder="192.168.1.100"
                    />
                  </div>
                  <div className="form-group">
                    <label>Port *</label>
                    <input
                      type="number"
                      value={formData.port}
                      onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) })}
                      required
                      min="1"
                      max="65535"
                    />
                  </div>
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label>Username *</label>
                    <input
                      type="text"
                      value={formData.username}
                      onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                      required
                      placeholder="root"
                    />
                  </div>
                  <div className="form-group">
                    <label>Auth Type *</label>
                    <select
                      value={formData.auth_type}
                      onChange={(e) => setFormData({ ...formData, auth_type: e.target.value })}
                    >
                      <option value="password">Password</option>
                      <option value="privatekey">Private Key</option>
                    </select>
                  </div>
                </div>

                {formData.auth_type === 'password' ? (
                  <div className="form-group">
                    <label>Password *</label>
                    <input
                      type="password"
                      value={formData.password}
                      onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                      required
                    />
                  </div>
                ) : (
                  <div className="form-group">
                    <label>Private Key *</label>
                    <textarea
                      value={formData.private_key}
                      onChange={(e) => setFormData({ ...formData, private_key: e.target.value })}
                      required
                      rows="6"
                      placeholder="-----BEGIN RSA PRIVATE KEY-----"
                    />
                  </div>
                )}

                <div className="form-actions">
                  <button type="submit" disabled={loading}>
                    {loading ? 'Saving...' : 'Save Connection'}
                  </button>
                  <button type="button" onClick={handleQuickConnect}>
                    Quick Connect (Don't Save)
                  </button>
                </div>
              </form>
            </div>
          )}

          <div className="connections-list">
            {connections.length === 0 ? (
              <p className="no-connections">No saved connections. Create one to get started!</p>
            ) : (
              connections.map((conn) => (
                <div key={conn.id} className="connection-card">
                  <div className="connection-info">
                    <h3>{conn.name}</h3>
                    <p>{conn.username}@{conn.host}:{conn.port}</p>
                    <span className="auth-badge">{conn.auth_type}</span>
                  </div>
                  <div className="connection-actions">
                    <button onClick={() => handleConnect(conn)}>Connect</button>
                    <button onClick={() => handleDeleteConnection(conn.id)} className="delete-btn">
                      Delete
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default Dashboard;
