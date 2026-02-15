package ssh

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type HostKeyInfo struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Algorithm         string `json:"algorithm"`
	PublicKey         string `json:"public_key"`
	FingerprintSHA256 string `json:"fingerprint_sha256"`
	FingerprintMD5    string `json:"fingerprint_md5"`
}

type HostKeyCallback func(hostname string, remote net.Addr, key ssh.PublicKey) error

// HostKeyChecker manages host key verification
type HostKeyChecker struct {
	knownKeys      map[string]ssh.PublicKey
	pendingKeys    map[string]*HostKeyInfo
	verifyCallback func(hostKeyInfo *HostKeyInfo) (bool, error)
}

func NewHostKeyChecker(verifyCallback func(*HostKeyInfo) (bool, error)) *HostKeyChecker {
	return &HostKeyChecker{
		knownKeys:      make(map[string]ssh.PublicKey),
		pendingKeys:    make(map[string]*HostKeyInfo),
		verifyCallback: verifyCallback,
	}
}

func (h *HostKeyChecker) AddKnownKey(host string, port int, key ssh.PublicKey) {
	keyID := fmt.Sprintf("%s:%d:%s", host, port, key.Type())
	h.knownKeys[keyID] = key
}

func (h *HostKeyChecker) GetCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		host, port, err := net.SplitHostPort(hostname)
		if err != nil {
			host = hostname
			port = "22"
		}

		portNum := 22
		if port != "22" {
			fmt.Sscanf(port, "%d", &portNum)
		}

		keyID := fmt.Sprintf("%s:%d:%s", host, portNum, key.Type())

		// Check if this key is already known
		if knownKey, ok := h.knownKeys[keyID]; ok {
			if ssh.FingerprintSHA256(knownKey) == ssh.FingerprintSHA256(key) {
				return nil
			}
			return fmt.Errorf("host key mismatch for %s", hostname)
		}

		// New key - need verification
		hostKeyInfo := &HostKeyInfo{
			Host:              host,
			Port:              portNum,
			Algorithm:         key.Type(),
			PublicKey:         base64.StdEncoding.EncodeToString(key.Marshal()),
			FingerprintSHA256: ssh.FingerprintSHA256(key),
			FingerprintMD5:    fingerprintMD5(key),
		}

		h.pendingKeys[keyID] = hostKeyInfo

		if h.verifyCallback != nil {
			accepted, err := h.verifyCallback(hostKeyInfo)
			if err != nil {
				return err
			}
			if !accepted {
				return fmt.Errorf("host key not accepted by user")
			}

			h.knownKeys[keyID] = key
			delete(h.pendingKeys, keyID)
			return nil
		}

		return fmt.Errorf("host key verification required")
	}
}

func fingerprintMD5(key ssh.PublicKey) string {
	hash := md5.Sum(key.Marshal())
	return hex.EncodeToString(hash[:])
}

func fingerprintSHA256(key ssh.PublicKey) string {
	hash := sha256.Sum256(key.Marshal())
	return base64.RawStdEncoding.EncodeToString(hash[:])
}

type SSHClient struct {
	client  *ssh.Client
	session *ssh.Session
	config  *ssh.ClientConfig
}

func NewSSHClient(host string, port int, username, password string, privateKey *string, timeout time.Duration, hostKeyCallback ssh.HostKeyCallback) (*SSHClient, error) {
	var authMethods []ssh.AuthMethod

	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if privateKey != nil && *privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(*privateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods provided")
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}

	return &SSHClient{
		client: client,
		config: config,
	}, nil
}

func (c *SSHClient) NewSession() (*ssh.Session, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	c.session = session
	return session, nil
}

func (c *SSHClient) Close() error {
	if c.session != nil {
		c.session.Close()
	}
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
