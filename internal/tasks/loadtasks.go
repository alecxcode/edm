package tasks

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"encoding/json"
	"net/http"

	"github.com/alecxcode/sqla"
)

type reqByID struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// LoadTasksAPI loads tasks with json request and responce
func (tb *TasksBase) LoadTasksAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	allow, uid := core.AuthVerifyAPI(w, r, tb.memorydb)
	if !allow {
		return
	}
	var reqObj reqByID
	err := json.NewDecoder(r.Body).Decode(&reqObj)
	if err != nil {
		accs.ThrowServerErrorAPI(w, accs.CurrentFunction()+": decoding json request", uid, reqObj.ID)
		return
	}
	taskList, err := reqObj.loadTasksByProj(tb.db, tb.dbType)
	if err != nil {
		accs.ThrowServerErrorAPI(w, accs.CurrentFunction()+": loading task list", uid, reqObj.ID)
		return
	}
	json.NewEncoder(w).Encode(taskList)
	return
}

func (p reqByID) loadTasksByProj(db *sql.DB, DBType byte) (TaskList []Task, err error) {
	rows, err := db.Query(`SELECT t.ID, t.Creator, t.Assignee, t.Topic, t.TaskStatus,
c.ID, c.FirstName, c.Surname,  c.JobTitle,
a.ID, a.FirstName, a.Surname,  a.JobTitle
FROM tasks t
LEFT JOIN profiles c ON c.ID = t.Creator
LEFT JOIN profiles a ON a.ID = t.Assignee
WHERE t.Project = `+sqla.MakeParam(DBType, 1)+` ORDER BY t.ID ASC`, p.ID)
	if err != nil {
		return TaskList, err
	}
	defer rows.Close()

	var Creator sql.NullInt64
	var Assignee sql.NullInt64
	var Topic sql.NullString
	var TaskStatus sql.NullInt64

	var CreatorID sql.NullInt64
	var CreatorFirstName sql.NullString
	var CreatorSurname sql.NullString
	var CreatorJobTitle sql.NullString

	var AssigneeID sql.NullInt64
	var AssigneeFirstName sql.NullString
	var AssigneeSurname sql.NullString
	var AssigneeJobTitle sql.NullString

	for rows.Next() {
		var t Task
		err = rows.Scan(&t.ID, &Creator, &Assignee, &Topic, &TaskStatus,
			&CreatorID, &CreatorFirstName, &CreatorSurname, &CreatorJobTitle,
			&AssigneeID, &AssigneeFirstName, &AssigneeSurname, &AssigneeJobTitle)
		if err != nil {
			return TaskList, err
		}
		if Creator.Valid {
			t.Creator = &team.Profile{
				ID:        int(CreatorID.Int64),
				FirstName: CreatorFirstName.String,
				Surname:   CreatorSurname.String,
				JobTitle:  CreatorJobTitle.String,
			}
		}
		if Assignee.Valid {
			t.Assignee = &team.Profile{
				ID:        int(AssigneeID.Int64),
				FirstName: AssigneeFirstName.String,
				Surname:   AssigneeSurname.String,
				JobTitle:  AssigneeJobTitle.String,
			}
		}
		t.Topic = Topic.String
		t.TaskStatus = int(TaskStatus.Int64)
		TaskList = append(TaskList, t)
	}
	return TaskList, nil
}
