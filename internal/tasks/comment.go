package tasks

import (
	"database/sql"
	"edm/internal/team"
	"edm/pkg/datetime"

	"github.com/alecxcode/sqla"
)

// Comment is a discussion item to a task
type Comment struct {
	//sql generate
	ID       int
	Created  datetime.DateTime `sql-gen:"bigint"`
	Creator  *team.Profile     `sql-gen:"FK_NULL"`
	Task     int               `sql-gen:"IDX,FK_CASCADE,fktable(tasks)"`
	Content  string            `sql-gen:"varchar(max)"`
	FileList []string          `sql-gen:"varchar(max)"`
}

// Create a comment in DB
func (c *Comment) Create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	if c.Created.Day != 0 {
		args = args.AppendInt64("Created", datetime.DateTimeToInt64(c.Created))
	}
	args = args.AppendInt("Creator", c.Creator.ID)
	args = args.AppendInt("Task", c.Task)
	args = args.AppendNonEmptyString("Content", c.Content)
	args = args.AppendJSONList("FileList", c.FileList)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "comments", args)
	return lastid, rowsaff
}

// GiveCreatorID executes in a template to deliver the creator ID of this object
func (c Comment) GiveCreatorID() int {
	if c.Creator == nil {
		return 0
	} else {
		return c.Creator.ID
	}
}

// Inci executes in a template to deliver value increment of +1
func (c Comment) Inci(n int) int {
	return n + 1
}

// GiveDateTime executes in a template to deliver the queried date and time of a comment
func (c Comment) GiveDateTime(dateFmt string, timeFmt string, sep string) string {
	var dt = c.Created
	var rt string
	if timeFmt == "12h am/pm" {
		rt = datetime.TimeToString12(dt.Hour, dt.Minute)
	} else if timeFmt == "24h" {
		rt = datetime.TimeToString24(dt.Hour, dt.Minute)
	} else {
		rt = datetime.TimeToString24(dt.Hour, dt.Minute)
	}
	if dt.Day == 0 {
		return ""
	}
	return datetime.DateToString(datetime.Date{Year: dt.Year, Month: dt.Month, Day: dt.Day}, dateFmt) + sep + rt
}

func (t *Task) loadComments(db *sql.DB, DBType byte) (CommList []Comment, err error) {

	rows, err := db.Query(`SELECT c.ID, c.Created, c.Creator, c.Content, c.FileList,
p.ID, p.FirstName, p.Surname,  p.JobTitle
FROM comments c
LEFT JOIN profiles p ON p.ID = c.Creator
WHERE Task = `+sqla.MakeParam(DBType, 1)+` ORDER BY c.Created ASC, c.ID ASC`, t.ID)
	if err != nil {
		return CommList, err
	}
	defer rows.Close()

	var Created sql.NullInt64
	var Creator sql.NullInt64
	var Content sql.NullString
	var FileList sql.NullString

	var CreatorID sql.NullInt64
	var CreatorFirstName sql.NullString
	var CreatorSurname sql.NullString
	var CreatorJobTitle sql.NullString

	for rows.Next() {
		var c Comment
		err = rows.Scan(&c.ID, &Created, &Creator, &Content, &FileList,
			&CreatorID, &CreatorFirstName, &CreatorSurname, &CreatorJobTitle)
		if err != nil {
			return CommList, err
		}
		c.Created = datetime.Int64ToDateTime(Created.Int64)
		if Creator.Valid {
			c.Creator = &team.Profile{
				ID:        int(CreatorID.Int64),
				FirstName: CreatorFirstName.String,
				Surname:   CreatorSurname.String,
				JobTitle:  CreatorJobTitle.String,
			}
		}
		c.Content = Content.String
		c.FileList = sqla.UnmarshalNonEmptyJSONList(FileList.String)
		CommList = append(CommList, c)
	}

	return CommList, nil
}
