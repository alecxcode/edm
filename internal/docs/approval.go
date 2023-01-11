package docs

import (
	"database/sql"
	"edm/internal/team"
	"edm/pkg/datetime"

	"github.com/alecxcode/sqla"
)

// Approval is an approval item to a document
type Approval struct {
	//sql generate
	ID           int
	Written      datetime.DateTime `sql-gen:"bigint"`
	Approver     *team.Profile     `sql-gen:"FK_NULL"`
	ApproverSign string            `sql-gen:"varchar(2000)"`
	DocID        int               `sql-gen:"IDX,FK_CASCADE,fktable(documents)"`
	Approved     int
	Note         string `sql-gen:"varchar(max)"`
}

// Consts for approval states
const (
	NOACTION = 0
	APPROVED = 1
	REJECTED = 2
	BROKEN   = 3
)

func (a *Approval) create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	if a.Approver != nil {
		args = args.AppendInt("Approver", a.Approver.ID)
	}
	args = args.AppendInt("DocID", a.DocID)
	args = args.AppendInt("Approved", NOACTION)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "approvals", args)
	return lastid, rowsaff
}

func (a *Approval) sign(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	if a.Written.Day != 0 {
		args = args.AppendInt64("Written", datetime.DateTimeToInt64(a.Written))
	}
	args = args.AppendStringOrNil("ApproverSign", a.ApproverSign)
	args = args.AppendInt("DocID", a.DocID)
	args = args.AppendInt("Approved", a.Approved)
	args = args.AppendStringOrNil("Note", a.Note)
	rowsaff = sqla.UpdateObject(db, DBType, "approvals", args, a.ID)
	return rowsaff
}

// GiveApproverID executes in a template to deliver the approver ID of this object
func (a Approval) GiveApproverID() int {
	if a.Approver == nil {
		return 0
	} else {
		return a.Approver.ID
	}
}

func (d *Document) loadApprovals(db *sql.DB, DBType byte) (ApprovalList []Approval, err error) {

	rows, err := db.Query(`SELECT a.ID, a.Written, a.Approver, a.ApproverSign, a.Approved, a.Note,
p.ID, p.FirstName, p.Surname, p.JobTitle
FROM approvals a
LEFT JOIN profiles p ON p.ID = a.Approver
WHERE DocID = `+sqla.MakeParam(DBType, 1)+` ORDER BY a.Written ASC, a.ID ASC`, d.ID)
	if err != nil {
		return ApprovalList, err
	}
	defer rows.Close()

	var ID sql.NullInt64
	var Written sql.NullInt64
	var Approver sql.NullInt64
	var ApproverSign sql.NullString
	var Approved sql.NullInt64
	var Note sql.NullString

	var ApproverID sql.NullInt64
	var ApproverFirstName sql.NullString
	var ApproverSurname sql.NullString
	var ApproverJobTitle sql.NullString

	for rows.Next() {
		err = rows.Scan(&ID, &Written, &Approver, &ApproverSign, &Approved, &Note,
			&ApproverID, &ApproverFirstName, &ApproverSurname, &ApproverJobTitle)
		if err != nil {
			return ApprovalList, err
		}
		a := Approval{
			ID:           int(ID.Int64),
			Written:      datetime.Int64ToDateTime(Written.Int64),
			ApproverSign: ApproverSign.String,
			DocID:        d.ID,
			Approved:     int(Approved.Int64),
			Note:         Note.String,
		}
		if Approver.Valid {
			a.Approver = &team.Profile{
				ID:        int(ApproverID.Int64),
				FirstName: ApproverFirstName.String,
				Surname:   ApproverSurname.String,
				JobTitle:  ApproverJobTitle.String,
			}
		}
		ApprovalList = append(ApprovalList, a)
	}

	return ApprovalList, nil
}

// GiveDateTime executes in a template to deliver the queried date and time of an approval
func (a Approval) GiveDateTime(dateFmt string, timeFmt string, sep string) string {
	var dt = a.Written
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

type approvals []Approval

func (as approvals) getApprovalsIDsSlice() []int {
	appids := make([]int, len(as))
	for i := 0; i < len(as); i++ {
		appids[i] = as[i].ID
	}
	return appids
}

func (as approvals) getApprovalsIDsSliceApproved() []int {
	appids := []int{}
	for i := 0; i < len(as); i++ {
		if as[i].Approved == APPROVED {
			appids = append(appids, as[i].ID)
		}
	}
	return appids
}

func (as approvals) getApproversIDsSlice() []int {
	cids := make([]int, len(as))
	for i := 0; i < len(as); i++ {
		cids[i] = as[i].GiveApproverID()
	}
	return cids
}

func (as approvals) getApprovalIDbyDocIDandApproverID(docID int, pID int) int {
	for _, a := range as {
		if a.Approver != nil && a.DocID == docID && a.Approver.ID == pID {
			return a.ID
		}
	}
	return 0
}

func (as approvals) getApprovalByID(aID int) Approval {
	for _, a := range as {
		if a.ID == aID {
			return a
		}
	}
	return Approval{ID: 0}
}

func (as approvals) approved(docID int, pID int) int {
	for _, a := range as {
		if a.Approver != nil && a.DocID == docID && a.Approver.ID == pID {
			return a.Approved
		}
	}
	return NOACTION
}

// GetApprovalNote executes in template
func (as approvals) GetApprovalNote(docID int, pID int) string {
	for _, a := range as {
		if a.Approver != nil && a.DocID == docID && a.Approver.ID == pID {
			return a.Note
		}
	}
	return ""
}
