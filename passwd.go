package main

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func genPasswd(provided string) string {
	if provided == "" {
		return ""
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(provided), bcrypt.DefaultCost)
	if err != nil {
		log.Println(currentFunction()+":", err)
		return provided
	}
	return string(hash)
}

func comparePasswd(stored string, provided string) bool {
	if provided == "" && stored == "" {
		return true
	}
	if provided != "" && stored == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(provided))
	if err == nil {
		return true
	} else {
		log.Println(currentFunction()+":", err)
		return false
	}
}
