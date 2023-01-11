package tasks

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/datetime"
	"fmt"
	"log"
	"path/filepath"

	"github.com/alecxcode/sqla"
)

// Task describes a task to perform
type Task struct {
	//sql generate
	ID           int
	Created      datetime.DateTime `sql-gen:"bigint,IDX"`
	PlanStart    datetime.DateTime `sql-gen:"bigint,IDX"`
	PlanDue      datetime.DateTime `sql-gen:"bigint,IDX"`
	StatusSet    datetime.DateTime `sql-gen:"bigint,IDX"`
	Creator      *team.Profile     `sql-gen:"FK_NULL,FK_NOACTION"`
	Assignee     *team.Profile     `sql-gen:"FK_NULL,FK_NOACTION"`
	Participants []int             `sql-gen:"varchar(4000)"`
	Topic        string            `sql-gen:"varchar(255)"`
	Content      string            `sql-gen:"varchar(max)"`
	TaskStatus   int
	Project      int      `sql-gen:"IDX,FK_NULL,fktable(projects)"`
	FileList     []string `sql-gen:"varchar(max)"`
}

// Task statuses
const (
	CREATED = iota
	ASSIGNED
	INPROGRESS
	STUCK
	DONE
	CANCELED
	INREVIEW
)

func (t Task) print() {
	fmt.Printf("%#v\n", t)
}

// GiveStatus executes in a template to deliver the status of an object
func (t Task) GiveStatus(stslice []string, unknown string) string {
	if t.TaskStatus < len(stslice) && t.TaskStatus >= core.Undefined {
		return stslice[t.TaskStatus]
	} else {
		return unknown
	}
}

// GiveCreatorID executes in a template to deliver the creator ID of this object
func (t Task) GiveCreatorID() int {
	if t.Creator == nil {
		return 0
	} else {
		return t.Creator.ID
	}
}

// GiveAssigneeID executes in a template to deliver the assignee ID of this object
func (t Task) GiveAssigneeID() int {
	if t.Assignee == nil {
		return 0
	} else {
		return t.Assignee.ID
	}
}

// GiveDateTime executes in a template to deliver the queried date and time of a task
func (t Task) GiveDateTime(dateWhat string, dateFmt string, timeFmt string, sep string) string {

	var dt datetime.DateTime
	var rt string

	switch dateWhat {
	case "Created":
		dt = t.Created
	case "PlanStart":
		dt = t.PlanStart
	case "PlanDue":
		dt = t.PlanDue
	case "StatusSet":
		dt = t.StatusSet
	default:
		return "wrong arg"
	}

	if dt.Day == 0 {
		return ""
	}

	if timeFmt == "12h am/pm" {
		rt = datetime.TimeToString12(dt.Hour, dt.Minute)
	} else if timeFmt == "24h" {
		rt = datetime.TimeToString24(dt.Hour, dt.Minute)
	} else {
		rt = datetime.TimeToString24(dt.Hour, dt.Minute)
	}

	return datetime.DateToString(datetime.Date{Year: dt.Year, Month: dt.Month, Day: dt.Day}, dateFmt) + sep + rt
}

// GiveShortFileName executes in a template to deliver the shortened file name
func (t Task) GiveShortFileName(index int) string {
	if index >= len(t.FileList) {
		return ""
	} else {
		return string([]rune(t.FileList[index])[0]) + ".." + filepath.Ext(t.FileList[index])
	}
}

// Create creates an object in DB
func (t *Task) Create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	if t.Created.Day != 0 {
		args = args.AppendInt64("Created", datetime.DateTimeToInt64(t.Created))
	}
	if t.PlanStart.Day != 0 {
		args = args.AppendInt64("PlanStart", datetime.DateTimeToInt64(t.PlanStart))
	}
	if t.PlanDue.Day != 0 {
		args = args.AppendInt64("PlanDue", datetime.DateTimeToInt64(t.PlanDue))
	}
	if t.StatusSet.Day != 0 {
		args = args.AppendInt64("StatusSet", datetime.DateTimeToInt64(t.StatusSet))
	}
	if t.Creator != nil {
		args = args.AppendInt("Creator", t.Creator.ID)
	}
	if t.Assignee != nil {
		args = args.AppendInt("Assignee", t.Assignee.ID)
	}
	args = args.AppendNonEmptyString("Topic", t.Topic)
	args = args.AppendNonEmptyString("Content", t.Content)
	args = args.AppendInt("TaskStatus", t.TaskStatus)
	if t.Project != core.Undefined {
		args = args.AppendInt("Project", t.Project)
	}
	args = args.AppendJSONList("FileList", t.FileList)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "tasks", args)
	return lastid, rowsaff
}

func (t *Task) update(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	if t.PlanStart.Day != 0 {
		args = args.AppendInt64("PlanStart", datetime.DateTimeToInt64(t.PlanStart))
	} else {
		args = args.AppendNil("PlanStart")
	}
	if t.PlanDue.Day != 0 {
		args = args.AppendInt64("PlanDue", datetime.DateTimeToInt64(t.PlanDue))
	} else {
		args = args.AppendNil("PlanDue")
	}
	if t.StatusSet.Day != 0 {
		args = args.AppendInt64("StatusSet", datetime.DateTimeToInt64(t.StatusSet))
	}
	if t.Assignee != nil {
		args = args.AppendInt("Assignee", t.Assignee.ID)
	} else {
		args = args.AppendNil("Assignee")
	}
	args = args.AppendStringOrNil("Topic", t.Topic)
	args = args.AppendStringOrNil("Content", t.Content)
	args = args.AppendInt("TaskStatus", t.TaskStatus)
	if t.Project != core.Undefined {
		args = args.AppendInt("Project", t.Project)
	} else {
		args = args.AppendNil("Project")
	}
	args = args.AppendJSONList("FileList", t.FileList)
	rowsaff = sqla.UpdateObject(db, DBType, "tasks", args, t.ID)
	return rowsaff
}

func (t *Task) updateStatus(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendInt64("StatusSet", datetime.DateTimeToInt64(t.StatusSet))
	args = args.AppendInt("TaskStatus", t.TaskStatus)
	rowsaff = sqla.UpdateObject(db, DBType, "tasks", args, t.ID)
	return rowsaff
}

func (t *Task) load(db *sql.DB, DBType byte) error {

	row := db.QueryRow(`SELECT
tasks.ID, Created, PlanStart, PlanDue, StatusSet, Creator, Assignee, Participants, Topic, Content, TaskStatus, Project, FileList,
creator.ID, creator.FirstName, creator.Surname, creator.JobTitle, creator.Contacts, creator.UserLock,
assignee.ID, assignee.FirstName, assignee.Surname, assignee.JobTitle, assignee.Contacts, assignee.UserLock
FROM tasks
LEFT JOIN profiles creator ON creator.ID = Creator
LEFT JOIN profiles assignee ON assignee.ID = Assignee
WHERE tasks.ID = `+sqla.MakeParam(DBType, 1), t.ID)

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
	var Project sql.NullInt64
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

	err := row.Scan(&t.ID, &Created, &PlanStart, &PlanDue, &StatusSet, &Creator, &Assignee, &Participants, &Topic, &Content, &TaskStatus, &Project, &FileList,
		&CreatorID, &CreatorFirstName, &CreatorSurname, &CreatorJobTitle, &CreatorContacts, &CreatorUserLock,
		&AssigneeID, &AssigneeFirstName, &AssigneeSurname, &AssigneeJobTitle, &AssigneeContacts, &AssigneeUserLock)
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
	t.Project = int(Project.Int64)
	t.FileList = sqla.UnmarshalNonEmptyJSONList(FileList.String)

	return nil
}

func (t *Task) loadParticipants(db *sql.DB, DBType byte) (ProfList []team.Profile, err error) {

	if len(t.Participants) > 0 {
		var sq = "SELECT ID, FirstName, Surname, JobTitle, Contacts, UserLock FROM profiles "
		var args, argstoAppend []interface{}

		_, sq, argstoAppend = sqla.BuildSQLIN(DBType, sq, 0, "ID", t.Participants)
		args = append(args, argstoAppend...)

		if core.DEBUG {
			log.Println(sq, t.Participants)
		}
		rows, err := db.Query(sq, args...)
		if err != nil {
			return ProfList, err
		}
		defer rows.Close()

		var FirstName sql.NullString
		var Surname sql.NullString
		var JobTitle sql.NullString
		var Contacts sql.NullString
		var UserLock sql.NullInt64

		for rows.Next() {
			var p team.Profile
			err = rows.Scan(&p.ID, &FirstName, &Surname, &JobTitle, &Contacts, &UserLock)
			if err != nil {
				return ProfList, err
			}
			p.FirstName = FirstName.String
			p.Surname = Surname.String
			p.JobTitle = JobTitle.String
			p.Contacts = team.UnmarshalNonEmptyProfileContacts(Contacts.String)
			p.UserLock = int(UserLock.Int64)
			ProfList = append(ProfList, p)
		}

	}
	return ProfList, nil
}

func (t *Task) giveTaskProjectCreatorID(db *sql.DB, DBType byte) (int, error) {
	row := db.QueryRow(`SELECT profiles.ID FROM profiles
LEFT JOIN projects ON profiles.ID = projects.Creator
WHERE projects.ID = `+sqla.MakeParam(DBType, 1), t.Project)

	var ID int
	err := row.Scan(&ID)
	if err != nil {
		return 0, err
	}
	return ID, nil
}

func selectTaskStatus(val string) (statusCode int) {
	if val == "inprogress" {
		statusCode = INPROGRESS
	} else if val == "stuck" {
		statusCode = STUCK
	} else if val == "done" {
		statusCode = DONE
	} else if val == "canceled" {
		statusCode = CANCELED
	} else if val == "inreview" {
		statusCode = INREVIEW
	}
	return statusCode
}
