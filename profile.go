package main

import (
	"database/sql"
	"edm/pkg/accs"
	"edm/pkg/datetime"
	"edm/pkg/memdb"
	"edm/pkg/passwd"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// GetID returns Profile ID to satisfy some interfaces
func (p Profile) GetID() int {
	return p.ID
}

func unmarshalToProfile(obj string) Profile {
	user := Profile{}
	json.Unmarshal([]byte(obj), &user)
	return user
}

func unmarshalNonEmptyProfileContacts(c string) (res UserContacts) {
	if c != "" {
		err := json.Unmarshal([]byte(c), &res)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
	}
	return res
}

func memoryUpdateProfile(db *sql.DB, dbType byte, m memdb.ObjectsInMemory, id int) {
	if m.IsObjectInMemory(id) {
		u := Profile{ID: id}
		err := u.loadByIDorLogin(db, dbType, "ID")
		if err != nil {
			log.Println("Critical memory update error:", err)
		}
		m.Update(u, (u.UserLock == 1 || u.Login == ""))
	}
}

func createFirstAdmin(db *sql.DB, DBType byte, langCode string) {
	admin := Profile{
		UserRole:  1,
		Login:     "admin",
		FirstName: "Administrator",
		Passwd:    "",
		UserConfig: UserConfig{
			SystemTheme:          "dark",
			ElemsOnPage:          20,
			ElemsOnPageTeam:      500,
			DateFormat:           "dd.mm.yyyy",
			TimeFormat:           "24h",
			LangCode:             langCode,
			UseCalendarInConrols: true,
			CurrencyBeforeAmount: true,
			ShowFinishedTasks:    true,
			ReturnAfterCreation:  true,
		},
	}
	uniqueLogin, _ := admin.isLoginUniqueorBlank(db, DBType)
	if uniqueLogin {
		admin.create(db, DBType)
	}
}

func (p *Profile) create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
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

func (p *Profile) updateConfig(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendJSONStruct("UserConfig", p.UserConfig)
	rowsaff = sqla.UpdateObject(db, DBType, "profiles", args, p.ID)
	return rowsaff
}

func (p *Profile) load(db *sql.DB, DBType byte) error {

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
	p.Contacts = unmarshalNonEmptyProfileContacts(Contacts.String)
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

func (p *Profile) preload(db *sql.DB, DBType byte) error {
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
	p.Contacts = unmarshalNonEmptyProfileContacts(Contacts.String)
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
	if DEBUG {
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

// ProfilePage is passed into template
type ProfilePage struct {
	AppTitle      string
	AppVersion    string
	PageTitle     string
	LoggedinID    int
	UserConfig    UserConfig
	Profile       Profile //payload
	Message       string
	LoggedinAdmin bool
	Editable      bool
	New           bool
	UserList      []memdb.ObjHasID
	UnitList      []memdb.ObjHasID
}

func (bs *BaseStruct) profileHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := bs.authVerify(w, r)
	if !allow {
		return
	}

	if bs.validURLs.Team.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = ProfilePage{
		AppTitle:      bs.text.AppTitle,
		AppVersion:    AppVersion,
		LoggedinID:    id,
		LoggedinAdmin: false,
		Editable:      false,
		New:           false,
	}

	user := unmarshalToProfile(bs.team.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == 1 {
		Page.LoggedinAdmin = true
	}

	TextID := accs.GetTextIDfromURL(r.URL.Path)
	IntID, _ := strconv.Atoi(TextID)
	if Page.LoggedinID == IntID || Page.LoggedinAdmin {
		Page.Editable = true
	}

	var created int
	var updated int
	var doupdate = false
	var updatedPasswd int
	var uniqueLogin bool

	// Create code =============================================
	if r.Method == "POST" && r.FormValue("createButton") != "" {
		if Page.LoggedinAdmin == false {
			accs.ThrowAccessDenied(w, "creating profile", Page.LoggedinID, IntID)
			return
		}
		p := Profile{
			FirstName: r.FormValue("firstName"),
			OtherName: r.FormValue("otherName"),
			Surname:   r.FormValue("surname"),
			Contacts: UserContacts{
				TelOffice: r.FormValue("telOffice"),
				TelMobile: r.FormValue("telMobile"),
				Email:     r.FormValue("email"),
				Other:     r.FormValue("otherContacts"),
			},
			BirthDate: datetime.StringToDate(r.FormValue("birthDate")),
			JobTitle:  r.FormValue("jobTitle"),
		}
		if r.FormValue("jobUnit") != "" && r.FormValue("jobUnit") != "0" {
			p.JobUnit = &Unit{ID: accs.StrToInt(r.FormValue("jobUnit"))}
		}
		if r.FormValue("boss") != "" && r.FormValue("boss") != "0" {
			p.Boss = &Profile{ID: accs.StrToInt(r.FormValue("boss"))}
		}
		p.UserConfig = UserConfig{
			SystemTheme:          "dark",
			ElemsOnPage:          20,
			ElemsOnPageTeam:      500,
			DateFormat:           "dd.mm.yyyy",
			TimeFormat:           "24h",
			LangCode:             bs.cfg.DefaultLang,
			UseCalendarInConrols: true,
			CurrencyBeforeAmount: true,
			ShowFinishedTasks:    true,
			ReturnAfterCreation:  true,
		}
		p.Login = r.FormValue("login")
		if r.FormValue("loginSameEmail") == "true" {
			p.Login = p.Contacts.Email
		}
		p.Passwd = passwd.GenPasswd(r.FormValue("passwd"))
		uniqueLogin, err = p.isLoginUniqueorBlank(bs.db, bs.dbt)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "checking unique login in DB", Page.LoggedinID, IntID)
			return
		}
		if uniqueLogin == false {
			Page.Message = "loginNotUnique"
			Page.Profile = p
		} else {
			p.ID, created = p.create(bs.db, bs.dbt)
			if created > 0 {
				if r.FormValue("notifyCreatedUser") == "true" {
					mail := AnyMail{bs.i18n.Messages.Subj.ProfileRegistered,
						bs.i18n.Messages.Cont.ProfileRegistered + p.Login + ", " + r.FormValue("passwd"),
						bs.systemURL, bs.i18n.Messages.DoNotReply, bs.systemURL, bs.i18n.Messages.MailerName}
					mail.constructToChannel(bs.db, bs.dbt, bs.anymailtmpl, bs.mailchan, p)
				}
				constructUserList(bs.db, bs.dbt, bs.team)
				if Page.UserConfig.ReturnAfterCreation {
					http.Redirect(w, r, "/team/", http.StatusSeeOther)
				} else {
					http.Redirect(w, r, fmt.Sprintf("/team/profile/%d", p.ID), http.StatusSeeOther)
				}
				return
			} else {
				Page.Message = "dataNotWritten"
			}
		}
	}

	// Update code =============================================
	if r.Method == "POST" && r.FormValue("updateButton") != "" {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "updating profile", Page.LoggedinID, IntID)
			return
		}
		p := Profile{ID: IntID,
			FirstName: r.FormValue("firstName"),
			OtherName: r.FormValue("otherName"),
			Surname:   r.FormValue("surname"),
			Contacts: UserContacts{
				TelOffice: r.FormValue("telOffice"),
				TelMobile: r.FormValue("telMobile"),
				Email:     r.FormValue("email"),
				Other:     r.FormValue("otherContacts"),
			},
			BirthDate: datetime.StringToDate(r.FormValue("birthDate")),
			JobTitle:  r.FormValue("jobTitle"),
		}
		if r.FormValue("jobUnit") != "" && r.FormValue("jobUnit") != "0" {
			p.JobUnit = &Unit{}
			p.JobUnit.ID, _ = strconv.Atoi(r.FormValue("jobUnit"))
		}
		if r.FormValue("boss") != "" && r.FormValue("boss") != "0" {
			p.Boss = &Profile{}
			p.Boss.ID, _ = strconv.Atoi(r.FormValue("boss"))
		}
		updated = p.update(bs.db, bs.dbt)
		doupdate = true
	}

	// Update lock =============================================
	if r.Method == "POST" && r.FormValue("updateLock") != "" {
		if Page.LoggedinAdmin == false {
			accs.ThrowAccessDenied(w, "updating UserLock", Page.LoggedinID, IntID)
			return
		}
		LastAdmin := false
		p := Profile{ID: IntID, UserLock: accs.StrToInt(r.FormValue("userLock"))}
		if p.UserLock == 1 {
			LastAdmin, err = p.isTheLastAdmin(bs.db, bs.dbt)
			if err != nil {
				log.Println(accs.CurrentFunction()+":", err)
			}
		}
		if LastAdmin {
			Page.Message = "lastAdminRejection"
		} else {
			updated = sqla.UpdateSingleInt(bs.db, bs.dbt, "profiles", "UserLock", p.UserLock, p.ID)
			doupdate = true
		}
	}

	// Update role =============================================
	if r.Method == "POST" && r.FormValue("updateRole") != "" {
		if Page.LoggedinAdmin == false {
			accs.ThrowAccessDenied(w, "updating UserRole", Page.LoggedinID, IntID)
			return
		}
		LastAdmin := false
		p := Profile{ID: IntID, UserRole: accs.StrToInt(r.FormValue("userRole"))}
		if p.UserRole == 0 {
			LastAdmin, err = p.isTheLastAdmin(bs.db, bs.dbt)
			if err != nil {
				log.Println(accs.CurrentFunction()+":", err)
			}
		}
		if LastAdmin {
			Page.Message = "lastAdminRejection"
		} else {
			updated = sqla.UpdateSingleInt(bs.db, bs.dbt, "profiles", "UserRole", p.UserRole, p.ID)
			doupdate = true
		}
	}

	// Common memory refresh code after updating user profile ==
	if doupdate && updated > 0 {
		Page.Message = "dataWritten"
		constructUserList(bs.db, bs.dbt, bs.team)
		memoryUpdateProfile(bs.db, bs.dbt, bs.team, IntID)
	} else if doupdate {
		Page.Message = "dataNotWritten"
	}

	// Update login and passwd =================================
	if r.Method == "POST" && r.FormValue("updatePasswd") != "" {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "updating passwd", Page.LoggedinID, IntID)
			return
		}
		p := Profile{ID: IntID}
		p.Login = r.FormValue("login")
		rawpasswd := r.FormValue("passwd")
		p.Passwd = passwd.GenPasswd(rawpasswd)
		uniqueLogin, err = p.isLoginUniqueorBlank(bs.db, bs.dbt)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "checking unique login in DB", Page.LoggedinID, IntID)
			return
		}
		LastAdmin := false
		if p.Login == "" {
			LastAdmin, err = p.isTheLastAdmin(bs.db, bs.dbt)
			if err != nil {
				log.Println(accs.CurrentFunction()+":", err)
			}
		}
		if !uniqueLogin {
			Page.Message = "loginNotUnique"
		} else if len(rawpasswd) < 6 {
			Page.Message = "passwdTooShort"
		} else if LastAdmin {
			Page.Message = "lastAdminRejection"
		} else {
			updatedPasswd = p.updatePasswd(bs.db, bs.dbt)
			if updatedPasswd > 0 {
				p.preload(bs.db, bs.dbt)
				mail := AnyMail{bs.i18n.Messages.Subj.SecurityAlert,
					bs.i18n.Messages.Cont.LoginPasswdChanged,
					bs.systemURL + "/profiles/" + strconv.Itoa(p.ID),
					bs.i18n.Messages.DoNotReply, bs.systemURL, bs.i18n.Messages.MailerName}
				mail.constructToChannel(bs.db, bs.dbt, bs.anymailtmpl, bs.mailchan, p)
				Page.Message = "dataWritten"
				memoryUpdateProfile(bs.db, bs.dbt, bs.team, p.ID)
			} else {
				Page.Message = "dataNotWritten"
			}
		}
	}

	// Loading code ============================================
	Page.UserList = bs.team.GetObjectArr("UserList")
	Page.UnitList = bs.team.GetObjectArr("UnitList")
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = bs.text.NewProfile
		if Page.Message == "" {
			Page.Message = "onlyAdminCanCreate"
		}
	} else {
		Page.Profile.ID = IntID
		err = Page.Profile.load(bs.db, bs.dbt)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			http.NotFound(w, r)
			return
		}
		if Page.Editable {
			Page.Profile.Passwd = "no access"
		} else {
			Page.Profile.Login = "no access"
			Page.Profile.Passwd = "no access"
			Page.Profile.BirthDate.Year = 0
			Page.Profile.UserConfig = UserConfig{}
		}
		Page.PageTitle = strings.TrimSpace(Page.Profile.FirstName + " " + Page.Profile.Surname)
		if Page.PageTitle == "" && Page.Editable {
			Page.PageTitle = Page.Profile.Login
		}
		if Page.PageTitle == "" {
			Page.PageTitle = bs.text.Profile + " ID: " + strconv.Itoa(Page.Profile.ID)
		}
	}

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		//jsonOut, _ := json.Marshal(Page)
		//fmt.Fprintln(w, string(jsonOut))
		return
	}

	// HTML output
	err = bs.templates.ExecuteTemplate(w, "profile.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		accs.ThrowServerError(w, "executing profile template", Page.LoggedinID, Page.Profile.ID)
		return
	}

}

// UserConfigPage is passed into template
type UserConfigPage struct {
	AppTitle    string
	AppVersion  string
	PageTitle   string
	LoggedinID  int
	Message     string
	UserConfig  UserConfig
	Themes      []string
	DateFormats []string
	TimeFormats []string
	LangCodes   []string
}

func (bs *BaseStruct) userConfigHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := bs.authVerify(w, r)
	if !allow {
		return
	}

	var Page = UserConfigPage{
		AppTitle:   bs.text.AppTitle,
		AppVersion: AppVersion,
		LoggedinID: id,
	}

	var err error
	var updated int

	// Update code =============================================
	if r.Method == "POST" && r.FormValue("updateButton") != "" {
		p := Profile{ID: Page.LoggedinID}
		p.UserConfig = UserConfig{
			SystemTheme:     r.FormValue("systemTheme"),
			ElemsOnPage:     accs.StrToInt(r.FormValue("elemsOnPage")),
			ElemsOnPageTeam: accs.StrToInt(r.FormValue("elemsOnPageTeam")),
			DateFormat:      r.FormValue("dateFormat"),
			TimeFormat:      r.FormValue("timeFormat"),
			LangCode:        r.FormValue("langCode"),
		}
		p.UserConfig.UseCalendarInConrols, _ = strconv.ParseBool(r.FormValue("useCalendarInConrols"))
		p.UserConfig.CurrencyBeforeAmount, _ = strconv.ParseBool(r.FormValue("currencyBeforeAmount"))
		p.UserConfig.ShowFinishedTasks, _ = strconv.ParseBool(r.FormValue("showFinishedTasks"))
		p.UserConfig.ReturnAfterCreation, _ = strconv.ParseBool(r.FormValue("returnAfterCreation"))
		updated = p.updateConfig(bs.db, bs.dbt)
		if updated > 0 {
			memoryUpdateProfile(bs.db, bs.dbt, bs.team, p.ID)
			Page.Message = "dataWritten"
		}
	}

	// Loading code ============================================
	Page.PageTitle = bs.text.ConfigPageTitle
	Page.Themes = bs.options.Themes
	Page.DateFormats = bs.options.DateFormats
	Page.TimeFormats = bs.options.TimeFormats
	Page.LangCodes = bs.options.LangCodes

	user := unmarshalToProfile(bs.team.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		//jsonOut, _ := json.Marshal(Page)
		//fmt.Fprintln(w, string(jsonOut))
		return
	}

	// HTML output
	err = bs.templates.ExecuteTemplate(w, "config.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		accs.ThrowServerError(w, "executing profile template", Page.LoggedinID, Page.LoggedinID)
		return
	}

}
