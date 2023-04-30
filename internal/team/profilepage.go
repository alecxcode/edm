package team

import (
	"edm/internal/core"
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

// ProfileHandler is http handler for profile page
func (tb *TeamBase) ProfileHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, tb.memorydb)
	if !allow {
		return
	}

	if tb.validURLs.team.FindStringSubmatch(r.URL.Path) == nil {
		accs.ThrowObjectNotFound(w, r)
		return
	}

	var err error

	var Page = ProfilePage{
		AppTitle:      tb.text.AppTitle,
		AppVersion:    core.AppVersion,
		LoggedinID:    id,
		LoggedinAdmin: false,
		Editable:      false,
		New:           false,
	}

	user := UnmarshalToProfile(tb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == ADMIN {
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
			SystemTheme:           "dark",
			ElemsOnPage:           20,
			ElemsOnPageTeam:       500,
			DateFormat:            "dd.mm.yyyy",
			TimeFormat:            "24h",
			LangCode:              tb.cfg.defaultLang,
			UseCalendarInControls: true,
			CurrencyBeforeAmount:  true,
			ShowFinishedTasks:     true,
			ReturnAfterCreation:   true,
		}
		p.Login = r.FormValue("login")
		if r.FormValue("loginSameEmail") == "true" {
			p.Login = p.Contacts.Email
		}
		p.Passwd = passwd.GenPasswd(r.FormValue("passwd"))
		uniqueLogin, err = p.isLoginUniqueorBlank(tb.db, tb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "checking unique login in DB", Page.LoggedinID, IntID)
			return
		}
		if uniqueLogin == false {
			Page.Message = "loginNotUnique"
			Page.Profile = p
		} else {
			p.ID, created = p.Create(tb.db, tb.dbType)
			if created > 0 {
				if r.FormValue("notifyCreatedUser") == "true" {
					mail := core.AnyMail{Subj: tb.i18n.messages.Subj.ProfileRegistered,
						Text:     tb.i18n.messages.Cont.ProfileRegistered + p.Login + ", " + r.FormValue("passwd"),
						SomeLink: tb.cfg.systemURL, DoNotReply: tb.i18n.messages.DoNotReply, SystemURL: tb.cfg.systemURL, MailerName: tb.i18n.messages.MailerName}
					mail.ConstructToChannel(tb.db, tb.dbType, tb.mailtmpl, tb.mailchan, p)
				}
				core.ConstructUserList(tb.db, tb.dbType, tb.memorydb)
				if Page.UserConfig.ReturnAfterCreation {
					http.Redirect(w, r, "/team/"+core.IfAddJSON(r), http.StatusSeeOther)
				} else {
					http.Redirect(w, r, fmt.Sprintf("/team/profile/%d"+core.IfAddJSON(r), p.ID), http.StatusSeeOther)
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
			p.JobUnit = &Unit{ID: accs.StrToInt(r.FormValue("jobUnit"))}
		}
		if r.FormValue("boss") != "" && r.FormValue("boss") != "0" {
			p.Boss = &Profile{ID: accs.StrToInt(r.FormValue("boss"))}
		}
		updated = p.update(tb.db, tb.dbType)
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
			LastAdmin, err = p.isTheLastAdmin(tb.db, tb.dbType)
			if err != nil {
				log.Println(accs.CurrentFunction()+":", err)
			}
		}
		if LastAdmin {
			Page.Message = "lastAdminRejection"
		} else {
			updated = sqla.UpdateSingleInt(tb.db, tb.dbType, "profiles", "UserLock", p.UserLock, p.ID)
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
		if p.UserRole == NOROLE {
			LastAdmin, err = p.isTheLastAdmin(tb.db, tb.dbType)
			if err != nil {
				log.Println(accs.CurrentFunction()+":", err)
			}
		}
		if LastAdmin {
			Page.Message = "lastAdminRejection"
		} else {
			updated = sqla.UpdateSingleInt(tb.db, tb.dbType, "profiles", "UserRole", p.UserRole, p.ID)
			doupdate = true
		}
	}

	// Common memory refresh code after updating user profile ==
	if doupdate && updated > 0 {
		Page.Message = "dataWritten"
		core.ConstructUserList(tb.db, tb.dbType, tb.memorydb)
		MemoryUpdateProfile(tb.db, tb.dbType, tb.memorydb, IntID)
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
		uniqueLogin, err = p.isLoginUniqueorBlank(tb.db, tb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "checking unique login in DB", Page.LoggedinID, IntID)
			return
		}
		LastAdmin := false
		if p.Login == "" {
			LastAdmin, err = p.isTheLastAdmin(tb.db, tb.dbType)
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
			updatedPasswd = p.updatePasswd(tb.db, tb.dbType)
			if updatedPasswd > 0 {
				p.Preload(tb.db, tb.dbType)
				mail := core.AnyMail{Subj: tb.i18n.messages.Subj.SecurityAlert,
					Text:       tb.i18n.messages.Cont.LoginPasswdChanged,
					SomeLink:   tb.cfg.systemURL + "/profiles/" + strconv.Itoa(p.ID),
					DoNotReply: tb.i18n.messages.DoNotReply, SystemURL: tb.cfg.systemURL,
					MailerName: tb.i18n.messages.MailerName}
				mail.ConstructToChannel(tb.db, tb.dbType, tb.mailtmpl, tb.mailchan, p)
				Page.Message = "dataWritten"
				MemoryUpdateProfile(tb.db, tb.dbType, tb.memorydb, p.ID)
			} else {
				Page.Message = "dataNotWritten"
			}
		}
	}

	// Loading code ============================================
	Page.UserList = tb.memorydb.GetObjectArr("UserList")
	Page.UnitList = tb.memorydb.GetObjectArr("UnitList")
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = tb.text.NewProfile
		if Page.Message == "" {
			Page.Message = "onlyAdminCanCreate"
		}
	} else {
		Page.Profile.ID = IntID
		err = Page.Profile.Load(tb.db, tb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowObjectNotFound(w, r)
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
			Page.PageTitle = tb.text.Profile + " ID: " + strconv.Itoa(Page.Profile.ID)
		}
	}

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		return
	}

	// HTML output
	err = tb.templates.ExecuteTemplate(w, "profile.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		accs.ThrowServerError(w, "executing profile template", Page.LoggedinID, Page.Profile.ID)
		return
	}

}
