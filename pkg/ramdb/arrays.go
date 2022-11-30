package ramdb

import (
	"edm/pkg/memdb"
	"encoding/json"
)

// Set sets session storage with an object and a cookie
func (m *ObjectsInMemory) Set(cookie string, obj memdb.ObjHasID) {
	m.Lock()
	m.Cjar[cookie] = obj.GetID()
	m.Aarr[obj.GetID()] = obj
	m.Unlock()
}

// Update updates session storage with an object
func (m *ObjectsInMemory) Update(obj memdb.ObjHasID, delete bool) {
	if delete {
		m.DelObject(obj.GetID())
	} else {
		m.Lock()
		m.Aarr[obj.GetID()] = obj
		m.Unlock()
	}
}

// GetByID returns an object by its id
func (m *ObjectsInMemory) GetByID(id int) string {
	m.RLock()
	elem := m.Aarr[id]
	m.RUnlock()
	res, _ := json.Marshal(elem)
	return string(res)
}

// IsObjectInMemory checks if an object is in memory
func (m *ObjectsInMemory) IsObjectInMemory(id int) bool {
	res := false
	m.RLock()
	if _, ok := m.Aarr[id]; ok {
		res = true
	}
	m.RUnlock()
	return res
}

// CheckSession returns an object id by cookie value if present
func (m *ObjectsInMemory) CheckSession(cookie string) (result bool, id int) {
	m.RLock()
	id, ok := m.Cjar[cookie]
	m.RUnlock()
	if ok {
		result = true
	} else {
		result = false
	}
	return result, id
}

// DelObject deletes an object by id and its cookie from session storage
func (m *ObjectsInMemory) DelObject(id int) {
	var keysForRemoval []string
	m.RLock()
	for k, cid := range m.Cjar {
		if cid == id {
			keysForRemoval = append(keysForRemoval, k)
		}
	}
	m.RUnlock()
	m.Lock()
	delete(m.Aarr, id)
	for _, k := range keysForRemoval {
		delete(m.Cjar, k)
	}
	m.Unlock()
}

// DelCookie deletes a cookie form session storage
func (m *ObjectsInMemory) DelCookie(cookie string) {
	var idForClearCheck int
	m.Lock()
	idForClearCheck = m.Cjar[cookie]
	delete(m.Cjar, cookie)
	m.Unlock()
	// Below code removes an object with no cookies remaining
	var remove = true
	m.RLock()
	for _, fid := range m.Cjar {
		if fid == idForClearCheck {
			remove = false
			break
		}
	}
	m.RUnlock()
	if remove {
		m.Lock()
		delete(m.Aarr, idForClearCheck)
		m.Unlock()
	}
}

// ClearAll deletes all objects and cookies from session storage
func (m *ObjectsInMemory) ClearAll() {
	m.Lock()
	for k := range m.Aarr {
		delete(m.Aarr, k)
	}
	for k := range m.Cjar {
		delete(m.Cjar, k)
	}
	m.Unlock()
}

// SetObjectArr sets the specified by name array of objects
func (m *ObjectsInMemory) SetObjectArr(name string, arr []memdb.ObjHasID) {
	m.Lock()
	m.Arrs[name] = arr
	m.Unlock()
}

// GetObjectArr returns the specified by name array of objects
func (m *ObjectsInMemory) GetObjectArr(name string) []memdb.ObjHasID {
	m.RLock()
	arr := make([]memdb.ObjHasID, len(m.Arrs[name]), cap(m.Arrs[name]))
	copy(arr, m.Arrs[name])
	m.RUnlock()
	return arr
}
