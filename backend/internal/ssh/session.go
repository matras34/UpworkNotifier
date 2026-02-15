package ssh

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

type SessionStatus string

const (
	StatusActive  SessionStatus = "active"
	StatusClosed  SessionStatus = "closed"
	StatusFailed  SessionStatus = "failed"
	StatusTimeout SessionStatus = "timeout"
)

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	SSHClient        *SSHClient
	SSHSession       *ssh.Session
	Status           SessionStatus
	CreatedAt        time.Time
	LastActivityAt   time.Time
	BytesSent        int64
	BytesReceived    int64
	mu               sync.RWMutex
	cancelFunc       context.CancelFunc
	idleTimeout      time.Duration
	maxDuration      time.Duration
	onDisconnect     func(sessionID uuid.UUID, reason string)
	stdin            io.WriteCloser
	stdout           io.Reader
	stderr           io.Reader
}

type SessionManager struct {
	sessions        map[uuid.UUID]*Session
	mu              sync.RWMutex
	maxSessions     int
	idleTimeout     time.Duration
	maxDuration     time.Duration
	cleanupInterval time.Duration
}

func NewSessionManager(maxSessions int, idleTimeout, maxDuration time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions:        make(map[uuid.UUID]*Session),
		maxSessions:     maxSessions,
		idleTimeout:     idleTimeout,
		maxDuration:     maxDuration,
		cleanupInterval: 1 * time.Minute,
	}

	go sm.cleanupLoop()

	return sm
}

func (sm *SessionManager) CreateSession(ctx context.Context, userID uuid.UUID, sshClient *SSHClient, onDisconnect func(uuid.UUID, string)) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.sessions) >= sm.maxSessions {
		return nil, fmt.Errorf("maximum concurrent sessions reached")
	}

	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := sshSession.RequestPty("xterm-256color", 24, 80, modes); err != nil {
		sshSession.Close()
		return nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	stdin, err := sshSession.StdinPipe()
	if err != nil {
		sshSession.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		sshSession.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := sshSession.StderrPipe()
	if err != nil {
		sshSession.Close()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := sshSession.Shell(); err != nil {
		sshSession.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	sessionCtx, cancel := context.WithCancel(ctx)

	session := &Session{
		ID:             uuid.New(),
		UserID:         userID,
		SSHClient:      sshClient,
		SSHSession:     sshSession,
		Status:         StatusActive,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		cancelFunc:     cancel,
		idleTimeout:    sm.idleTimeout,
		maxDuration:    sm.maxDuration,
		onDisconnect:   onDisconnect,
		stdin:          stdin,
		stdout:         stdout,
		stderr:         stderr,
	}

	sm.sessions[session.ID] = session

	// Start timeout monitoring
	go session.monitorTimeouts(sessionCtx)

	return session, nil
}

func (sm *SessionManager) GetSession(sessionID uuid.UUID) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, ok := sm.sessions[sessionID]
	return session, ok
}

func (sm *SessionManager) CloseSession(sessionID uuid.UUID, reason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}

	session.Close(reason)
	delete(sm.sessions, sessionID)

	return nil
}

func (sm *SessionManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(sm.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, session := range sm.sessions {
		session.mu.RLock()
		status := session.Status
		lastActivity := session.LastActivityAt
		session.mu.RUnlock()

		if status != StatusActive {
			delete(sm.sessions, id)
			continue
		}

		// Check idle timeout
		if now.Sub(lastActivity) > sm.idleTimeout {
			session.Close("idle timeout")
			delete(sm.sessions, id)
		}
	}
}

func (s *Session) UpdateActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActivityAt = time.Now()
}

func (s *Session) Resize(rows, cols int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.SSHSession == nil {
		return fmt.Errorf("session not active")
	}

	return s.SSHSession.WindowChange(rows, cols)
}

func (s *Session) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stdin == nil {
		return 0, fmt.Errorf("stdin not available")
	}

	n, err := s.stdin.Write(data)
	s.BytesSent += int64(n)
	s.LastActivityAt = time.Now()
	return n, err
}

func (s *Session) Read(buf []byte) (int, error) {
	if s.stdout == nil {
		return 0, fmt.Errorf("stdout not available")
	}

	n, err := s.stdout.Read(buf)
	s.mu.Lock()
	s.BytesReceived += int64(n)
	s.LastActivityAt = time.Now()
	s.mu.Unlock()
	return n, err
}

func (s *Session) Close(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status != StatusActive {
		return
	}

	if reason == "idle timeout" {
		s.Status = StatusTimeout
	} else {
		s.Status = StatusClosed
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	if s.SSHSession != nil {
		s.SSHSession.Close()
	}

	if s.SSHClient != nil {
		s.SSHClient.Close()
	}

	if s.onDisconnect != nil {
		go s.onDisconnect(s.ID, reason)
	}
}

func (s *Session) monitorTimeouts(ctx context.Context) {
	idleTicker := time.NewTicker(30 * time.Second)
	defer idleTicker.Stop()

	maxDurationTimer := time.NewTimer(s.maxDuration)
	defer maxDurationTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-maxDurationTimer.C:
			s.Close("maximum duration exceeded")
			return
		case <-idleTicker.C:
			s.mu.RLock()
			lastActivity := s.LastActivityAt
			s.mu.RUnlock()

			if time.Since(lastActivity) > s.idleTimeout {
				s.Close("idle timeout")
				return
			}
		}
	}
}
