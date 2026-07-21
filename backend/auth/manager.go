package auth

import (
	"sync"

	"github.com/google/uuid"
)

type AuthManager struct {
	Password        string
	Sessions        map[string]string
	UsernameSession map[string]string
	mu              sync.RWMutex
}

func (m *AuthManager) Login(username string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldToken, ok := m.UsernameSession[username]; ok {
		return oldToken // if he already has a token, reuse it
	}

	id := uuid.New().String()
	m.Sessions[id] = username
	m.UsernameSession[username] = id
	return id
}

func (m *AuthManager) GetSession(token string) string {
	m.mu.RLock()
	username, ok := m.Sessions[token]
	m.mu.RUnlock()

	if !ok {
		return ""
	}

	return username
}

func (m *AuthManager) GetPassword() string {
	return m.Password
}
