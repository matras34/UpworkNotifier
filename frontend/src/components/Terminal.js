import { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { authService, getWebSocketUrl } from '../services/api';
import './Terminal.css';

function TerminalComponent({ connection, onClose }) {
  const terminalRef = useRef(null);
  const [terminal, setTerminal] = useState(null);
  const [fitAddon, setFitAddon] = useState(null);
  const [ws, setWs] = useState(null);
  const [status, setStatus] = useState('connecting');
  const [hostKeyInfo, setHostKeyInfo] = useState(null);

  useEffect(() => {
    // Create terminal
    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
      },
      rows: 24,
      cols: 80,
    });

    const fit = new FitAddon();
    term.loadAddon(fit);

    if (terminalRef.current) {
      term.open(terminalRef.current);
      fit.fit();
    }

    setTerminal(term);
    setFitAddon(fit);

    // Connect WebSocket
    const token = authService.getToken();
    const wsUrl = `${getWebSocketUrl()}?token=${token}`;
    const socket = new WebSocket(wsUrl);

    socket.onopen = () => {
      setStatus('connected');
      term.writeln('Connecting to SSH server...');
      
      // Send connection request
      socket.send(JSON.stringify({
        type: 'connect',
        data: connection,
      }));
    };

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        
        switch (message.type) {
          case 'data':
            term.write(message.data);
            break;
          
          case 'connected':
            term.writeln('\r\n' + message.data + '\r\n');
            break;
          
          case 'error':
            term.writeln('\r\n\x1b[31mError: ' + message.data + '\x1b[0m\r\n');
            setStatus('error');
            break;
          
          case 'hostkey':
            setHostKeyInfo(message.data);
            term.writeln('\r\n\x1b[33mHost key verification required!\x1b[0m');
            term.writeln('Fingerprint: ' + message.data.fingerprint_sha256);
            break;
          
          case 'close':
            term.writeln('\r\n\x1b[33mConnection closed\x1b[0m');
            setStatus('closed');
            break;
          
          default:
            break;
        }
      } catch (err) {
        console.error('Failed to parse message:', err);
      }
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      term.writeln('\r\n\x1b[31mConnection error\x1b[0m');
      setStatus('error');
    };

    socket.onclose = () => {
      term.writeln('\r\n\x1b[33mDisconnected from server\x1b[0m');
      setStatus('closed');
    };

    // Handle terminal input
    term.onData((data) => {
      if (socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
          type: 'data',
          data: data,
        }));
      }
    });

    setWs(socket);

    // Cleanup
    return () => {
      if (socket) {
        socket.close();
      }
      if (term) {
        term.dispose();
      }
    };
  }, [connection]);

  useEffect(() => {
    const handleResize = () => {
      if (fitAddon && ws && ws.readyState === WebSocket.OPEN) {
        fitAddon.fit();
        const { rows, cols } = terminal.buffer.active;
        ws.send(JSON.stringify({
          type: 'resize',
          data: { rows, cols },
        }));
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [fitAddon, ws, terminal]);

  const handleAcceptHostKey = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'hostkey',
        data: { accept: true },
      }));
      setHostKeyInfo(null);
    }
  };

  const handleRejectHostKey = () => {
    if (ws) {
      ws.close();
    }
    setHostKeyInfo(null);
    onClose();
  };

  return (
    <div className="terminal-container">
      <div className="terminal-header">
        <div className="terminal-title">
          {connection.host || 'SSH Terminal'}
          <span className={`status-indicator status-${status}`}>
            {status}
          </span>
        </div>
        <button className="close-button" onClick={onClose}>×</button>
      </div>

      {hostKeyInfo && (
        <div className="hostkey-modal">
          <div className="hostkey-content">
            <h3>Host Key Verification</h3>
            <p>The authenticity of host '{hostKeyInfo.host}' can't be established.</p>
            <p><strong>Algorithm:</strong> {hostKeyInfo.algorithm}</p>
            <p><strong>Fingerprint:</strong></p>
            <code>{hostKeyInfo.fingerprint_sha256}</code>
            <p>Do you want to continue connecting?</p>
            <div className="hostkey-buttons">
              <button onClick={handleAcceptHostKey}>Accept</button>
              <button onClick={handleRejectHostKey}>Reject</button>
            </div>
          </div>
        </div>
      )}

      <div ref={terminalRef} className="terminal"></div>
    </div>
  );
}

export default TerminalComponent;
