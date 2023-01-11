package tasks

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"edm/pkg/datetime"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alecxcode/sqla"
)

type reqAssignTask struct {
	Task     int `json:"task"`
	Assignee int `json:"assignee"`
}

// AssignTaskAPI updates task assignee
func (tb *TasksBase) AssignTaskAPI(w http.ResponseWriter, r *http.Request) {
	var err error
	var eventAssigneeSet bool
	w.Header().Set("Content-Type", "application/json")
	allow, loggedinID := core.AuthVerifyAPI(w, r, tb.memorydb)
	if !allow {
		return
	}
	var reqObj reqAssignTask
	err = json.NewDecoder(r.Body).Decode(&reqObj)
	if err != nil {
		accs.ThrowServerErrorAPI(w, accs.CurrentFunction()+": decoding json request", loggedinID, reqObj.Task)
		return
	}
	task := Task{ID: reqObj.Task}
	err = task.load(tb.db, tb.dbType)
	iuid := task.GiveAssigneeID()
	if err == sql.ErrNoRows {
		fmt.Fprint(w, `{"error":804,"description":"object not found"}`)
		return
	} else if err != nil {
		accs.ThrowServerErrorAPI(w, accs.CurrentFunction()+": loading task before", loggedinID, reqObj.Task)
		return
	}
	user := team.UnmarshalToProfile(tb.memorydb.GetByID(loggedinID))
	taskProjectCreatorID, _ := task.giveTaskProjectCreatorID(tb.db, tb.dbType)
	if user.UserRole == team.ADMIN || task.GiveCreatorID() == loggedinID || task.GiveAssigneeID() == loggedinID ||
		(task.TaskStatus == CREATED && !accs.IntToBool(reqObj.Assignee)) || taskProjectCreatorID == loggedinID {
		var res int
		var uid int
		if accs.IntToBool(reqObj.Assignee) {
			uid = reqObj.Assignee
		} else {
			uid = user.ID
		}
		res = sqla.UpdateSingleInt(tb.db, tb.dbType, "tasks", "Assignee", uid, reqObj.Task)
		if res > 0 {
			task.Assignee = &team.Profile{ID: uid}
			task.Assignee.Preload(tb.db, tb.dbType)
			eventAssigneeSet = true
			if uid != iuid && task.TaskStatus < DONE {
				t := Task{ID: task.ID,
					StatusSet:  datetime.GetCurrentDateTime(),
					TaskStatus: ASSIGNED}
				ress := t.updateStatus(tb.db, tb.dbType)
				if ress > 0 {
					task.TaskStatus = ASSIGNED
					task.StatusSet = t.StatusSet
				}
			}
		}
		tb.taskEventProcessing(task, nil, nil, eventAssigneeSet, false, false, false, 0, false, 0)
		putTaskIntoMsg(tb.memorydb, task)
		json.NewEncoder(w).Encode(task)
		return
	}
	accs.ThrowAccessDeniedAPI(w, r.URL.Path, loggedinID)
	return
}
