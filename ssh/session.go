package ssh

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/semaphore"
)

type Session struct {
	sess *ssh.Session
	sem  *semaphore.Weighted
}

func NewSession(sess *ssh.Session, sem *semaphore.Weighted) *Session {
	return &Session{sess: sess, sem: sem}
}

func (this Session) Get() *ssh.Session {
	return this.sess
}
func (this Session) Close() {
	this.sem.Release(1)
	this.sess.Close()
}
