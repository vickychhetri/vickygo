package store

import (
	"database/sql"
	"sync"
	"time"
)

// ---- Note ----

type Note struct {
	Content   string
	ExpiresAt time.Time
}

// ---- Secret (one-time read) ----

type Secret struct {
	Content   string
	ExpiresAt time.Time
}

// ---- Clipboard (PIN-based) ----

type Clipboard struct {
	Content    string
	LastAccess time.Time
}

// ---- DB Session ----

type DBSession struct {
	DB         *sql.DB
	Driver     string
	DSN        string
	DBName     string
	LastAccess time.Time
}

// ---- Memory Store ----

type MemoryStore struct {
	mu         sync.RWMutex
	notes      map[string]Note
	secrets    map[string]Secret
	clipboards map[string]Clipboard
	dbSessions map[string]*DBSession
}

var Global = &MemoryStore{
	notes:      make(map[string]Note),
	secrets:    make(map[string]Secret),
	clipboards: make(map[string]Clipboard),
	dbSessions: make(map[string]*DBSession),
}

// StartGC starts a background goroutine to clean up expired entries every 5 minutes.
func (s *MemoryStore) StartGC() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			s.mu.Lock()

			for id, n := range s.notes {
				if now.After(n.ExpiresAt) {
					delete(s.notes, id)
				}
			}
			for id, sec := range s.secrets {
				if now.After(sec.ExpiresAt) {
					delete(s.secrets, id)
				}
			}
			// Expire clipboards idle for more than 2 hours
			for pin, c := range s.clipboards {
				if now.Sub(c.LastAccess) > 2*time.Hour {
					delete(s.clipboards, pin)
				}
			}
			// Expire DB sessions idle for more than 30 minutes
			for id, sess := range s.dbSessions {
				if now.Sub(sess.LastAccess) > 30*time.Minute {
					sess.DB.Close()
					delete(s.dbSessions, id)
				}
			}

			s.mu.Unlock()
		}
	}()
}

// ---- Note Methods ----

func (s *MemoryStore) SaveNote(id, content string, expiry time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes[id] = Note{Content: content, ExpiresAt: expiry}
}

func (s *MemoryStore) GetNote(id string) (Note, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.notes[id]
	if !ok || time.Now().After(n.ExpiresAt) {
		return Note{}, false
	}
	return n, true
}

// ---- Secret Methods ----

func (s *MemoryStore) SaveSecret(id, content string, expiry time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[id] = Secret{Content: content, ExpiresAt: expiry}
}

// GetAndDestroySecret reads a secret once and immediately deletes it.
func (s *MemoryStore) GetAndDestroySecret(id string) (Secret, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sec, ok := s.secrets[id]
	if !ok || time.Now().After(sec.ExpiresAt) {
		return Secret{}, false
	}
	delete(s.secrets, id)
	return sec, true
}

// ---- Clipboard Methods ----

func (s *MemoryStore) SetClipboard(pin, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clipboards[pin] = Clipboard{Content: content, LastAccess: time.Now()}
}

func (s *MemoryStore) GetClipboard(pin string) (Clipboard, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.clipboards[pin]
	if !ok {
		return Clipboard{}, false
	}
	// Touch last access time
	c.LastAccess = time.Now()
	s.clipboards[pin] = c
	return c, true
}

// ---- DB Session Methods ----

func (s *MemoryStore) SaveDBSession(id string, sess *DBSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess.LastAccess = time.Now()
	s.dbSessions[id] = sess
}

func (s *MemoryStore) GetDBSession(id string) (*DBSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.dbSessions[id]
	if !ok {
		return nil, false
	}
	sess.LastAccess = time.Now()
	return sess, true
}

func (s *MemoryStore) DeleteDBSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.dbSessions[id]; ok {
		sess.DB.Close()
		delete(s.dbSessions, id)
	}
}
