package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	SessionKey string
	User       User
	LoginTime  time.Time
}

type SessionManager struct {
	cookieName  string
	lock        sync.Mutex
	maxLifeTime int64
	cleanUpTime int64
	sessions    map[string]Session
}

func NewSessionManager(cookieName string, maxLifeTime, cleanUpTime int64) (*SessionManager, error) {
	sessionManager := &SessionManager{
		cookieName:  cookieName,
		maxLifeTime: maxLifeTime,
		cleanUpTime: cleanUpTime,
		sessions:    make(map[string]Session),
	}

	// Clean up expired sessions every 1 hour
	go func() {
		for {
			sessionManager.CleanupSessions(maxLifeTime)
			time.Sleep(time.Duration(sessionManager.cleanUpTime) * time.Second)
		}
	}()

	return sessionManager, nil
}

func (sessionManager *SessionManager) sessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (sessionManager *SessionManager) SessionStart(w http.ResponseWriter, r *http.Request, user User) (Session, error) {
	if sessionManager == nil {
		log.Println("session manager is not initialized")
		return Session{}, nil
	}

	sessionManager.lock.Lock()
	defer sessionManager.lock.Unlock()

	sid, err := sessionManager.getSessionID(r)
	if err != nil || len(sid) == 0 {
		return sessionManager.setCookie(w, r, user)
	}

	session, err := sessionManager.ReadSession(sid)
	if err != nil || len(session.SessionKey) == 0 {
		return sessionManager.setCookie(w, r, user)
	}

	return session, nil
}

func (sessionManager *SessionManager) InitSession(sid string, userID int64) (Session, error) {
	session := Session{
		SessionKey: sid,
		User: User{
			ID: userID,
		},
		LoginTime: time.Now().UTC(),
	}

	err := db.InsertUserSession(session)
	if err != nil {
		return session, err
	}

	sessionManager.sessions[session.SessionKey] = session

	return session, nil
}

func (sessionManager *SessionManager) ReadSession(sid string) (Session, error) {
	if session, ok := sessionManager.sessions[sid]; ok {
		if session.LoginTime.Add(time.Duration(sessionManager.maxLifeTime) * time.Second).After(time.Now()) {
			return session, nil
		}
	}

	session, err := db.GetUserActiveSessions(sid, sessionManager.maxLifeTime)
	if err != nil {
		return session, err
	}

	return session, nil
}

func (sessionManager *SessionManager) DestroySession(sid string) error {
	sessionManager.lock.Lock()
	defer sessionManager.lock.Unlock()

	delete(sessionManager.sessions, sid)
	return db.DeleteUserSessionByID(sid)
}

func (sessionManager *SessionManager) CleanupSessions(maxLifeTime int64) {
	log.Println("Clean up expired session tokens")
	t := time.Now().UTC().Add(-1 * time.Duration(maxLifeTime) * time.Second)
	db.DeleteExpiredUserSessions(t)
}

func (sessionManager *SessionManager) setCookie(w http.ResponseWriter, r *http.Request, user User) (Session, error) {
	sid := sessionManager.sessionID()
	session, err := sessionManager.InitSession(sid, user.ID)
	if err != nil {
		return session, err
	}

	cookie := http.Cookie{
		Name:     sessionManager.cookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(sessionManager.maxLifeTime),
	}

	http.SetCookie(w, &cookie)
	return session, nil
}

func (sessionManager *SessionManager) getSessionID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionManager.cookieName)
	if err != nil {
		return "", err
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	return sid, nil
}

func (sessionManager *SessionManager) EncryptPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
