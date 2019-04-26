package ssh

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/semaphore"
)

// Session is a simple wrapper for golang's ssh.Session,
// which decrements a semaphore on destruction.
type Session struct {
	sess *ssh.Session
	sem  *semaphore.Weighted
}

// NewSession creates a new session from a native library ssh session and a semaphore
func NewSession(sess *ssh.Session, sem *semaphore.Weighted) *Session {
	return &Session{sess: sess, sem: sem}
}

// Get returns the internal native library ssh session
func (session Session) Get() *ssh.Session {
	return session.sess
}

// Close closes the internal ssh session and decrements the semaphore
func (session Session) Close() {
	session.sem.Release(1)
	session.sess.Close()
}
