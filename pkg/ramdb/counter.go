package ramdb

import (
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
	m.Lock()
	m.BruteForceCounter = 0
	m.Unlock()
}

// IncreaseBruteForceCounter increases BruteForceCounter by 1
func (m *ObjectsInMemory) IncreaseBruteForceCounter(ipaddr string, login string) {
	m.Lock()
	m.BruteForceCounter++
	BruteForceCounter := m.BruteForceCounter
	m.Unlock()
	if BruteForceCounter >= memdb.ALLOWED_LOGIN_ATTEMPTS {
		log.Printf("System bruteforce shield activated, last attempt from IP addr: %s, login used: %s", ipaddr, login)
		go m.ResetBruteForceCounterAfterMinutes(memdb.MINUTES_TO_WAIT_BRUTEFORCE_UNLOCK)
	}
}

// VerifyBruteForceCounter checks if BruteForceCounter is more than ALLOWED_LOGIN_ATTEMPTS
func (m *ObjectsInMemory) VerifyBruteForceCounter() (res bool) {
	m.RLock()
	BruteForceCounter := m.BruteForceCounter
	m.RUnlock()
	if BruteForceCounter >= memdb.ALLOWED_LOGIN_ATTEMPTS {
		res = false
	} else {
		res = true
	}
	return res
}
