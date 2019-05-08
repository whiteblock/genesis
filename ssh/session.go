/*
	Copyright 2019 Whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
