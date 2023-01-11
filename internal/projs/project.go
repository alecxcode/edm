package projs

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"

	"github.com/alecxcode/sqla"
)

// Project describes a project
type Project struct {
	//sql generate
	ID          int
	ProjName    string        `sql-gen:"varchar(255)"`
	Description string        `sql-gen:"varchar(max)"`
	Creator     *team.Profile `sql-gen:"FK_NULL"`
	ProjStatus  int
}

// GiveStatus executes in a template to deliver the status of an object
func (p Project) GiveStatus(stslice []string, unknown string) string {
	if p.ProjStatus < len(stslice) && p.ProjStatus >= core.Undefined {
		return stslice[p.ProjStatus]
	} else {
		return unknown
	}
}

// GiveCreatorID executes in a template to deliver the creator ID of this object
func (p Project) GiveCreatorID() int {
	if p.Creator == nil {
		return 0
	} else {
		return p.Creator.ID
	}
}

// Create creates this object
func (p *Project) Create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	if p.Creator != nil {
		args = args.AppendInt("Creator", p.Creator.ID)
	}
	args = args.AppendNonEmptyString("ProjName", p.ProjName)
	args = args.AppendNonEmptyString("Description", p.Description)
	args = args.AppendInt("ProjStatus", p.ProjStatus)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "projects", args)
	return lastid, rowsaff
}

func (p *Project) update(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	if p.Creator != nil {
		args = args.AppendInt("Creator", p.Creator.ID)
	}
	args = args.AppendStringOrNil("ProjName", p.ProjName)
	args = args.AppendStringOrNil("Description", p.Description)
	args = args.AppendInt("ProjStatus", p.ProjStatus)
	rowsaff = sqla.UpdateObject(db, DBType, "projects", args, p.ID)
	return rowsaff
}

func (p *Project) load(db *sql.DB, DBType byte) error {
	row := db.QueryRow(`SELECT
projects.ID, ProjName, Description, Creator, ProjStatus,
creator.ID, creator.FirstName, creator.Surname, creator.JobTitle
FROM projects
LEFT JOIN profiles creator ON creator.ID = Creator
WHERE projects.ID = `+sqla.MakeParam(DBType, 1), p.ID)

	var ProjName sql.NullString
	var Description sql.NullString
	var Creator sql.NullInt64
	var ProjStatus sql.NullInt64

	var CreatorID sql.NullInt64
	var CreatorFirstName sql.NullString
	var CreatorSurname sql.NullString
	var CreatorJobTitle sql.NullString

	err := row.Scan(&p.ID, &ProjName, &Description, &Creator, &ProjStatus,
		&CreatorID, &CreatorFirstName, &CreatorSurname, &CreatorJobTitle)
	if err != nil {
		return err
	}
	p.ProjName = ProjName.String
	p.Description = Description.String
	p.ProjStatus = int(ProjStatus.Int64)
	if CreatorID.Valid == true {
		p.Creator = &team.Profile{
			ID:        int(CreatorID.Int64),
			FirstName: CreatorFirstName.String,
			Surname:   CreatorSurname.String,
			JobTitle:  CreatorJobTitle.String,
		}
	} else {
		p.Creator = nil
	}

	return nil
}
