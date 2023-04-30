package memdb

// ALLOWED_LOGIN_ATTEMPTS is a bruteforce shield related constant
const ALLOWED_LOGIN_ATTEMPTS = 100

// MINUTES_TO_WAIT_BRUTEFORCE_UNLOCK is a bruteforce shield related constant
const MINUTES_TO_WAIT_BRUTEFORCE_UNLOCK = 60

// ObjectsInMemory interface defines methods for a package implementing memory storage
type ObjectsInMemory interface {
	SetRaw(key string, data []byte, durationMSec int)
	GetRaw(key string) []byte
	DelRaw(key string)
	ReplaceRawMany(prefix, oldPattern, newPattern string)
	Set(cookie string, obj ObjHasID)
	Update(obj ObjHasID, delete bool)
	GetByID(id int) string
	IsObjectInMemory(id int) bool
	CheckSession(cookie string) (result bool, id int)
	DelObject(id int)
	DelCookie(cookie string)
	ClearAll()
	SetObjectArr(name string, arr []ObjHasID)
	GetObjectArr(name string) []ObjHasID
	ResetBruteForceCounterAfterMinutes(numberOfMinutes int)
	ResetBruteForceCounterImmediately()
	IncreaseBruteForceCounter(ipaddr string, login string)
	VerifyBruteForceCounter() (res bool)
	Close()
}

// ObjHasID is any object with method getID() returning int
type ObjHasID interface {
	GetID() int
}
