package tasks

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alecxcode/sqla"
)

type reqTaskProj struct {
	Task int `json:"task"`
	Proj int `json:"proj"`
}

// UpdateTaskProjectAPI updates task project
func (tb *TasksBase) UpdateTaskProjectAPI(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")
	allow, loggedinID := core.AuthVerifyAPI(w, r, tb.memorydb)
	if !allow {
		return
	}
	var reqObj reqTaskProj
	err = json.NewDecoder(r.Body).Decode(&reqObj)
	if err != nil {
		accs.ThrowServerError(w, accs.CurrentFunction()+": decoding json request", loggedinID, reqObj.Task)
		return
	}
	task := Task{ID: reqObj.Task}
	err = task.load(tb.db, tb.dbType)
	if err == sql.ErrNoRows {
		fmt.Fprint(w, `{"error":804,"description":"object not found"}`)
		return
	} else if err != nil {
		accs.ThrowServerError(w, accs.CurrentFunction()+": loading task", loggedinID, reqObj.Task)
		return
	}
	user := team.UnmarshalToProfile(tb.memorydb.GetByID(loggedinID))
	taskProjectCreatorID, _ := task.giveTaskProjectCreatorID(tb.db, tb.dbType)
	addAllowed := user.UserRole == team.ADMIN || task.GiveCreatorID() == loggedinID || task.GiveAssigneeID() == loggedinID
	detachAllowed := addAllowed || taskProjectCreatorID == loggedinID
	oldProj := task.Project
	var res int
	if accs.IntToBool(reqObj.Proj) && addAllowed {
		res = sqla.UpdateSingleInt(tb.db, tb.dbType, "tasks", "Project", reqObj.Proj, reqObj.Task)
		task.Project = reqObj.Proj
	} else if detachAllowed {
		res = sqla.SetToNullOneByID(tb.db, tb.dbType, "tasks", "Project", reqObj.Task)
		task.Project = 0
	} else {
		accs.ThrowAccessDenied(w, r.URL.Path, loggedinID, reqObj.Task)
		return
	}
	if res > 0 {
		putTaskIntoMsgIfProjChange(tb.memorydb, task, oldProj)
		json.NewEncoder(w).Encode(task)
		return
	}
	accs.ThrowServerError(w, accs.CurrentFunction()+": updating task", loggedinID, reqObj.Task)
	return
}
