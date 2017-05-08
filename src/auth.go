package main

import (
	"fmt"
	"log"
	"net/http"
)

type User struct {
	ID           int64
	Username     string
	Fullname     string
	PasswordHash string
}

func isAuthorized(w http.ResponseWriter, r *http.Request) bool {
	if err := authSession(w, r); err != nil {
		log.Println("failed to authenticate session:", err)
		return false
	}

	return true
}

func authSession(w http.ResponseWriter, r *http.Request) error {
	sid, err := sessionManager.getSessionID(r)
	if err != nil {
		return err
	}
	if len(sid) == 0 {
		return fmt.Errorf("session token not found: %s", sid)
	}

	_, err = sessionManager.ReadSession(sid)
	if err != nil {
		return err
	}

	return nil
}
