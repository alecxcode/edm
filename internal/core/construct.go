package core

import (
	"database/sql"
	"edm/pkg/memdb"
	"strconv"
	"strings"
)

// ObjectArrElem is an element for buildibg lists
type ObjectArrElem struct {
	ID    int
	Value string
}

// GetID is to satisfy ObjHasID interface
func (e ObjectArrElem) GetID() int {
	return e.ID
}

// ConstructUserList is to build user list to store in memory
func ConstructUserList(db *sql.DB, DBType byte, m memdb.ObjectsInMemory) error {
	rows, err := db.Query(`SELECT ID, FirstName, OtherName, Surname, JobTitle, UserLock FROM profiles
ORDER BY Surname ASC, FirstName ASC, OtherName ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ID int
	var FirstName sql.NullString
	var OtherName sql.NullString
	var Surname sql.NullString
	var JobTitle sql.NullString
	var UserLock sql.NullInt64
	UserList := []memdb.ObjHasID{}
	for rows.Next() {
		err = rows.Scan(&ID, &FirstName, &OtherName, &Surname, &JobTitle, &UserLock)
		if err != nil {
			return err
		}
		if int(UserLock.Int64) == 0 {
			var fullNameJob string
			if Surname.String != "" {
				fullNameJob += Surname.String + " "
			}
			if FirstName.String != "" {
				fullNameJob += FirstName.String + " "
			}
			if OtherName.String != "" {
				fullNameJob += OtherName.String
			}
			fullNameJob = strings.TrimSpace(fullNameJob)
			if fullNameJob == "" {
				fullNameJob = "ID: " + strconv.Itoa(ID)
			}
			if JobTitle.String != "" {
				fullNameJob += ", " + JobTitle.String
			}
			UserList = append(UserList, ObjectArrElem{ID, fullNameJob})
		}
	}
	m.SetObjectArr("UserList", UserList)
	return nil
}

// ConstructUnitList is to build unit list to store in memory
func ConstructUnitList(db *sql.DB, DBType byte, m memdb.ObjectsInMemory) error {
	rows, err := db.Query(`SELECT units.ID, units.UnitName,
companies.ShortName
FROM units
LEFT JOIN companies ON companies.ID = units.Company
ORDER BY units.UnitName ASC`)
	defer rows.Close()
	if err != nil {
		return err
	}
	var ID int
	var UnitName sql.NullString
	var CompanyShortName sql.NullString
	UnitList := []memdb.ObjHasID{}
	for rows.Next() {
		err = rows.Scan(&ID, &UnitName, &CompanyShortName)
		if err != nil {
			return err
		}
		var unitNameComp string
		unitNameComp = UnitName.String
		if unitNameComp == "" {
			unitNameComp = "ID: " + strconv.Itoa(ID)
		}
		if CompanyShortName.String != "" {
			unitNameComp += ", " + CompanyShortName.String
		}
		UnitList = append(UnitList, ObjectArrElem{ID, unitNameComp})
	}
	m.SetObjectArr("UnitList", UnitList)
	return nil
}

// ConstructCorpList is to build company list to store in memory
func ConstructCorpList(db *sql.DB, DBType byte, m memdb.ObjectsInMemory) error {
	rows, err := db.Query(`SELECT ID, ShortName FROM companies ORDER BY ShortName ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ID int
	var ShortName sql.NullString
	CorpList := []memdb.ObjHasID{}
	for rows.Next() {
		err = rows.Scan(&ID, &ShortName)
		if err != nil {
			return err
		}
		var companyName string
		companyName = ShortName.String
		if companyName == "" {
			companyName = "ID: " + strconv.Itoa(ID)
		}
		CorpList = append(CorpList, ObjectArrElem{ID, companyName})
	}
	m.SetObjectArr("CorpList", CorpList)
	return nil
}
