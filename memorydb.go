package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

// ALLOWED_LOGIN_ATTEMPTS is a bruteforce shield related constant
const ALLOWED_LOGIN_ATTEMPTS = 100

// MINUTES_TO_WAIT_BRUTEFORCE_UNLOCK is a bruteforce shield related constant
const MINUTES_TO_WAIT_BRUTEFORCE_UNLOCK = 60

// ProfilesInMemory holds all users in a single memory map with mutex to prevent data race
type ProfilesInMemory struct {
	sync.RWMutex
	Aarr              map[int]Profile
	Cjar              map[string]int
	UserList          []UserListElem
	UnitList          []UnitListElem
	CorpList          []CorpListElem
	BruteForceCounter int
}

// UserListElem is an element for buildibg user lists
type UserListElem struct {
	ID          int
	FullNameJob string
}

// UnitListElem is an element for buildibg unit lists
type UnitListElem struct {
	ID           int
	UnitNameComp string
}

// CorpListElem is an element for buildibg company lists
type CorpListElem struct {
	ID          int
	CompanyName string
}

func (m *ProfilesInMemory) isProfileInMemory(id int) (res bool) {
	res = false
	m.RLock()
	if _, ok := m.Aarr[id]; ok {
		res = true
	}
	m.RUnlock()
	return res
}

func (m *ProfilesInMemory) delCookie(cookval string) {
	var idForClearCheck int
	m.Lock()
	idForClearCheck = m.Cjar[cookval]
	delete(m.Cjar, cookval)
	m.Unlock()
	// Below code removes completely-logout user
	var remove = true
	m.RLock()
	for _, fid := range m.Cjar {
		if fid == idForClearCheck {
			remove = false
			break
		}
	}
	m.RUnlock()
	if remove {
		m.Lock()
		delete(m.Aarr, idForClearCheck)
		m.Unlock()
	}
}

func (m *ProfilesInMemory) getByID(id int) Profile {
	m.RLock()
	elem := m.Aarr[id]
	m.RUnlock()
	return elem
}

func (m *ProfilesInMemory) getByLogin(login string) Profile {
	var elem Profile
	m.RLock()
	for _, p := range m.Aarr {
		if p.Login == login {
			elem = p
			break
		}
	}
	m.RUnlock()
	return elem
}

func (m *ProfilesInMemory) set(cookval string, user Profile) {
	m.Lock()
	m.Cjar[cookval] = user.ID
	m.Aarr[user.ID] = user
	m.Unlock()
}

func (m *ProfilesInMemory) update(user Profile) {
	if user.UserLock == 1 || user.Login == "" {
		m.delProfile(user.ID)
	} else {
		m.Lock()
		m.Aarr[user.ID] = user
		m.Unlock()
	}
}

func (m *ProfilesInMemory) updateConfig(user Profile) {
	m.Lock()
	userToChange := m.Aarr[user.ID]
	userToChange.UserConfig = user.UserConfig
	m.Aarr[user.ID] = userToChange
	m.Unlock()
}

func (m *ProfilesInMemory) updatePasswd(user Profile) {
	m.Lock()
	userToChange := m.Aarr[user.ID]
	userToChange.Login = user.Login
	userToChange.Passwd = user.Passwd
	m.Aarr[user.ID] = userToChange
	m.Unlock()
}

func (m *ProfilesInMemory) delProfile(id int) {
	var keysForRemoval []string
	m.RLock()
	for k, cid := range m.Cjar {
		if cid == id {
			keysForRemoval = append(keysForRemoval, k)
		}
	}
	m.RUnlock()
	m.Lock()
	delete(m.Aarr, id)
	for _, k := range keysForRemoval {
		delete(m.Cjar, k)
	}
	m.Unlock()
}

func (m *ProfilesInMemory) clearAll() {
	m.Lock()
	for k := range m.Aarr {
		delete(m.Aarr, k)
	}
	for k := range m.Cjar {
		delete(m.Cjar, k)
	}
	m.Unlock()
}

func (m *ProfilesInMemory) constructUserList(db *sql.DB, DBType byte) error {
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
	UserList := []UserListElem{}
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
			UserList = append(UserList, UserListElem{ID, fullNameJob})
		}
	}
	m.Lock()
	m.UserList = UserList
	m.Unlock()
	return nil
}

func (m *ProfilesInMemory) returnUserList() []UserListElem {
	m.RLock()
	list := make([]UserListElem, len(m.UserList), cap(m.UserList))
	copy(list, m.UserList)
	m.RUnlock()
	return list
}

func (m *ProfilesInMemory) constructUnitList(db *sql.DB, DBType byte) error {
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
	UnitList := []UnitListElem{}
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
		UnitList = append(UnitList, UnitListElem{ID, unitNameComp})
	}
	m.Lock()
	m.UnitList = UnitList
	m.Unlock()
	return nil
}

func (m *ProfilesInMemory) returnUnitList() []UnitListElem {
	m.RLock()
	list := make([]UnitListElem, len(m.UnitList), cap(m.UnitList))
	copy(list, m.UnitList)
	m.RUnlock()
	return list
}

func (m *ProfilesInMemory) constructCorpList(db *sql.DB, DBType byte) error {
	rows, err := db.Query(`SELECT ID, ShortName FROM companies
ORDER BY ShortName ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ID int
	var ShortName sql.NullString
	CorpList := []CorpListElem{}
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
		CorpList = append(CorpList, CorpListElem{ID, companyName})
	}
	m.Lock()
	m.CorpList = CorpList
	m.Unlock()
	return nil
}

func (m *ProfilesInMemory) returnCorpList() []CorpListElem {
	m.RLock()
	list := make([]CorpListElem, len(m.CorpList), cap(m.CorpList))
	copy(list, m.CorpList)
	m.RUnlock()
	return list
}

func (m *ProfilesInMemory) checkLoggedin(sessionCookie *http.Cookie) (result bool, id int) {
	m.RLock()
	id, ok := m.Cjar[sessionCookie.Value]
	m.RUnlock()
	if ok {
		result = true
	} else {
		result = false
	}
	return result, id
}

func getLoginTemplate() *template.Template {
	logintmpl := `<!DOCTYPE html>
<head><meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="ie=edge">
<title>{{.AppTitle}}: {{.LoginPageTitle}}</title>
<link rel="shortcut icon" href="/assets/favicon.png">
<link rel="stylesheet" href="/assets/fonts.css">
<link rel="stylesheet" href="/assets/system-dark.css">
</head>
<body><div id="container"><div id="control">
<h1>{{.LoginPageTitle}}</h1><br>
<form id="loginForm" action="/login" method="POST">
<p>{{.LoninPrompt}}</p><br><p class="msgred">{{.WrongLoginMsg}}</p>
<div><label for="loginName" style="display: inline-block; width: 80px;">{{.LoninFieldLabel}}</label> <input id="loginName" type="text" size="15" class="field" name="loginName" placeholder="" value=""></div>
<div><label for="loginPasswd" style="display: inline-block; width: 80px;">{{.PasswordFieldLabel}}</label> <input id="loginPasswd" type="password" size="15" class="field" name="loginPasswd" placeholder="" value=""></div>
<br><input type="submit" class ="sbut" name="loginButton" value="{{.LoginButton}}"><br>
</form></div><div id="bottom">© 2022 <a href="https://github.com/alecxcode/edm" target="_blank">EDM Project</a></div></div></body></html>`
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
		log.Println(currentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (bs *BaseStruct) getLoginLang() LoginLang {
	return LoginLang{AppTitle: bs.lng.AppTitle, LoginPageTitle: bs.lng.LoginPageTitle, LoninPrompt: bs.lng.LoninPrompt,
		LoninFieldLabel: bs.lng.LoninFieldLabel, PasswordFieldLabel: bs.lng.PasswordFieldLabel,
		LoginButton: bs.lng.LoginButton, WrongLoginMsg: bs.lng.WrongLoginMsg}
}

func (bs *BaseStruct) authVerify(w http.ResponseWriter, r *http.Request) (res bool, id int) {
	thecookie, err := r.Cookie("sessionid")
	if err == http.ErrNoCookie {
		if !bs.team.verifyBruteForceCounter() {
			writeBruteForceStub(w)
			return false, 0
		}
		writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, false)
		return false, 0
	}
	allow, id := bs.team.checkLoggedin(thecookie)
	if !allow {
		if !bs.team.verifyBruteForceCounter() {
			writeBruteForceStub(w)
			return false, 0
		}
		writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, false)
		return false, 0
	}
	return true, id
}

func (bs *BaseStruct) loginHandler(w http.ResponseWriter, r *http.Request) {
	if !bs.team.verifyBruteForceCounter() {
		writeBruteForceStub(w)
		return
	}
	if r.Method == "POST" && r.FormValue("loginName") != "" {
		user := bs.team.getByLogin(r.FormValue("loginName"))
		var err error
		err = nil
		if user.Login == "" {
			user = Profile{Login: r.FormValue("loginName")}
			err = user.loadByIDorLogin(bs.db, bs.dbt, "Login")
			if err != nil {
				log.Println(currentFunction()+":", err)
			}
		}
		if err == nil && user.Login != "" && user.UserLock == 0 && r.FormValue("loginName") == user.Login &&
			comparePasswd(user.Passwd, r.FormValue("loginPasswd")) {
			cookval, _ := uuid.NewV4()
			var cookie = http.Cookie{
				Name:     "sessionid",
				Value:    cookval.String(),
				HttpOnly: true,
				Expires:  time.Now().AddDate(0, 0, 38000),
			}
			bs.team.set(cookie.Value, user)
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusFound)
			bs.team.resetBruteForceCounterImmediately()
		} else {
			bs.team.increaseBruteForceCounter(getIPAddr(r), r.FormValue("loginName"))
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
	bs.team.delCookie(thecookie.Value)
	thecookie.Expires = time.Now().AddDate(-1, -1, -1)
	http.SetCookie(w, thecookie)
	http.Redirect(w, r, "/", http.StatusFound)
	writeLoginPage(w, bs.getLoginLang(), bs.logintmpl, false)
	return
}

func (m *ProfilesInMemory) resetBruteForceCounterAfterMinutes(numberOfMinutes int) {
	time.Sleep(time.Duration(numberOfMinutes) * time.Minute)
	m.Lock()
	m.BruteForceCounter = 0
	m.Unlock()
}

func (m *ProfilesInMemory) resetBruteForceCounterImmediately() {
	m.Lock()
	m.BruteForceCounter = 0
	m.Unlock()
}

func (m *ProfilesInMemory) increaseBruteForceCounter(ipaddr string, login string) {
	m.Lock()
	m.BruteForceCounter++
	BruteForceCounter := m.BruteForceCounter
	m.Unlock()
	if BruteForceCounter >= ALLOWED_LOGIN_ATTEMPTS {
		log.Printf("System bruteforce shield activated, last attempt from IP addr: %s, login used: %s", ipaddr, login)
		go m.resetBruteForceCounterAfterMinutes(MINUTES_TO_WAIT_BRUTEFORCE_UNLOCK)
	}
}

func (m *ProfilesInMemory) verifyBruteForceCounter() (res bool) {
	m.RLock()
	BruteForceCounter := m.BruteForceCounter
	m.RUnlock()
	if BruteForceCounter >= ALLOWED_LOGIN_ATTEMPTS {
		res = false
	} else {
		res = true
	}
	return res
}

func writeBruteForceStub(w http.ResponseWriter) {
	fmt.Fprintf(w, `<!DOCTYPE html>
	<head><meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>EDM: System Bruteforce Attack Shield</title>
	<link rel="shortcut icon" href="/assets/favicon.png">
	<link rel="stylesheet" href="/assets/fonts.css">
	<link rel="stylesheet" href="/assets/system-pastel.css">
	</head>
	<body><div id="container"><div id="control"><h1>System Bruteforce Attack Shield</h1>
	<br><p class="msgredfx">System login function is temporarily locked due to bruteforce attack. Normally, this should not happen. Please, contact system admin.</p><br><br><br>
	</div></div></body></html>`)
}
