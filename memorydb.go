package main

import (
	"database/sql"
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"edm/pkg/passwd"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
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

func constructUserList(db *sql.DB, DBType byte, m memdb.ObjectsInMemory) error {
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

func constructUnitList(db *sql.DB, DBType byte, m memdb.ObjectsInMemory) error {
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

func constructCorpList(db *sql.DB, DBType byte, m memdb.ObjectsInMemory) error {
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

func getLoginTemplate() *template.Template {
	logintmpl := `<!DOCTYPE html>
<head><meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="ie=edge">
<title>{{.AppTitle}}: {{.LoginPageTitle}}</title>
<link rel="shortcut icon" href="/static/favicon.png">
<link rel="icon" href="/static/favicon.png">
<link rel="stylesheet" href="/static/fonts.css">
<link rel="stylesheet" href="/static/system-dark.css">
</head>
<body><div id="container"><div id="control">
<h1>{{.LoginPageTitle}}</h1><br>
<form id="loginForm" action="/login" method="POST">
<p>{{.LoninPrompt}}</p><br><p class="msgred">{{.WrongLoginMsg}}</p>
<div><label for="loginName" style="display: inline-block; width: 80px;">{{.LoninFieldLabel}}</label> <input id="loginName" type="text" size="15" class="field" name="loginName" placeholder="" value=""></div>
<div><label for="loginPasswd" style="display: inline-block; width: 80px;">{{.PasswordFieldLabel}}</label> <input id="loginPasswd" type="password" size="15" class="field" name="loginPasswd" placeholder="" value=""></div>
<br><input type="submit" class ="sbut" name="loginButton" value="{{.LoginButton}}"><br>
</form></div>
<div id="bottom">
<span>© 2022 <a href="https://github.com/alecxcode/edm" target="_blank">EDM Project</a></span>
<span>v` + AppVersion + `.</span>
<span><a href="/static/manual.html">Manual</a></span>
</div>
</div></body></html>`
	return template.Must(template.New("login").Parse(logintmpl))
}

// LoginLang applies to a login template
type LoginLang struct {
	AppTitle           string
	LoginPageTitle     string
	LoninPrompt        string
	LoninFieldLabel    string
	PasswordFieldLabel string
	LoginButton        string
	WrongLoginMsg      string
}

func writeLoginPage(w http.ResponseWriter, LoginLang LoginLang, LoginTemplate *template.Template, wrongLogin bool) {
	if wrongLogin == false {
		LoginLang.WrongLoginMsg = ""
	}
	err := LoginTemplate.ExecuteTemplate(w, "login", LoginLang)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (bs *BaseStruct) getLoginLang() LoginLang {
	return bs.i18n.LoginLang
}

func (bs *BaseStruct) authVerify(w http.ResponseWriter, r *http.Request) (res bool, id int) {
	thecookie, err := r.Cookie("sessionid")
	if err == http.ErrNoCookie {
		if !bs.team.VerifyBruteForceCounter() {
			writeBruteForceStub(w)
			return false, 0
		}
		writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, false)
		return false, 0
	}
	allow, id := bs.team.CheckSession(thecookie.Value)
	if !allow {
		if !bs.team.VerifyBruteForceCounter() {
			writeBruteForceStub(w)
			return false, 0
		}
		writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, false)
		return false, 0
	}
	return true, id
}

func (bs *BaseStruct) loginHandler(w http.ResponseWriter, r *http.Request) {
	if !bs.team.VerifyBruteForceCounter() {
		writeBruteForceStub(w)
		return
	}
	if r.Method == "POST" && r.FormValue("loginName") != "" {
		var err error
		err = nil
		user := Profile{Login: r.FormValue("loginName")}
		err = user.loadByIDorLogin(bs.db, bs.dbt, "Login")
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
		if err == nil && user.Login != "" && user.UserLock == 0 && r.FormValue("loginName") == user.Login &&
			passwd.ComparePasswd(user.Passwd, r.FormValue("loginPasswd")) {
			cookval, _ := uuid.NewV4()
			var cookie = http.Cookie{
				Name:     "sessionid",
				Value:    cookval.String(),
				HttpOnly: true,
				Expires:  time.Now().AddDate(0, 0, 38000),
			}
			bs.team.Set(cookie.Value, user)
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusFound)
			bs.team.ResetBruteForceCounterImmediately()
		} else {
			bs.team.IncreaseBruteForceCounter(accs.GetIPAddr(r), r.FormValue("loginName"))
			writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, true)
			return
		}
	} else if r.Method == "POST" && r.FormValue("loginName") == "" {
		writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, true)
		return
	} else {
		writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, false)
		return
	}
}

func (bs *BaseStruct) logoutHandler(w http.ResponseWriter, r *http.Request) {
	thecookie, err := r.Cookie("sessionid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	bs.team.DelCookie(thecookie.Value)
	thecookie.Expires = time.Now().AddDate(-1, -1, -1)
	http.SetCookie(w, thecookie)
	http.Redirect(w, r, "/", http.StatusFound)
	writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, false)
	return
}

func writeBruteForceStub(w http.ResponseWriter) {
	fmt.Fprintf(w, `<!DOCTYPE html>
	<head><meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>EDM: System Bruteforce Attack Shield</title>
	<link rel="shortcut icon" href="/static/favicon.png">
	<link rel="icon" href="/static/favicon.png">
	<link rel="stylesheet" href="/static/fonts.css">
	<link rel="stylesheet" href="/static/system-dark.css">
	</head>
	<body><div id="container"><div id="control"><h1>System Bruteforce Attack Shield</h1>
	<br><p class="msgredfx">System login function is temporarily locked due to bruteforce attack. Normally, this should not happen. Please, contact system admin.</p><br><br><br>
	</div></div></body></html>`)
}
