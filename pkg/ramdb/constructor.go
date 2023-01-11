package ramdb

import (
	"edm/pkg/memdb"
	"sync"
)

// ObjectsInMemory holds various data in memory with mutex to prevent data race
type ObjectsInMemory struct {
	sync.RWMutex
	Rarr              map[string][]byte
	Aarr              map[int]memdb.ObjHasID
	Cjar              map[string]int
	Arrs              map[string][]memdb.ObjHasID
	BruteForceCounter int
}

// NewObjectsInMemory is a constructor for the ObjectsInMemory type
func NewObjectsInMemory(arrsNames []string) *ObjectsInMemory {
	objs := ObjectsInMemory{
		Rarr:              make(map[string][]byte),
		Aarr:              make(map[int]memdb.ObjHasID),
		Cjar:              make(map[string]int),
		Arrs:              make(map[string][]memdb.ObjHasID, len(arrsNames)),
		BruteForceCounter: 0,
	}
	for name := range arrsNames {
		objs.Aarr[name] = nil
	}
	return &objs
}

// Close deletes all data from memory
func (m *ObjectsInMemory) Close() {
	m.ClearAll()
	m = nil
}
