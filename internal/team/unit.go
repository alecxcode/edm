package team

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/alecxcode/sqla"
)

// Unit is a company department
type Unit struct {
	//sql generate
	ID       int
	UnitName string   `sql-gen:"varchar(1024)"`
	Company  *Company `sql-gen:"FK_CASCADE"`
	UnitHead *Profile `sql-gen:"FK_NULL"`
}

// Create saves to DB
func (u *Unit) Create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendNonEmptyString("UnitName", u.UnitName)
	if u.Company != nil {
		args = args.AppendInt("Company", u.Company.ID)
	}
	if u.UnitHead != nil {
		args = args.AppendInt("UnitHead", u.UnitHead.ID)
	}
	lastid, rowsaff = sqla.InsertObject(db, DBType, "units", args)
	return lastid, rowsaff
}

func (u *Unit) update(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendStringOrNil("UnitName", u.UnitName)
	if u.Company != nil {
		args = args.AppendInt("Company", u.Company.ID)
	} else {
		args = args.AppendNil("Company")
	}
	if u.UnitHead != nil {
		args = args.AppendInt("UnitHead", u.UnitHead.ID)
	} else {
		args = args.AppendNil("UnitHead")
	}
	rowsaff = sqla.UpdateObject(db, DBType, "units", args, u.ID)
	return rowsaff
}

// GiveHeadID executes in template to provide UnitHead ID
func (u Unit) GiveHeadID() int {
	if u.UnitHead == nil {
		return 0
	}
	return u.UnitHead.ID
}

// GiveHeadNameJob executes in template to provide UnitHead data
func (u Unit) GiveHeadNameJob() (head string) {
	if u.UnitHead == nil {
		return ""
	}
	head = strings.TrimSpace(u.UnitHead.FirstName + " " + u.UnitHead.Surname)
	if head == "" {
		head = "ID: " + strconv.Itoa(u.UnitHead.ID)
	}
	if u.UnitHead.JobTitle != "" {
		head += ", " + u.UnitHead.JobTitle
	}
	return head
}

func (c *Company) loadUnits(db *sql.DB, DBType byte) (UnitList []Unit, err error) {

	rows, err := db.Query(`SELECT u.ID, u.UnitName, u.UnitHead,
p.ID, p.FirstName, p.Surname,  p.JobTitle
FROM units u
LEFT JOIN profiles p ON p.ID = u.UnitHead
WHERE Company = `+sqla.MakeParam(DBType, 1)+` ORDER BY u.UnitName ASC`, c.ID)
	if err != nil {
		return UnitList, err
	}
	defer rows.Close()

	var UnitName sql.NullString
	var UnitHead sql.NullInt64

	var UnitHeadID sql.NullInt64
	var UnitHeadFirstName sql.NullString
	var UnitHeadSurname sql.NullString
	var UnitHeadJobTitle sql.NullString

	for rows.Next() {
		var unit Unit
		err = rows.Scan(&unit.ID, &UnitName, &UnitHead, &UnitHeadID, &UnitHeadFirstName, &UnitHeadSurname, &UnitHeadJobTitle)
		if err != nil {
			return UnitList, err
		}
		unit.UnitName = UnitName.String
		if UnitHead.Valid {
			unit.UnitHead = &Profile{
				ID:        int(UnitHead.Int64),
				FirstName: UnitHeadFirstName.String,
				Surname:   UnitHeadSurname.String,
				JobTitle:  UnitHeadJobTitle.String,
			}
		}
		UnitList = append(UnitList, unit)
	}

	return UnitList, nil
}
