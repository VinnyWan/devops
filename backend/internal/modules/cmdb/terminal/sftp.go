package terminal

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPClient wraps an SFTP session for remote file operations.
type SFTPClient struct {
	client *sftp.Client
}

// NewSFTPClient creates a new SFTPClient by opening an SFTP subsystem
// session over the provided SSH connection.
func NewSFTPClient(sshClient *ssh.Client) (*SFTPClient, error) {
	sess, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, err
	}
	return &SFTPClient{client: sess}, nil
}

// Close closes the underlying SFTP session.
func (c *SFTPClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
