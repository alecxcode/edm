package team

import (
	"database/sql"
	"edm/pkg/accs"
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/alecxcode/sqla"
)

// Company is an organization
type Company struct {
	//sql generate
	ID          int
	ShortName   string          `sql-gen:"varchar(255)"`
	FullName    string          `sql-gen:"varchar(512)"`
	ForeignName string          `sql-gen:"varchar(512)"`
	Contacts    CompanyContacts `sql-gen:"varchar(4000)"`
	CompanyHead *Profile        `sql-gen:"FK_NULL"`
	RegNo       string          `sql-gen:"varchar(255)"`
	TaxNo       string          `sql-gen:"varchar(255)"`
	BankDetails string          `sql-gen:"varchar(4000)"`
}

// CompanyContacts stores company contacts (JSON in DB)
type CompanyContacts struct {
	AddressReg  string
	AddressFact string
	Phone       string
	Email       string
	WebSite     string
	Other       string
}

func unmarshalNonEmptyCompanyContacts(c string) (res CompanyContacts) {
	if c != "" {
		err := json.Unmarshal([]byte(c), &res)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
	}
	return res
}

// Create saves to DB
func (c *Company) Create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendNonEmptyString("ShortName", c.ShortName)
	args = args.AppendNonEmptyString("FullName", c.FullName)
	args = args.AppendNonEmptyString("ForeignName", c.ForeignName)
	args = args.AppendJSONStruct("Contacts", c.Contacts)
	if c.CompanyHead != nil {
		args = args.AppendInt("CompanyHead", c.CompanyHead.ID)
	}
	args = args.AppendNonEmptyString("RegNo", c.RegNo)
	args = args.AppendNonEmptyString("TaxNo", c.TaxNo)
	args = args.AppendNonEmptyString("BankDetails", c.BankDetails)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "companies", args)
	return lastid, rowsaff
}

func (c *Company) update(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendStringOrNil("ShortName", c.ShortName)
	args = args.AppendStringOrNil("FullName", c.FullName)
	args = args.AppendStringOrNil("ForeignName", c.ForeignName)
	args = args.AppendJSONStruct("Contacts", c.Contacts)
	if c.CompanyHead != nil {
		args = args.AppendInt("CompanyHead", c.CompanyHead.ID)
	} else {
		args = args.AppendNil("CompanyHead")
	}
	args = args.AppendStringOrNil("RegNo", c.RegNo)
	args = args.AppendStringOrNil("TaxNo", c.TaxNo)
	args = args.AppendStringOrNil("BankDetails", c.BankDetails)
	rowsaff = sqla.UpdateObject(db, DBType, "companies", args, c.ID)
	return rowsaff
}

func (c *Company) load(db *sql.DB, DBType byte) error {

	row := db.QueryRow(`SELECT
c.ID, c.ShortName, c.FullName, c.ForeignName, c.Contacts, c.CompanyHead, c.RegNo, c.TaxNo, c.BankDetails,
p.ID, p.FirstName, p.Surname,  p.JobTitle
FROM companies c
LEFT JOIN profiles p ON p.ID = c.CompanyHead
WHERE c.ID = `+sqla.MakeParam(DBType, 1), c.ID)

	var ShortName sql.NullString
	var FullName sql.NullString
	var ForeignName sql.NullString
	var Contacts sql.NullString
	var CompanyHead sql.NullInt64
	var RegNo sql.NullString
	var TaxNo sql.NullString
	var BankDetails sql.NullString
	var CompanyHeadID sql.NullInt64
	var CompanyHeadFirstName sql.NullString
	var CompanyHeadSurname sql.NullString
	var CompanyHeadJobTitle sql.NullString

	err := row.Scan(&c.ID, &ShortName, &FullName, &ForeignName, &Contacts, &CompanyHead, &RegNo, &TaxNo, &BankDetails,
		&CompanyHeadID, &CompanyHeadFirstName, &CompanyHeadSurname, &CompanyHeadJobTitle)
	if err != nil {
		return err
	}
	c.ShortName = ShortName.String
	c.FullName = FullName.String
	c.ForeignName = ForeignName.String
	c.Contacts = unmarshalNonEmptyCompanyContacts(Contacts.String)
	if CompanyHeadID.Valid {
		c.CompanyHead = &Profile{
			ID:        int(CompanyHeadID.Int64),
			FirstName: CompanyHeadFirstName.String,
			Surname:   CompanyHeadSurname.String,
			JobTitle:  CompanyHeadJobTitle.String,
		}
	}
	c.RegNo = RegNo.String
	c.TaxNo = TaxNo.String
	c.BankDetails = BankDetails.String

	return nil
}

// GiveHeadID executes in template to provide CompanyHead ID
func (c Company) GiveHeadID() int {
	if c.CompanyHead == nil {
		return 0
	}
	return c.CompanyHead.ID
}

// GiveHeadNameJob executes in template to provide CompanyHead data
func (c Company) GiveHeadNameJob() (head string) {
	if c.CompanyHead == nil {
		return ""
	}
	head = strings.TrimSpace(c.CompanyHead.FirstName + " " + c.CompanyHead.Surname)
	if head == "" {
		head = "ID: " + strconv.Itoa(c.CompanyHead.ID)
	}
	if c.CompanyHead.JobTitle != "" {
		head += ", " + c.CompanyHead.JobTitle
	}
	return head
}
