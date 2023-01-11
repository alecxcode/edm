package tasks

import (
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"encoding/json"
	"log"
	"strconv"
	"time"
)

func putTaskIntoMsg(mb memdb.ObjectsInMemory, t Task) {
	if accs.IntToBool(t.Project) {
		data, _ := json.Marshal(t)
		go putMsgIntoLoop(mb, t.Project, data)
	}
}

func putTaskIntoMsgIfProjChange(mb memdb.ObjectsInMemory, t Task, oldProj int) {
	if t.Project == oldProj {
		return
	}
	data, _ := json.Marshal(t)
	if accs.IntToBool(oldProj) {
		go putMsgIntoLoop(mb, oldProj, data)
	}
	if accs.IntToBool(t.Project) {
		go putMsgIntoLoop(mb, t.Project, data)
	}
}

func putMsgIntoLoop(mb memdb.ObjectsInMemory, pid int, data []byte) {
	counter := 0
	for {
		if len(mb.GetRaw("project:"+strconv.Itoa(pid))) == 0 {
			mb.SetRaw("project:"+strconv.Itoa(pid), data, 250)
			break
		}
		counter++
		if counter > 600000 {
			log.Println("Possible memory overflow: putMsgIntoLoop could not stop and was closed")
			break
		}
		time.Sleep(time.Duration(60) * time.Millisecond)
	}
}
