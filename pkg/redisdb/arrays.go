package redisdb

import (
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"encoding/json"
	"log"
	"strconv"

	"github.com/go-redis/redis"
)

// Set sets session storage with an object and a cookie
func (m *ObjectsInMemory) Set(cookie string, obj memdb.ObjHasID) {
	err := m.rdb.Set("Cjar:"+cookie, obj.GetID(), 0).Err()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}

	data, _ := json.Marshal(obj)
	err = m.rdb.Set("Aarr:"+strconv.Itoa(obj.GetID()), data, 0).Err()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}

	err = m.rdb.SAdd("Sarr:"+strconv.Itoa(obj.GetID()), cookie).Err()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
}

// Update updates session storage with an object
func (m *ObjectsInMemory) Update(obj memdb.ObjHasID, delete bool) {
	if delete {
		m.DelObject(obj.GetID())
	} else {
		data, _ := json.Marshal(obj)
		err := m.rdb.Set("Aarr:"+strconv.Itoa(obj.GetID()), data, 0).Err()
		if err != nil {
			log.Println(accs.CurrentFunction(), err)
		}
	}
}

// GetByID returns an object by its id
func (m *ObjectsInMemory) GetByID(id int) string {
	elem, err := m.rdb.Get("Aarr:" + strconv.Itoa(id)).Result()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
	return elem
}

// IsObjectInMemory checks if an object is in memory
func (m *ObjectsInMemory) IsObjectInMemory(id int) bool {
	_, err := m.rdb.Get("Aarr:" + strconv.Itoa(id)).Result()
	if err == redis.Nil {
		return false
	} else if err != nil {
		log.Println(accs.CurrentFunction(), err)
		return false
	} else {
		return true
	}
}

// CheckSession returns an object id by cookie value if present
func (m *ObjectsInMemory) CheckSession(cookie string) (result bool, id int) {
	idStr, err := m.rdb.Get("Cjar:" + cookie).Result()
	id = accs.StrToInt(idStr)
	if err == redis.Nil {
		return false, id
	} else if err != nil {
		log.Println(accs.CurrentFunction(), err)
		return false, id
	} else {
		return true, id
	}
}

// DelObject deletes an object by id and its cookie from session storage
func (m *ObjectsInMemory) DelObject(id int) {
	err := m.rdb.Del("Aarr:" + strconv.Itoa(id)).Err()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}

	cookies, err := m.rdb.SMembers("Sarr:" + strconv.Itoa(id)).Result()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}

	for _, cookie := range cookies {
		m.rdb.Del("Cjar:" + cookie).Err()
	}

	err = m.rdb.Del("Sarr:" + strconv.Itoa(id)).Err()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
}

// DelCookie deletes a cookie form session storage
func (m *ObjectsInMemory) DelCookie(cookie string) {
	res, id := m.CheckSession(cookie)
	if res {
		err := m.rdb.Del("Cjar:" + cookie).Err()
		if err != nil {
			log.Println(accs.CurrentFunction(), err)
		}
	}
	cookies, err := m.rdb.SMembers("Sarr:" + strconv.Itoa(id)).Result()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
	if len(cookies) == 0 {
		m.DelObject(id)
	}
}

// ClearAll deletes all objects and cookies from session storage
func (m *ObjectsInMemory) ClearAll() {
	m.rdb.FlushAll()
}

// ObjectArrElem is an element for buildibg lists
type ObjectArrElem struct {
	ID    int
	Value string
}

// GetID is to satisfy ObjHasID interface
func (e ObjectArrElem) GetID() int {
	return e.ID
}

// SetObjectArr sets the specified by name array of objects
func (m *ObjectsInMemory) SetObjectArr(name string, arr []memdb.ObjHasID) {
	data, _ := json.Marshal(arr)
	err := m.rdb.Set(name, data, 0).Err()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
}

// GetObjectArr returns the specified by name array of objects
func (m *ObjectsInMemory) GetObjectArr(name string) []memdb.ObjHasID {
	elems, err := m.rdb.Get(name).Result()
	if err != nil {
		log.Println(accs.CurrentFunction(), err)
	}
	var arr []memdb.ObjHasID
	var arrTmp []ObjectArrElem
	json.Unmarshal([]byte(elems), &arrTmp)
	for _, elem := range arrTmp {
		arr = append(arr, elem)
	}
	return arr
}
