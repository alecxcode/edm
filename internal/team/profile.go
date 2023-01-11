package team

import (
	"database/sql"
	"edm/internal/core"
	"edm/pkg/accs"
	"edm/pkg/datetime"
	"edm/pkg/memdb"
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/alecxcode/sqla"
)

// Profile is a stuff member
type Profile struct {
	//sql generate
	ID         int
	FirstName  string        `sql-gen:"varchar(255)"`
	OtherName  string        `sql-gen:"varchar(255)"`
	Surname    string        `sql-gen:"varchar(255)"`
	BirthDate  datetime.Date `sql-gen:"bigint"`
	JobTitle   string        `sql-gen:"varchar(255)"`
	JobUnit    *Unit         `sql-gen:"FK_NULL"`
	Boss       *Profile      `sql-gen:"FK_NULL,FK_NOACTION"`
	Contacts   UserContacts  `sql-gen:"varchar(4000)"`
	UserRole   int           // admin = 1, others = 0
	UserLock   int           // locked = 1, unlocked = 0
	UserConfig UserConfig    `sql-gen:"varchar(4000)"`
	Login      string        `sql-gen:"varchar(255)"`
	Passwd     string        `sql-gen:"varchar(255)"`
}

// User roles consts
const (
	NOROLE = 0
	ADMIN  = 1
)

// UserConfig stores person-related settings; it should not include data for query filters
type UserConfig struct {
	SystemTheme          string
	ElemsOnPage          int
	ElemsOnPageTeam      int
	DateFormat           string
	TimeFormat           string
	LangCode             string
	UseCalendarInConrols bool
	CurrencyBeforeAmount bool
	ShowFinishedTasks    bool
	ReturnAfterCreation  bool
}

// UserContacts contains user contact data
type UserContacts struct {
	TelOffice string
	TelMobile string
	Email     string
	Other     string
}

// UnmarshalToProfile parses JSON string and returns Profile
func UnmarshalToProfile(obj string) Profile {
	user := Profile{}
	json.Unmarshal([]byte(obj), &user)
	return user
}

// UnmarshalNonEmptyProfileContacts  parses JSON string and returns UserContacts
func UnmarshalNonEmptyProfileContacts(c string) (res UserContacts) {
	if c != "" {
		err := json.Unmarshal([]byte(c), &res)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
	}
	return res
}

// MemoryUpdateProfile saves profile to memory store
func MemoryUpdateProfile(db *sql.DB, dbType byte, m memdb.ObjectsInMemory, id int) {
	if m.IsObjectInMemory(id) {
		u := Profile{ID: id}
		err := u.loadByIDorLogin(db, dbType, "ID")
		if err != nil {
			log.Println("Critical memory update error:", err)
		}
		m.Update(u, (u.UserLock == 1 || u.Login == ""))
	}
}

// Create saves to DB
func (p *Profile) Create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendNonEmptyString("FirstName", p.FirstName)
	args = args.AppendNonEmptyString("OtherName", p.OtherName)
	args = args.AppendNonEmptyString("Surname", p.Surname)
	args = args.AppendJSONStruct("Contacts", p.Contacts)
	if p.BirthDate.Day != 0 {
		args = args.AppendInt64("BirthDate", datetime.DateToInt64(p.BirthDate))
	}
	args = args.AppendNonEmptyString("JobTitle", p.JobTitle)
	if p.JobUnit != nil {
		args = args.AppendInt("JobUnit", p.JobUnit.ID)
	}
	if p.Boss != nil {
		args = args.AppendInt("Boss", p.Boss.ID)
	}
	args = args.AppendInt("UserRole", p.UserRole)
	args = args.AppendInt("UserLock", p.UserLock)
	args = args.AppendJSONStruct("UserConfig", p.UserConfig)
	args = args.AppendNonEmptyString("Login", p.Login)
	args = args.AppendNonEmptyString("Passwd", p.Passwd)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "profiles", args)
	return lastid, rowsaff
}

func (p *Profile) update(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendStringOrNil("FirstName", p.FirstName)
	args = args.AppendStringOrNil("OtherName", p.OtherName)
	args = args.AppendStringOrNil("Surname", p.Surname)
	args = args.AppendJSONStruct("Contacts", p.Contacts)
	if p.BirthDate.Day != 0 {
		args = args.AppendInt64("BirthDate", datetime.DateToInt64(p.BirthDate))
	} else {
		args = args.AppendNil("BirthDate")
	}
	args = args.AppendStringOrNil("JobTitle", p.JobTitle)
	if p.JobUnit != nil {
		args = args.AppendInt("JobUnit", p.JobUnit.ID)
	} else {
		args = args.AppendNil("JobUnit")
	}
	if p.Boss != nil {
		args = args.AppendInt("Boss", p.Boss.ID)
	} else {
		args = args.AppendNil("Boss")
	}
	rowsaff = sqla.UpdateObject(db, DBType, "profiles", args, p.ID)
	return rowsaff
}

func (p *Profile) updatePasswd(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendStringOrNil("Login", p.Login)
	args = args.AppendStringOrNil("Passwd", p.Passwd)
	rowsaff = sqla.UpdateObject(db, DBType, "profiles", args, p.ID)
	return rowsaff
}

// UpdateConfig updates profile UserConfig in memory
func (p *Profile) UpdateConfig(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendJSONStruct("UserConfig", p.UserConfig)
	rowsaff = sqla.UpdateObject(db, DBType, "profiles", args, p.ID)
	return rowsaff
}

// Load loads a Profile from DB
func (p *Profile) Load(db *sql.DB, DBType byte) error {

	row := db.QueryRow(`SELECT
p.ID, p.FirstName, p.OtherName, p.Surname, p.Contacts, p.BirthDate, p.JobTitle, p.JobUnit, p.Boss, p.UserRole, p.UserLock, p.UserConfig, p.Login,
units.ID, units.Company, units.UnitName,
companies.ID, companies.ShortName,
b.ID, b.FirstName, b.Surname, b.JobTitle
FROM profiles p
LEFT JOIN units ON units.ID = p.JobUnit
LEFT JOIN companies ON companies.ID = units.Company
LEFT JOIN profiles b ON b.ID = p.Boss
WHERE p.ID = `+sqla.MakeParam(DBType, 1), p.ID)

	var FirstName sql.NullString
	var OtherName sql.NullString
	var Surname sql.NullString
	var Contacts sql.NullString
	var BirthDate sql.NullInt64
	var JobTitle sql.NullString
	var JobUnit sql.NullInt64
	var Boss sql.NullInt64
	var UserRole sql.NullInt64
	var UserLock sql.NullInt64
	var UserConfig sql.NullString
	var Login sql.NullString

	var UnitID sql.NullInt64
	var UnitCompany sql.NullInt64
	var UnitName sql.NullString

	var CompanyID sql.NullInt64
	var ShortName sql.NullString

	var BossID sql.NullInt64
	var BossFirstName sql.NullString
	var BossSurname sql.NullString
	var BossJobTitle sql.NullString

	err := row.Scan(&p.ID, &FirstName, &OtherName, &Surname, &Contacts, &BirthDate, &JobTitle, &JobUnit, &Boss, &UserRole, &UserLock, &UserConfig, &Login,
		&UnitID, &UnitCompany, &UnitName,
		&CompanyID, &ShortName,
		&BossID, &BossFirstName, &BossSurname, &BossJobTitle)
	if err != nil {
		return err
	}

	p.FirstName = FirstName.String
	p.OtherName = OtherName.String
	p.Surname = Surname.String
	p.Contacts = UnmarshalNonEmptyProfileContacts(Contacts.String)
	p.BirthDate = datetime.GetValidDateFromSQL(BirthDate)
	p.JobTitle = JobTitle.String
	if UnitID.Valid {
		p.JobUnit = &Unit{
			ID: int(UnitID.Int64),
			Company: &Company{
				ID:        int(CompanyID.Int64),
				ShortName: ShortName.String,
			},
			UnitName: UnitName.String,
		}
	}
	if BossID.Valid {
		p.Boss = &Profile{
			ID:        int(BossID.Int64),
			FirstName: BossFirstName.String,
			Surname:   BossSurname.String,
			JobTitle:  BossJobTitle.String,
		}
	}
	p.UserRole = int(UserRole.Int64)
	p.UserLock = int(UserLock.Int64)
	if UserConfig.String != "" {
		err := json.Unmarshal([]byte(UserConfig.String), &p.UserConfig)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
	}
	p.Login = Login.String

	return nil
}

// Preload loads incomplete profile from DB
func (p *Profile) Preload(db *sql.DB, DBType byte) error {
	row := db.QueryRow(`SELECT ID, FirstName, Surname, JobTitle, Contacts, UserLock FROM profiles WHERE ID = `+sqla.MakeParam(DBType, 1), p.ID)
	var FirstName sql.NullString
	var Surname sql.NullString
	var JobTitle sql.NullString
	var Contacts sql.NullString
	var UserLock sql.NullInt64
	err := row.Scan(&p.ID, &FirstName, &Surname, &JobTitle, &Contacts, &UserLock)
	if err != nil {
		return err
	}
	p.FirstName = FirstName.String
	p.Surname = Surname.String
	p.JobTitle = JobTitle.String
	p.Contacts = UnmarshalNonEmptyProfileContacts(Contacts.String)
	p.UserLock = int(UserLock.Int64)
	return nil
}

func (p *Profile) loadByIDorLogin(db *sql.DB, DBType byte, what string) (err error) {
	var row *sql.Row
	if what == "ID" {
		sq := `SELECT ID, FirstName, OtherName, Surname, UserRole, UserLock, UserConfig, Login, Passwd FROM profiles WHERE ID = `
		row = db.QueryRow(sq+sqla.MakeParam(DBType, 1), p.ID)
	} else if what == "Login" {
		sq := `SELECT ID, FirstName, OtherName, Surname, UserRole, UserLock, UserConfig, Login, Passwd FROM profiles WHERE Login = `
		row = db.QueryRow(sq+sqla.MakeParam(DBType, 1), p.Login)
	}
	p.ID = 0     // clearing out for security reasons
	p.Login = "" // clearing out for security reasons
	var FirstName sql.NullString
	var OtherName sql.NullString
	var Surname sql.NullString
	var UserRole sql.NullInt64
	var UserLock sql.NullInt64
	var UserConfig sql.NullString
	var Login sql.NullString
	var Passwd sql.NullString
	err = row.Scan(&p.ID, &FirstName, &OtherName, &Surname, &UserRole, &UserLock, &UserConfig, &Login, &Passwd)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	p.FirstName = FirstName.String
	p.OtherName = OtherName.String
	p.Surname = Surname.String
	p.UserRole = int(UserRole.Int64)
	p.UserLock = int(UserLock.Int64)
	if UserConfig.String != "" {
		err := json.Unmarshal([]byte(UserConfig.String), &p.UserConfig)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
	}
	p.UserConfig.correctIfEmpty()
	p.Login = Login.String
	p.Passwd = Passwd.String
	return nil
}

func (uc *UserConfig) correctIfEmpty() {
	if uc.SystemTheme == "" {
		uc.SystemTheme = "dark"
	}
	if uc.LangCode == "" {
		uc.LangCode = "en"
	}
	if uc.ElemsOnPage == 0 {
		uc.ElemsOnPage = 20
	}
	if uc.ElemsOnPageTeam == 0 {
		uc.ElemsOnPageTeam = 500
	}
}

func (p *Profile) isLoginUniqueorBlank(db *sql.DB, DBType byte) (res bool, err error) {
	if p.Login == "" {
		return true, nil
	}
	if core.DEBUG {
		log.Println("SELECT Login FROM profiles WHERE ID <> "+sqla.MakeParam(DBType, 1)+" AND Login = "+sqla.MakeParam(DBType, 2), p.ID, p.Login)
	}
	rows, err := db.Query("SELECT Login FROM profiles WHERE ID <> "+
		sqla.MakeParam(DBType, 1)+" AND Login = "+sqla.MakeParam(DBType, 2), p.ID, p.Login)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	var Login sql.NullString
	for rows.Next() {
		err = rows.Scan(&Login)
		if err != nil {
			return false, err
		}
		if Login.String == p.Login {
			return false, nil
		}
	}
	return true, nil
}

func (p *Profile) isTheLastAdmin(db *sql.DB, DBType byte) (res bool, err error) {
	const NOROLE = 0
	const ADMIN = 1
	const LOCKED = 1
	row := db.QueryRow("SELECT COUNT(*) FROM profiles WHERE ID <> "+
		sqla.MakeParam(DBType, 1)+" AND UserRole = "+sqla.MakeParam(DBType, 2)+
		" AND UserLock <> "+sqla.MakeParam(DBType, 3)+" AND Login IS NOT NULL", p.ID, ADMIN, LOCKED)
	var counted sql.NullInt64
	err = row.Scan(&counted)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		return false, nil
	}
	AdminsRemains := int(counted.Int64)
	if AdminsRemains < 1 {
		return true, nil
	} else {
		return false, nil
	}
}

// GiveUnitID runs in a template to give unit ID
func (p Profile) GiveUnitID() int {
	if p.JobUnit == nil {
		return 0
	}
	return p.JobUnit.ID
}

// GiveUnitName runs in a template to give unit name
func (p Profile) GiveUnitName() (uname string) {
	if p.JobUnit == nil {
		return ""
	}
	if p.JobUnit.UnitName == "" {
		return "ID: " + strconv.Itoa(p.JobUnit.ID)
	}
	uname += p.JobUnit.UnitName
	if p.JobUnit.Company != nil {
		if p.JobUnit.Company.ShortName != "" {
			uname += ", " + p.JobUnit.Company.ShortName
		}
	}
	return uname
}

// GiveBossID runs in a template to give boss ID
func (p Profile) GiveBossID() int {
	if p.Boss == nil {
		return 0
	}
	return p.Boss.ID
}

// GiveBossNameJob runs in a template to give boss data
func (p Profile) GiveBossNameJob() (head string) {
	if p.Boss == nil {
		return ""
	}
	head = strings.TrimSpace(p.Boss.FirstName + " " + p.Boss.Surname)
	if head == "" {
		head = "ID: " + strconv.Itoa(p.Boss.ID)
	}
	if p.Boss.JobTitle != "" {
		head += ", " + p.Boss.JobTitle
	}
	return head
}

// GiveSelfNameJob runs in a template to give data of the user
func (p Profile) GiveSelfNameJob() (n string) {
	n = strings.TrimSpace(p.FirstName + " " + p.Surname)
	if n == "" {
		n = "ID: " + strconv.Itoa(p.ID)
	}
	if p.JobTitle != "" {
		n += ", " + p.JobTitle
	}
	return n
}

// GiveBirthDay executes in a template to deliver the queried date
func (p Profile) GiveBirthDay(dateFmt string) string {
	switch dateFmt {
	case "yyyy-mm-dd":
		return datetime.DateToStringWOY(p.BirthDate, true)
	case "yyyy.mm.dd":
		return datetime.DateToStringWOY(p.BirthDate, true)
	case "dd.mm.yyyy":
		return datetime.DateToStringWOY(p.BirthDate, false)
	case "dd/mm/yyyy":
		return datetime.DateToStringWOY(p.BirthDate, false)
	case "Mon dd, yyyy":
		return datetime.DateToStringWOY(p.BirthDate, true)
	case "mm/dd/yyyy":
		return datetime.DateToStringWOY(p.BirthDate, true)
	default:
		return datetime.DateToStringWOY(p.BirthDate, false)
	}
}

// GiveBirthDate executes in a template to deliver the queried date
func (p Profile) GiveBirthDate() string {
	return datetime.DateToString(p.BirthDate, "yyyy-mm-dd")
}
