package passwd

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

// GenPasswd creates password hash
func GenPasswd(provided string) string {
	if provided == "" {
		return ""
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(provided), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Generating password hash: ", err)
		return provided
	}
	return string(hash)
}

// ComparePasswd compares password to match hash
func ComparePasswd(stored string, provided string) bool {
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
		log.Println("Comparing password hash: ", err)
		return false
	}
}
