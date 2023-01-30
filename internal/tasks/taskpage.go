package tasks

import (
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"edm/pkg/datetime"
	"edm/pkg/memdb"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecxcode/sqla"
)

// TaskPage is passed into template
type TaskPage struct {
	AppTitle       string
	AppVersion     string
	PageTitle      string
	LoggedinID     int
	UserConfig     team.UserConfig
	Task           Task      //payload
	Comments       []Comment //payload
	Participants   []team.Profile
	Message        string
	RemoveAllowed  bool
	Editable       bool
	IamAssignee    bool
	IamParticipant bool
	New            bool
	TaskStatuses   []string
	UserList       []memdb.ObjHasID
}

// TaskHandler is http handler for task page
func (tb *TasksBase) TaskHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, tb.memorydb)
	if !allow {
		return
	}

	if tb.validURLs.FindStringSubmatch(r.URL.Path) == nil {
		accs.ThrowObjectNotFound(w, r)
		return
	}

	var err error

	var Page = TaskPage{
		AppTitle:       tb.text.AppTitle,
		AppVersion:     core.AppVersion,
		LoggedinID:     id,
		Editable:       false,
		IamAssignee:    false,
		IamParticipant: false,
		New:            false,
		TaskStatuses:   tb.text.TaskStatuses,
	}

	TextID := accs.GetTextIDfromURL(r.URL.Path)
	IntID, _ := strconv.Atoi(TextID)
	if TextID == "new" {
		Page.New = true
	} else {
		Page.Task = Task{ID: IntID}
		err = Page.Task.load(tb.db, tb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowObjectNotFound(w, r)
			return
		}
	}

	user := team.UnmarshalToProfile(tb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == team.ADMIN || Page.New || Page.Task.GiveCreatorID() == Page.LoggedinID {
		Page.Editable = true
	}
	Page.RemoveAllowed = tb.cfg.removeAllowed
	if user.UserRole == team.ADMIN {
		Page.RemoveAllowed = true
	}
	Page.IamAssignee = Page.Task.GiveAssigneeID() == Page.LoggedinID
	Page.IamParticipant = accs.SliceContainsInt(Page.Task.Participants, Page.LoggedinID)

	var created int
	var updated int

	var (
		eventAssigneeSet        = false
		eventTaskEdited         = false
		eventTaskStatusChanged  = false
		eventParticipantToAdded = false
		newParticipantID        = 0
		eventNewTaskComment     = false
		newCommentID            = 0
	)

	oldAssigneeID := Page.Task.GiveAssigneeID()
	oldTaskTopic := Page.Task.Topic
	oldProectID := Page.Task.Project

	defaultUploadPath := filepath.Join(tb.cfg.serverRoot, "files", "tasks", TextID)

	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		r.Body = http.MaxBytesReader(w, r.Body, core.MAX_UPLOAD_SIZE+1048576)
	}

	// Create or update code ==========================================================================
	if r.Method == "POST" && (r.FormValue("createButton") != "" || r.FormValue("updateButton") != "") {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "writing task", Page.LoggedinID, IntID)
			return
		}
		t := Task{
			ID:         IntID,
			Created:    datetime.GetCurrentDateTime(),
			PlanStart:  datetime.StringToDateTime(r.FormValue("planStart")),
			PlanDue:    datetime.StringToDateTime(r.FormValue("planDue")),
			Creator:    &team.Profile{ID: Page.LoggedinID},
			Topic:      r.FormValue("topic"),
			Content:    r.FormValue("content"),
			TaskStatus: Page.Task.TaskStatus,
			Project:    accs.StrToInt(r.FormValue("project")),
		}
		if r.FormValue("assignee") != "" && r.FormValue("assignee") != "0" {
			t.Assignee = &team.Profile{ID: accs.StrToInt(r.FormValue("assignee"))}
			if t.Assignee.ID != Page.Task.GiveAssigneeID() && Page.Task.TaskStatus < STUCK {
				t.TaskStatus = ASSIGNED
				t.StatusSet = datetime.GetCurrentDateTime()
				eventAssigneeSet = true
			}
		} else {
			t.Assignee = nil
			if Page.Task.TaskStatus < STUCK && Page.Task.TaskStatus != CREATED {
				t.TaskStatus = CREATED
				t.StatusSet = datetime.GetCurrentDateTime()
				eventTaskStatusChanged = true
			}
		}

		t.FileList, err = core.Uploader(r, defaultUploadPath, "fileList")
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			Page.Message = "uploadError"
		}

		if r.FormValue("createButton") != "" && !strings.Contains(Page.Message, "Error") {
			t.StatusSet = datetime.GetCurrentDateTime()
			t.ID, created = t.Create(tb.db, tb.dbType)
			if created > 0 {
				if eventAssigneeSet {
					t.Creator.Preload(tb.db, tb.dbType)
					t.Assignee.Preload(tb.db, tb.dbType)
					tb.taskEventProcessing(t, nil, nil, eventAssigneeSet, false, false, false, 0, false, 0)
				}
				core.MoveUploadedFilesToFinalDest(defaultUploadPath,
					filepath.Join(tb.cfg.serverRoot, "files", "tasks", strconv.Itoa(t.ID)),
					t.FileList)
				redirAddr := fmt.Sprintf("/tasks/task/%d", t.ID)
				if Page.UserConfig.ReturnAfterCreation {
					redirAddr = "/tasks/?anyparticipants=my"
					if !Page.UserConfig.ShowFinishedTasks {
						redirAddr += "&taskstatuses=0&taskstatuses=1&taskstatuses=2&taskstatuses=3&taskstatuses=6"
					}
				}
				if r.FormValue("projectFrom") != "" {
					redirAddr = "/projs/project/" + r.FormValue("projectFrom")
				}
				putTaskIntoMsg(tb.memorydb, t)
				http.Redirect(w, r, redirAddr+core.IfAddJSON(r), http.StatusSeeOther)
				return
			} else {
				Page.Message = "dataNotWritten"
			}
		}

		if r.FormValue("updateButton") != "" && !strings.Contains(Page.Message, "Error") {
			t.FileList = append(Page.Task.FileList, t.FileList...)
			updated = t.update(tb.db, tb.dbType)
			if updated > 0 {
				Page.Message = "dataWritten"
				Page.Task.load(tb.db, tb.dbType)
				eventTaskEdited = true
				eventTaskStatusChanged = false
			} else {
				Page.Message = "dataNotWritten"
				eventTaskEdited = false
				eventTaskStatusChanged = false
			}
		}

	}

	// Delete files ===========================================
	if r.Method == "POST" && r.FormValue("deleteFiles") != "" {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "writing task", Page.LoggedinID, IntID)
			return
		}
		r.ParseForm()
		filesToRemove := r.Form["filesToRemove"]
		err = core.RemoveUploadedFiles(defaultUploadPath, filesToRemove)
		if err == nil {
			FileList := accs.FilterSliceStrList(Page.Task.FileList, filesToRemove)
			updated = sqla.UpdateSingleJSONListStr(tb.db, tb.dbType, "tasks", "FileList", FileList, IntID)
			if updated > 0 {
				Page.Message = "dataWritten"
				eventTaskEdited = true
				Page.Task.load(tb.db, tb.dbType)
			} else {
				Page.Message = "dataNotWritten"
			}
		} else {
			Page.Message = "removalError"
		}
	}

	// Forward task ========================================================================================
	if r.Method == "POST" && r.FormValue("assigneeForward") != "" && r.FormValue("assigneeForward") != "0" {
		taskProjectCreatorID, _ := Page.Task.giveTaskProjectCreatorID(tb.db, tb.dbType)
		if Page.Editable || Page.IamAssignee || taskProjectCreatorID == Page.LoggedinID ||
			(Page.Task.TaskStatus == CREATED && Page.LoggedinID == accs.StrToInt(r.FormValue("assigneeForward"))) {
			actualAssigneeID := accs.StrToInt(r.FormValue("assigneeForward"))
			if actualAssigneeID != oldAssigneeID {
				res := sqla.UpdateSingleInt(tb.db, tb.dbType, "tasks", "Assignee", actualAssigneeID, IntID)
				if res > 0 {
					Page.Task.Assignee = &team.Profile{ID: actualAssigneeID}
					Page.Task.Assignee.Preload(tb.db, tb.dbType)
					Page.Message = "dataWritten"
					eventAssigneeSet = true
					if actualAssigneeID != oldAssigneeID && Page.Task.TaskStatus < STUCK {
						t := Task{ID: IntID,
							StatusSet:  datetime.GetCurrentDateTime(),
							TaskStatus: ASSIGNED}
						ress := t.updateStatus(tb.db, tb.dbType)
						if ress > 0 {
							Page.Task.TaskStatus = ASSIGNED
							Page.Task.StatusSet = t.StatusSet
						}
					}
				} else {
					Page.Message = "dataNotWritten"
				}
			}
		} else {
			accs.ThrowAccessDenied(w, "forwarding task", Page.LoggedinID, IntID)
			return
		}
	}

	// Create Comments ==========================================
	if r.Method == "POST" && r.FormValue("createComment") != "" {
		if Page.Editable || Page.IamAssignee || Page.IamParticipant {
			c := Comment{
				Created: datetime.GetCurrentDateTime(),
				Creator: &team.Profile{ID: Page.LoggedinID},
				Task:    IntID,
				Content: r.FormValue("content"),
			}
			var caff int
			c.ID, caff = c.Create(tb.db, tb.dbType)
			if caff > 0 {
				Page.Message = "commentWritten"
				eventNewTaskComment = true
				newCommentID = c.ID
			} else {
				Page.Message = "commentNotWritten"
			}
			c.FileList, err = core.Uploader(r, filepath.Join(defaultUploadPath, strconv.Itoa(c.ID)), "fileListComm")
			if err != nil {
				log.Println(accs.CurrentFunction()+":", err)
				Page.Message = "uploadError"
			} else {
				sqla.UpdateSingleJSONListStr(tb.db, tb.dbType, "comments", "FileList", c.FileList, c.ID)
			}
		} else {
			accs.ThrowAccessDenied(w, "writing comment", Page.LoggedinID, IntID)
			return
		}
	}

	// Add participant ===========================================
	if r.Method == "POST" && r.FormValue("participantAdd") != "" {
		if Page.Editable || Page.IamAssignee {
			pID := accs.StrToInt(r.FormValue("participantAdd"))
			if accs.SliceContainsInt(Page.Task.Participants, pID) {
				Page.Message = "participantAlreadyInList"
			} else {
				participants := append(Page.Task.Participants, pID)
				res := sqla.UpdateSingleJSONListInt(tb.db, tb.dbType, "tasks", "Participants", participants, IntID)
				if res > 0 {
					Page.Task.Participants = participants
					Page.Message = "dataWritten"
					eventParticipantToAdded = true
					newParticipantID = pID
				} else {
					Page.Message = "dataNotWritten"
				}
			}
		} else {
			accs.ThrowAccessDenied(w, "adding participant", Page.LoggedinID, IntID)
			return
		}
	}

	// Remove participant ===========================================
	if r.Method == "POST" && r.FormValue("participantRemove") != "" {
		if Page.Editable || Page.IamAssignee {
			pID := accs.StrToInt(r.FormValue("participantRemove"))
			participants := accs.FilterSliceInt(Page.Task.Participants, pID)
			res := sqla.UpdateSingleJSONListInt(tb.db, tb.dbType, "tasks", "participants", participants, IntID)
			if res > 0 {
				Page.Task.Participants = participants
				Page.Message = "dataWritten"
			} else {
				Page.Message = "dataNotWritten"
			}

		} else {
			accs.ThrowAccessDenied(w, "removing participant", Page.LoggedinID, IntID)
			return
		}
	}

	// Other fields code ============================================
	Page.UserList = tb.memorydb.GetObjectArr("UserList")
	Page.IamAssignee = Page.Task.GiveAssigneeID() == Page.LoggedinID
	Page.IamParticipant = accs.SliceContainsInt(Page.Task.Participants, Page.LoggedinID)
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = tb.text.NewTask
	} else {
		Page.PageTitle = tb.text.Task + " #" + TextID
		if Page.Task.Topic != "" {
			Page.PageTitle += ": " + Page.Task.Topic
		}
		//Status INPROGRESS on open
		if Page.IamAssignee && Page.Task.TaskStatus == ASSIGNED {
			t := Task{ID: IntID,
				StatusSet:  datetime.GetCurrentDateTime(),
				TaskStatus: INPROGRESS}
			res := t.updateStatus(tb.db, tb.dbType)
			if res > 0 {
				Page.Task.TaskStatus = INPROGRESS
				Page.Task.StatusSet = t.StatusSet
				eventTaskStatusChanged = true
			}
		}
		Page.Comments, err = Page.Task.loadComments(tb.db, tb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "loading task comments", Page.LoggedinID, Page.Task.ID)
			return
		}
		Page.Participants, err = Page.Task.loadParticipants(tb.db, tb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "loading task participants", Page.LoggedinID, Page.Task.ID)
			return
		}
		//Statuses all
		if Page.Editable || Page.IamAssignee {
			var statusCode int
			if r.Method == "POST" && r.FormValue("taskStatus") != "" {
				statusCode = selectTaskStatus(r.FormValue("taskStatus"))
				if statusCode >= CREATED {
					t := Task{ID: IntID,
						StatusSet:  datetime.GetCurrentDateTime(),
						TaskStatus: statusCode}
					res := t.updateStatus(tb.db, tb.dbType)
					if res > 0 {
						Page.Task.TaskStatus = statusCode
						Page.Task.StatusSet = t.StatusSet
						eventTaskStatusChanged = true
					}
				}
			}
		}
	}

	if Page.Task.GiveAssigneeID() != oldAssigneeID || Page.Task.Topic != oldTaskTopic || eventTaskStatusChanged {
		putTaskIntoMsg(tb.memorydb, Page.Task)
	}
	putTaskIntoMsgIfProjChange(tb.memorydb, Page.Task, oldProectID)

	if core.DEBUG {
		evns := "Events:"
		if eventTaskEdited {
			evns += " eventTaskEdited"
		}
		if eventTaskStatusChanged {
			evns += " eventTaskStatusChanged"
		}
		if eventAssigneeSet {
			evns += " eventAssigneeSet"
		}
		if eventParticipantToAdded {
			evns += " eventParticipantToAdded"
		}
		if eventNewTaskComment {
			evns += " eventNewTaskComment"
		}
		log.Println(evns)
	}

	tb.taskEventProcessing(Page.Task,
		Page.Participants,
		Page.Comments,
		eventAssigneeSet,
		eventTaskEdited,
		eventTaskStatusChanged,
		eventParticipantToAdded,
		newParticipantID,
		eventNewTaskComment,
		newCommentID)

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		return
	}

	// HTML output
	err = tb.templates.ExecuteTemplate(w, "task.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		accs.ThrowServerError(w, "executing task template", Page.LoggedinID, Page.Task.ID)
		return
	}

}
