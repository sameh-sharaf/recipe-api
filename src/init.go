package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Recipe struct {
	ID         int64
	Name       string
	PrepTime   int
	Difficulty int8
	Vegeterian bool
	Rating     float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

var db *DBManager
var sessionManager *SessionManager
var err error

func init() {
	log.Println("initiate web server..")
	envMSG := CheckEnvVars()
	if len(envMSG) != 0 {
		log.Fatalln("ENVIRONMENT VARIABLE NOT SET:", envMSG)
		return
	}

	// Create DB object
	db, err = InitConnection(os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalln("cannot connect to db", err)
	}

	// Create authentication objects
	maxAge, err := strconv.ParseInt(os.Getenv("COOKIE_MAX_AGE"), 10, 64)
	if err != nil {
		log.Println("invalid COOKIE_MAX_AGE value: Set to default (3600)")
		maxAge = 3600
	}
	cleanUpTime, err := strconv.ParseInt(os.Getenv("CLEANUP_SESSIONS"), 10, 64)
	if err != nil {
		log.Println("invalid CLEANUP_SESSIONS value: Set to default (3600)")
		cleanUpTime = 3600
	}
	sessionManager, err = NewSessionManager(os.Getenv("COOKIE_SID"), maxAge, cleanUpTime)
	if err != nil {
		log.Fatalln("cannot create session manager", err)
	}
}

func CheckEnvVars() string {
	if len(os.Getenv("PORT")) == 0 {
		return "PORT NOT SET"
	}
	if len(os.Getenv("DB_HOST")) == 0 {
		return "DB_HOST NOT SET"
	}
	if len(os.Getenv("DB_NAME")) == 0 {
		return "DB_NAME NOT SET"
	}
	if len(os.Getenv("DB_USER")) == 0 {
		return "DB_USER NOT SET"
	}
	if len(os.Getenv("DB_PASS")) == 0 {
		return "DB_PASS NOT SET"
	}
	if len(os.Getenv("DB_PORT")) == 0 {
		return "DB_PORT NOT SET"
	}
	if len(os.Getenv("COOKIE_SID")) == 0 {
		return "COOKIE_SID NOT SET"
	}

	return ""
}
