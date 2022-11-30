package redisdb

import (
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"log"
	"time"
)

// ResetBruteForceCounterAfterMinutes sets BruteForceCounter to 0 after numberOfMinutes specified
func (m *ObjectsInMemory) ResetBruteForceCounterAfterMinutes(numberOfMinutes int) {
	time.Sleep(time.Duration(numberOfMinutes) * time.Minute)
	m.ResetBruteForceCounterImmediately()
}

// ResetBruteForceCounterImmediately sets BruteForceCounter to 0 now
func (m *ObjectsInMemory) ResetBruteForceCounterImmediately() {
	err := m.rdb.Set("BruteForceCounter", 0, 0).Err()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
}

// IncreaseBruteForceCounter increases BruteForceCounter by 1
func (m *ObjectsInMemory) IncreaseBruteForceCounter(ipaddr string, login string) {
	val, err := m.rdb.Get("BruteForceCounter").Result()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
	BruteForceCounter := accs.StrToInt(val)
	BruteForceCounter++
	m.rdb.Set("BruteForceCounter", BruteForceCounter, 0).Err()
	if BruteForceCounter >= memdb.ALLOWED_LOGIN_ATTEMPTS {
		log.Printf("System bruteforce shield activated, last attempt from IP addr: %s, login used: %s", ipaddr, login)
		go m.ResetBruteForceCounterAfterMinutes(memdb.MINUTES_TO_WAIT_BRUTEFORCE_UNLOCK)
	}
}

// VerifyBruteForceCounter checks if BruteForceCounter is more than ALLOWED_LOGIN_ATTEMPTS
func (m *ObjectsInMemory) VerifyBruteForceCounter() (res bool) {
	val, err := m.rdb.Get("BruteForceCounter").Result()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
	BruteForceCounter := accs.StrToInt(val)
	if BruteForceCounter >= memdb.ALLOWED_LOGIN_ATTEMPTS {
		res = false
	} else {
		res = true
	}
	return res
}
