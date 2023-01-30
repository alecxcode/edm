package tasks

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/mail"
	"edm/internal/team"
	"edm/pkg/accs"
	"edm/pkg/datetime"
	"edm/pkg/memdb"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/alecxcode/sqla"
)

// TasksPage is passed into template
type TasksPage struct {
	AppTitle      string
	AppVersion    string
	PageTitle     string
	LoggedinID    int
	LoggedinAdmin bool
	Message       string
	RemoveAllowed bool
	UserConfig    team.UserConfig
	Tasks         []Task //payload
	SortedBy      string
	SortedHow     int
	Filters       sqla.Filter
	PageNumber    int
	FilteredNum   int
	RemovedNum    int
	UpdatedNum    int
	TaskStatuses  []string
	UserList      []memdb.ObjHasID
}

// TasksHandler is http handler for docs page
func (tb *TasksBase) TasksHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, tb.memorydb)
	if !allow {
		return
	}

	if tb.validURLs.FindStringSubmatch(r.URL.Path) == nil {
		accs.ThrowObjectNotFound(w, r)
		return
	}

	var err error

	var Page = TasksPage{
		AppTitle:     tb.text.AppTitle,
		AppVersion:   core.AppVersion,
		PageTitle:    tb.text.TasksPageTitle,
		TaskStatuses: tb.text.TaskStatuses,
		SortedBy:     "ID",
		SortedHow:    0, // 0 - DESC, 1 - ASC
		Filters: sqla.Filter{
			ClassFilter: []sqla.ClassFilter{
				{Name: "taskstatuses", Column: "TaskStatus"},
				{Name: "creators", Selector: "userSelector", Column: "creator.ID"},
				{Name: "assignees", Selector: "userSelector", Column: "assignee.ID"},
				{Name: "participants", Selector: "userSelector", InJSON: true, Column: "Participants"},
			},
			ClassFilterOR: []sqla.ClassFilter{
				{Name: "creatorsOrAssignees", Selector: "userSelector", Column: "creator.ID"},
				{Name: "creatorsOrAssignees", Selector: "userSelector", Column: "assignee.ID"},
				{Name: "anyparticipants", Selector: "userSelector", Column: "creator.ID"},
				{Name: "anyparticipants", Selector: "userSelector", Column: "assignee.ID"},
				{Name: "anyparticipants", Selector: "userSelector", InJSON: true, Column: "Participants"},
			},
			DateFilter: []sqla.DateFilter{
				{Name: "createdDates", Column: "tasks.Created"},
				{Name: "planStartDates", Column: "PlanStart"},
				{Name: "planDueDates", Column: "PlanDue"},
				{Name: "statusSetDates", Column: "StatusSet"},
			},
			TextFilterName:    "searchText",
			TextFilterColumns: []string{"Topic", "tasks.Content", "tasks.FileList", "comments.Content", "comments.FileList"},
		},
		PageNumber: 1,
		LoggedinID: id,
	}

	Page.RemoveAllowed = tb.cfg.removeAllowed
	user := team.UnmarshalToProfile(tb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == team.ADMIN {
		Page.LoggedinAdmin = true
		Page.RemoveAllowed = true
	}

	if r.FormValue("from") == "core" && !Page.UserConfig.ShowFinishedTasks {
		http.Redirect(w, r, "/tasks/?anyparticipants=my&taskstatuses=0&taskstatuses=1&taskstatuses=2&taskstatuses=3&taskstatuses=6", http.StatusSeeOther)
		return
	}

	Page.Filters.GetFilterFromForm(r,
		datetime.ConvDateStrToInt64, datetime.ConvDateTimeStrToInt64,
		map[string]int{"my": Page.LoggedinID})

	// Parsing other fields
	if r.FormValue("sortedBy") != "" {
		Page.SortedBy = r.FormValue("sortedBy")
	}
	var SQLSortedBy string
	switch Page.SortedBy {
	case "ID":
		SQLSortedBy = "tasks.ID"
	case "Created":
		SQLSortedBy = "tasks.Created"
	case "PlanStart":
		SQLSortedBy = "PlanStart"
	case "PlanDue":
		SQLSortedBy = "PlanDue"
	case "StatusSet":
		SQLSortedBy = "StatusSet"
	case "Topic":
		SQLSortedBy = "Topic, tasks.Content"
	default:
		SQLSortedBy = "tasks.ID"
	}
	if r.FormValue("sortedHow") != "" {
		Page.SortedHow, _ = strconv.Atoi(r.FormValue("sortedHow"))
	}

	if r.FormValue("elemsOnPage") != "" {
		Page.UserConfig.ElemsOnPage, _ = strconv.Atoi(r.FormValue("elemsOnPage"))
		if r.FormValue("elemsOnPageChanged") == "true" {
			p := team.Profile{ID: Page.LoggedinID, UserConfig: Page.UserConfig}
			updated := p.UpdateConfig(tb.db, tb.dbType)
			if updated > 0 {
				team.MemoryUpdateProfile(tb.db, tb.dbType, tb.memorydb, p.ID)
			}
		}
	}
	if r.FormValue("pageNumber") != "" {
		Page.PageNumber, _ = strconv.Atoi(r.FormValue("pageNumber"))
	}

	// Processing status change
	if r.Method == "POST" && r.FormValue("taskStatus") != "" {
		var statusCode int
		statusCode = selectTaskStatus(r.FormValue("taskStatus"))
		if statusCode >= CREATED {
			r.ParseForm()
			ids := []int{}
			for _, v := range r.Form["ids"] {
				id, _ := strconv.Atoi(v)
				ids = append(ids, id)
			}
			if len(ids) > 0 {
				allowedToUpdateStatus := checkStatusModifyPermissions(tb.db, tb.dbType, "tasks", Page.LoggedinID, Page.LoggedinAdmin, ids)
				if allowedToUpdateStatus {
					squpd := `SELECT tasks.ID, tasks.Created, PlanStart, PlanDue, StatusSet, tasks.Creator, Assignee, Participants, Topic, tasks.Content, TaskStatus, tasks.FileList,
creator.ID, creator.FirstName, creator.Surname, creator.JobTitle, creator.Contacts, creator.Userlock,
assignee.ID, assignee.FirstName, assignee.Surname, assignee.JobTitle, assignee.Contacts, assignee.UserLock
FROM tasks
LEFT JOIN profiles creator ON creator.ID = tasks.Creator
LEFT JOIN profiles assignee ON assignee.ID = Assignee
WHERE tasks.TaskStatus <> ` + sqla.MakeParam(tb.dbType, 1) + " "
					var argsupd, argsadd []interface{}
					_, squpd, argsadd = sqla.BuildSQLIN(tb.dbType, squpd, 1, "tasks.ID", ids)
					argsupd = append(argsupd, statusCode)
					argsupd = append(argsupd, argsadd...)
					sqcountupd := "SELECT COUNT(*) FROM tasks WHERE TaskStatus <> " + sqla.MakeParam(tb.dbType, 1) + " "
					_, sqcountupd, _ = sqla.BuildSQLIN(tb.dbType, sqcountupd, 1, "tasks.ID", ids)
					if core.DEBUG {
						log.Println("Selecting tasks to update status:", squpd, argsupd, "\n", sqcountupd)
					}
					tasks, numtoupd, err := loadTasks(tb.db, squpd, sqcountupd, argsupd)
					if err != nil {
						log.Println(accs.CurrentFunction()+":", err)
					}
					if numtoupd > 0 {
						idstoupd := make([]int, numtoupd, numtoupd)
						for i := 0; i < len(tasks); i++ {
							idstoupd[i] = tasks[i].ID
						}
						updated := sqla.UpdateMultipleWithOneInt(tb.db, tb.dbType, "tasks", "TaskStatus", statusCode, "StatusSet", datetime.DateTimeToInt64(datetime.GetCurrentDateTime()), idstoupd)
						if updated > 0 {
							Page.Message = "statusUpdated"
							Page.UpdatedNum = updated
							for i := 0; i < len(tasks); i++ {
								tasks[i].TaskStatus = statusCode
							}
							for i := range idstoupd {
								t := tasks[i]
								participants, _ := t.loadParticipants(tb.db, tb.dbType)
								email := mail.EmailMessage{Subj: tb.i18n.messages.Subj.TaskStatusChanged + " [" + tb.i18n.taskCaption + " #" + strconv.Itoa(t.ID) + "]"}
								if t.Creator != nil && t.Creator.Contacts.Email != "" && t.Creator.UserLock == 0 {
									email.SendTo = append(email.SendTo, mail.UserToSend{Name: t.Creator.FirstName + " " + t.Creator.Surname, Email: t.Creator.Contacts.Email})
								}
								if t.Assignee != nil && t.Assignee.Contacts.Email != "" && t.Assignee.UserLock == 0 {
									email.SendTo = append(email.SendTo, mail.UserToSend{Name: t.Assignee.FirstName + " " + t.Assignee.Surname, Email: t.Assignee.Contacts.Email})
								}
								for i := 0; i < len(participants); i++ {
									if participants[i].Contacts.Email != "" && participants[i].UserLock == 0 {
										email.SendCc = append(email.SendCc, mail.UserToSend{Name: participants[i].FirstName + " " + participants[i].Surname, Email: participants[i].Contacts.Email})
									}
								}
								taskMail := TaskMail{email.Subj, t, tb.i18n.messages, tb.i18n.taskCaption, tb.i18n.taskStatuses, tb.cfg.systemURL}
								taskMail.constructToChannel(tb.db, tb.dbType, tb.taskmailtmpl, tb.mailchan, email, tb.emailCont)
							}
						} else {
							Page.Message = "statusUpdateError"
							log.Println("Error updating tasks status:", ids)
						}
					} else {
						Page.Message = "statusUpdated"
						Page.UpdatedNum = numtoupd
					}
				} else {
					Page.Message = "notAllorSomeElemsAllowedtoModify"
				}
			}
		}
	}

	// Processing removal
	if r.Method == "POST" && r.FormValue("deleteButton") != "" {
		elemsOnCurrentPage, _ := strconv.Atoi(r.FormValue("elemsOnCurrentPage"))
		r.ParseForm()
		ids := []int{}
		for _, v := range r.Form["ids"] {
			id, _ := strconv.Atoi(v)
			ids = append(ids, id)
		}
		if len(ids) > 0 {
			allowedToRemove := sqla.VerifyRemovalPermissions(tb.db, tb.dbType, "tasks", Page.LoggedinID, Page.LoggedinAdmin, Page.RemoveAllowed, ids)
			if allowedToRemove {
				removed := sqla.DeleteObjects(tb.db, tb.dbType, "tasks", "ID", ids)
				if removed > 0 {
					core.RemoveUploadedDirs(filepath.Join(tb.cfg.serverRoot, "files", "tasks"), ids)
					Page.Message = "removedElems"
					Page.RemovedNum = removed
					if removed >= elemsOnCurrentPage && Page.PageNumber > 1 {
						Page.PageNumber--
					}
				} else {
					Page.Message = "removalError"
					log.Println("Error removing tasks:", ids)
				}
			} else {
				Page.Message = "notAllorSomeElemsAllowedtoRemove"
				log.Println("Not allowed to remove attempt, LoggedinID and ids:", Page.LoggedinID, ids)
			}
		}
	}

	joins := `LEFT JOIN profiles creator ON creator.ID = tasks.Creator
LEFT JOIN profiles assignee ON assignee.ID = Assignee`
	columnsToCount := "*"
	if Page.Filters.TextFilter != "" {
		joins += " LEFT JOIN comments ON comments.Task = tasks.ID"
	}

	DISTINCT := false
	if Page.Filters.TextFilter != "" {
		DISTINCT = true
		columnsToCount = "tasks.ID"
	}

	OFFSET := (Page.PageNumber - 1) * Page.UserConfig.ElemsOnPage
	LIMIT := Page.UserConfig.ElemsOnPage
	if OFFSET < 0 {
		OFFSET = 0
	}
	SEEK := sqla.Seek{
		Value:   0,
		UseSeek: false,
	}

	sortedHowReverse := false
	if SQLSortedBy == "tasks.ID" && Page.PageNumber != 1 {
		SEEK.UseSeek = true
	}

	if SEEK.UseSeek {
		previousPageNumber, _ := strconv.Atoi(r.FormValue("previousPage"))
		filteredNum, _ := strconv.Atoi(r.FormValue("filteredNum"))
		if previousPageNumber == 0 {
			previousPageNumber = 1
		}
		pageDifference := Page.PageNumber - previousPageNumber

		if Page.PageNumber == accs.CalcMaxPages(Page.UserConfig.ElemsOnPage, filteredNum) {
			remainder := filteredNum - OFFSET
			SEEK.UseSeek = false
			sortedHowReverse = true
			if Page.SortedHow == 0 {
				Page.SortedHow = 1
			} else {
				Page.SortedHow = 0
			}
			LIMIT = remainder
			OFFSET = 0
		} else if pageDifference > 0 {
			SEEK.Value, _ = strconv.Atoi(r.FormValue("lastElemOnPage"))
			OFFSET = (pageDifference - 1) * Page.UserConfig.ElemsOnPage
		} else if pageDifference < 0 {
			SEEK.Value, _ = strconv.Atoi(r.FormValue("firstElemOnPage"))
			sortedHowReverse = true
			if Page.SortedHow == 0 {
				Page.SortedHow = 1
			} else {
				Page.SortedHow = 0
			}
			OFFSET = (-pageDifference - 1) * Page.UserConfig.ElemsOnPage
		} else {
			SEEK.Value, _ = strconv.Atoi(r.FormValue("firstElemOnPage"))
			SEEK.ValueInclude = true
			OFFSET = 0
		}
	}

	columns := `tasks.ID, tasks.Created, PlanStart, PlanDue, StatusSet, tasks.Creator, Assignee, Participants, Topic, tasks.Content, TaskStatus, tasks.FileList,
creator.ID, creator.FirstName, creator.Surname, creator.JobTitle,
assignee.ID, assignee.FirstName, assignee.Surname, assignee.JobTitle`
	if DISTINCT && tb.dbType == sqla.ORACLE {
		columns = `tasks.ID, tasks.Created, PlanStart, PlanDue, StatusSet, tasks.Creator, Assignee, Participants, Topic, dbms_lob.substr(tasks.Content, 4000, 1), TaskStatus, dbms_lob.substr(tasks.FileList, 4000, 1),
creator.ID, creator.FirstName, creator.Surname, creator.JobTitle,
assignee.ID, assignee.FirstName, assignee.Surname, assignee.JobTitle`
	}

	sq, sqcount, args, argscount := sqla.ConstructSELECTquery(
		tb.dbType,
		"tasks",
		columns,
		columnsToCount,
		joins,
		Page.Filters,
		SQLSortedBy,
		Page.SortedHow,
		LIMIT,
		OFFSET,
		DISTINCT,
		SEEK)

	// var timeBefore time.Time
	// if DEBUG {
	// 	timeBefore = time.Now()
	// }

	// Loading objects
	err = func() error {
		rows, err := tb.db.Query(sq, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		var Created sql.NullInt64
		var PlanStart sql.NullInt64
		var PlanDue sql.NullInt64
		var StatusSet sql.NullInt64
		var Creator sql.NullInt64
		var Assignee sql.NullInt64
		var Participants sql.NullString
		var Topic sql.NullString
		var Content sql.NullString
		var TaskStatus sql.NullInt64
		var FileList sql.NullString

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
			err = rows.Scan(&t.ID, &Created, &PlanStart, &PlanDue, &StatusSet, &Creator, &Assignee, &Participants, &Topic, &Content, &TaskStatus, &FileList,
				&CreatorID, &CreatorFirstName, &CreatorSurname, &CreatorJobTitle,
				&AssigneeID, &AssigneeFirstName, &AssigneeSurname, &AssigneeJobTitle)
			if err != nil {
				return err
			}

			t.Created = datetime.Int64ToDateTime(Created.Int64)
			t.PlanStart = datetime.Int64ToDateTime(PlanStart.Int64)
			t.PlanDue = datetime.Int64ToDateTime(PlanDue.Int64)
			t.StatusSet = datetime.Int64ToDateTime(StatusSet.Int64)
			if CreatorID.Valid == true {
				t.Creator = &team.Profile{
					ID:        int(CreatorID.Int64),
					FirstName: CreatorFirstName.String,
					Surname:   CreatorSurname.String,
					JobTitle:  CreatorJobTitle.String,
				}
			} else {
				t.Creator = nil
			}
			if AssigneeID.Valid == true {
				t.Assignee = &team.Profile{
					ID:        int(AssigneeID.Int64),
					FirstName: AssigneeFirstName.String,
					Surname:   AssigneeSurname.String,
					JobTitle:  AssigneeJobTitle.String,
				}
			} else {
				t.Assignee = nil
			}

			t.Participants = sqla.UnmarshalNonEmptyJSONListInt(Participants.String)
			t.Topic = Topic.String
			t.Content = Content.String
			t.TaskStatus = int(TaskStatus.Int64)
			t.FileList = sqla.UnmarshalNonEmptyJSONList(FileList.String)
			Page.Tasks = append(Page.Tasks, t)
		}

		var FilteredNum sql.NullInt64
		row := tb.db.QueryRow(sqcount, argscount...)
		err = row.Scan(&FilteredNum)
		if err != nil {
			return err
		}
		Page.FilteredNum = int(FilteredNum.Int64)

		return nil
	}()

	if err != nil {
		accs.ThrowServerError(w, fmt.Sprintf(accs.CurrentFunction()+":", err), Page.LoggedinID, 0)
		return
	}

	if sortedHowReverse {
		if Page.SortedHow == 0 {
			Page.SortedHow = 1
		} else {
			Page.SortedHow = 0
		}
		for i, j := 0, len(Page.Tasks)-1; i < j; i, j = i+1, j-1 {
			Page.Tasks[i], Page.Tasks[j] = Page.Tasks[j], Page.Tasks[i]
		}
	}

	// if DEBUG {
	// 	timeAfter := time.Now()
	// 	diff := timeAfter.Sub(timeBefore)
	// 	log.Printf("SQL execution time in milliseconds: %d\n", diff.Milliseconds())
	// }

	Page.UserList = tb.memorydb.GetObjectArr("UserList")

	// Attention! This removes db column lists in outputs like JSON.
	// Usually columns should not be available to a client.
	Page.Filters.ClearColumnsValues()

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		return
	}

	err = tb.templates.ExecuteTemplate(w, "tasks.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func checkStatusModifyPermissions(db *sql.DB, DBType byte, table string, Owner int, AdminPrivileges bool, ids []int) bool {
	if AdminPrivileges {
		return true
	}
	var argsCounter int
	var args, argstoAppend []interface{}
	var sqlids []int

	var sq = "SELECT ID FROM " + table + " WHERE (Creator = " + sqla.MakeParam(DBType, argsCounter+1) +
		" OR Assignee = " + sqla.MakeParam(DBType, argsCounter+2) + ") "
	argsCounter += 2
	args = append(args, Owner, Owner)

	argsCounter, sq, argstoAppend = sqla.BuildSQLIN(DBType, sq, argsCounter, "ID", ids)
	args = append(args, argstoAppend...)

	if core.DEBUG {
		log.Println(sq, args)
	}
	rows, err := db.Query(sq, args...)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
	}
	defer rows.Close()
	var ID sql.NullInt64
	for rows.Next() {
		err = rows.Scan(&ID)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
		sqlids = append(sqlids, int(ID.Int64))
	}
	sort.Ints(ids)
	sort.Ints(sqlids)
	if accs.IntSlicesEqual(ids, sqlids) {
		return true
	}

	return false
}

func loadTasks(db *sql.DB, sq string, sqcount string, args []interface{}) (Tasks []Task, FilteredNumRes int, err error) {
	rows, err := db.Query(sq, args...)
	if err != nil {
		return Tasks, FilteredNumRes, err
	}
	defer rows.Close()

	var Created sql.NullInt64
	var PlanStart sql.NullInt64
	var PlanDue sql.NullInt64
	var StatusSet sql.NullInt64
	var Creator sql.NullInt64
	var Assignee sql.NullInt64
	var Participants sql.NullString
	var Topic sql.NullString
	var Content sql.NullString
	var TaskStatus sql.NullInt64
	var FileList sql.NullString

	var CreatorID sql.NullInt64
	var CreatorFirstName sql.NullString
	var CreatorSurname sql.NullString
	var CreatorJobTitle sql.NullString
	var CreatorContacts sql.NullString
	var CreatorUserLock sql.NullInt64

	var AssigneeID sql.NullInt64
	var AssigneeFirstName sql.NullString
	var AssigneeSurname sql.NullString
	var AssigneeJobTitle sql.NullString
	var AssigneeContacts sql.NullString
	var AssigneeUserLock sql.NullInt64

	for rows.Next() {
		var t Task
		err = rows.Scan(&t.ID, &Created, &PlanStart, &PlanDue, &StatusSet, &Creator, &Assignee, &Participants, &Topic, &Content, &TaskStatus, &FileList,
			&CreatorID, &CreatorFirstName, &CreatorSurname, &CreatorJobTitle, &CreatorContacts, &CreatorUserLock,
			&AssigneeID, &AssigneeFirstName, &AssigneeSurname, &AssigneeJobTitle, &AssigneeContacts, &AssigneeUserLock)
		if err != nil {
			return Tasks, FilteredNumRes, err
		}

		t.Created = datetime.Int64ToDateTime(Created.Int64)
		t.PlanStart = datetime.Int64ToDateTime(PlanStart.Int64)
		t.PlanDue = datetime.Int64ToDateTime(PlanDue.Int64)
		t.StatusSet = datetime.Int64ToDateTime(StatusSet.Int64)
		if CreatorID.Valid == true {
			t.Creator = &team.Profile{
				ID:        int(CreatorID.Int64),
				FirstName: CreatorFirstName.String,
				Surname:   CreatorSurname.String,
				JobTitle:  CreatorJobTitle.String,
				Contacts:  team.UnmarshalNonEmptyProfileContacts(CreatorContacts.String),
				UserLock:  int(CreatorUserLock.Int64),
			}
		} else {
			t.Creator = nil
		}
		if AssigneeID.Valid == true {
			t.Assignee = &team.Profile{
				ID:        int(AssigneeID.Int64),
				FirstName: AssigneeFirstName.String,
				Surname:   AssigneeSurname.String,
				JobTitle:  AssigneeJobTitle.String,
				Contacts:  team.UnmarshalNonEmptyProfileContacts(AssigneeContacts.String),
				UserLock:  int(AssigneeUserLock.Int64),
			}
		} else {
			t.Assignee = nil
		}

		t.Participants = sqla.UnmarshalNonEmptyJSONListInt(Participants.String)
		t.Topic = Topic.String
		t.Content = Content.String
		t.TaskStatus = int(TaskStatus.Int64)
		t.FileList = sqla.UnmarshalNonEmptyJSONList(FileList.String)
		Tasks = append(Tasks, t)
	}

	var FilteredNum sql.NullInt64
	row := db.QueryRow(sqcount, args...)
	err = row.Scan(&FilteredNum)
	if err != nil {
		return Tasks, FilteredNumRes, err
	}
	FilteredNumRes = int(FilteredNum.Int64)

	return Tasks, FilteredNumRes, nil
}
