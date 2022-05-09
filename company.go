package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
			log.Println(currentFunction()+":", err)
		}
	}
	return res
}

func (c *Company) create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
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

// Unit is a company department
type Unit struct {
	//sql generate
	ID       int
	UnitName string   `sql-gen:"varchar(1024)"`
	Company  *Company `sql-gen:"FK_CASCADE"`
	UnitHead *Profile `sql-gen:"FK_NULL"`
}

func (u *Unit) create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
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

// CompanyPage is passed into template
type CompanyPage struct {
	AppTitle   string
	PageTitle  string
	LoggedinID int
	UserConfig UserConfig
	Company    Company //payload
	Units      []Unit  //payload+
	Message    string
	Editable   bool
	New        bool
	UserList   []UserListElem
}

func (bs *BaseStruct) companyHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := bs.authVerify(w, r)
	if !allow {
		return
	}

	if bs.validURLs.Comp.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = CompanyPage{
		AppTitle:   bs.lng.AppTitle,
		LoggedinID: id,
		Editable:   false,
		New:        false,
	}

	user := bs.team.getByID(Page.LoggedinID)
	Page.UserConfig = user.UserConfig
	if user.UserRole == 1 {
		Page.Editable = true
	}

	TextID := getTextIDfromURL(r.URL.Path)
	IntID, _ := strconv.Atoi(TextID)

	var created int
	var updated int

	// Create or update code ==========================================================================
	if r.Method == "POST" && (r.FormValue("createButton") != "" || r.FormValue("updateButton") != "") {
		if Page.Editable == false {
			throwAccessDenied(w, "writing company", Page.LoggedinID, IntID)
			return
		}
		c := Company{
			ID:          IntID,
			ShortName:   r.FormValue("shortName"),
			FullName:    r.FormValue("fullName"),
			ForeignName: r.FormValue("foreignName"),
			Contacts: CompanyContacts{
				AddressReg:  r.FormValue("addressReg"),
				AddressFact: r.FormValue("addressFact"),
				Phone:       r.FormValue("phone"),
				Email:       r.FormValue("email"),
				WebSite:     r.FormValue("website"),
				Other:       r.FormValue("otherContacts"),
			},
			RegNo:       r.FormValue("regNo"),
			TaxNo:       r.FormValue("taxNo"),
			BankDetails: r.FormValue("bankDetails"),
		}
		if r.FormValue("companyHead") != "" && r.FormValue("companyHead") != "0" {
			c.CompanyHead = &Profile{}
			c.CompanyHead.ID, _ = strconv.Atoi(r.FormValue("companyHead"))
		}

		if r.FormValue("createButton") != "" {
			c.ID, created = c.create(bs.db, bs.dbt)
			if created > 0 {
				bs.team.constructCorpList(bs.db, bs.dbt)
				http.Redirect(w, r, fmt.Sprintf("/companies/company/%d", c.ID), http.StatusSeeOther)
				return
			} else {
				Page.Message = "dataNotWritten"
			}
		}

		if r.FormValue("updateButton") != "" {
			updated = c.update(bs.db, bs.dbt)
			if updated > 0 {
				bs.team.constructCorpList(bs.db, bs.dbt)
				bs.team.constructUnitList(bs.db, bs.dbt)
				Page.Message = "dataWritten"
			} else {
				Page.Message = "dataNotWritten"
			}
		}

	}

	// Create or update Units =====================================================================
	if r.Method == "POST" && (r.FormValue("createUnit") != "" || r.FormValue("updateUnit") != "") {
		if Page.Editable == false {
			throwAccessDenied(w, "writing unit", Page.LoggedinID, IntID)
			return
		}
		UnitIntID, _ := strconv.Atoi(r.FormValue("unitID"))
		u := Unit{
			ID:       UnitIntID,
			UnitName: r.FormValue("unitName"),
			Company:  &Company{ID: IntID},
		}
		if r.FormValue("unitHead") != "" && r.FormValue("unitHead") != "0" {
			u.UnitHead = &Profile{}
			u.UnitHead.ID, _ = strconv.Atoi(r.FormValue("unitHead"))
		}
		var unitaff int
		if r.FormValue("createUnit") != "" {
			_, unitaff = u.create(bs.db, bs.dbt)
		}
		if r.FormValue("updateUnit") != "" {
			unitaff = u.update(bs.db, bs.dbt)
		}
		if unitaff > 0 {
			bs.team.constructUnitList(bs.db, bs.dbt)
			Page.Message = "dataWritten"
		} else {
			Page.Message = "dataNotWritten"
		}
	}

	// Delete Unit ================+==========================
	if r.Method == "POST" && r.FormValue("deleteUnit") != "" {
		if Page.Editable == false {
			throwAccessDenied(w, "writing unit", Page.LoggedinID, IntID)
			return
		}
		UnitIntID, _ := strconv.Atoi(r.FormValue("unitID"))
		unitaff := sqla.DeleteObjects(bs.db, bs.dbt, "units", "ID", []int{UnitIntID})
		if unitaff > 0 {
			bs.team.constructUnitList(bs.db, bs.dbt)
			Page.Message = "dataWritten"
		} else {
			Page.Message = "dataNotWritten"
		}
	}

	// Loading code ============================================
	Page.UserList = bs.team.returnUserList()
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = bs.lng.NewCompany
		if Page.Message == "" {
			Page.Message = "onlyAdminCanCreate"
		}
	} else {
		Page.Company.ID = IntID
		err = Page.Company.load(bs.db, bs.dbt)
		if err != nil {
			log.Println(currentFunction()+":", err)
			http.NotFound(w, r)
			return
		}
		Page.Units, err = Page.Company.loadUnits(bs.db, bs.dbt)
		if err != nil {
			log.Println(currentFunction()+":", err)
			http.NotFound(w, r)
			return
		}
		Page.PageTitle = Page.Company.ShortName
		if Page.PageTitle == "" {
			Page.PageTitle = Page.Company.FullName
		}
		if Page.PageTitle == "" {
			Page.PageTitle = Page.Company.ForeignName
		}
		if Page.PageTitle == "" {
			Page.PageTitle = bs.lng.Company + " ID: " + strconv.Itoa(Page.Company.ID)
		}
	}

	// JSON output
	if r.URL.Query().Get("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		//jsonOut, _ := json.Marshal(Page)
		//fmt.Fprintln(w, string(jsonOut))
		return
	}

	// HTML output
	err = bs.templates.ExecuteTemplate(w, "company.tmpl", Page)
	if err != nil {
		log.Println(currentFunction()+":", err)
		throwServerError(w, "executing company template", Page.LoggedinID, Page.Company.ID)
		return
	}

}
