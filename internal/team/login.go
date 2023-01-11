package team

import (
	"database/sql"
	"edm/internal/core"
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"edm/pkg/passwd"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

// LoginBase is a struct which methods are http handlers
type LoginBase struct {
	lang     core.LoginLang
	db       *sql.DB
	dbType   byte // DB type as defined by constants in sqla package
	memorydb memdb.ObjectsInMemory
	template *template.Template
}

// NewLoginBase is a constructor
func NewLoginBase(loginLang core.LoginLang,
	db *sql.DB, dbType byte,
	memorydb memdb.ObjectsInMemory) LoginBase {
	return LoginBase{
		lang:     loginLang,
		db:       db,
		dbType:   dbType,
		memorydb: memorydb,
		template: getLoginTemplate(),
	}
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
<link rel="stylesheet" href="/static/system-rust.css">
</head>
<body><div id="container"><div id="control">
<h1>{{.LoginPageTitle}}</h1><br>
<form id="loginForm" action="/login" method="POST">
<p>{{.LoninPrompt}}</p><br><p class="msgred">{{.WrongLoginMsg}}</p>
<div><label for="loginName" style="display: inline-block; width: 80px;">{{.LoninFieldLabel}}</label> <input id="loginName" type="text" size="15" class="field" name="loginName" placeholder="" value=""></div>
<div><label for="loginPasswd" style="display: inline-block; width: 80px;">{{.PasswordFieldLabel}}</label> <input id="loginPasswd" type="password" size="15" class="field" name="loginPasswd" placeholder="" value=""></div>
<br><input type="submit" class="sbut" name="loginButton" value="{{.LoginButton}}"><br>
</form></div>
<div id="bottom">
<span>© 2023 <a href="https://edmproject.github.io" target="_blank">EDM Project</a></span>
<span>v` + core.AppVersion + `.</span>
<span><a href="/static/manual.html">Manual</a></span>
</div>
</div></body></html>`
	return template.Must(template.New("login").Parse(logintmpl))
}

func writeLoginPage(w http.ResponseWriter, LoginLang core.LoginLang, LoginTemplate *template.Template, wrongLogin bool) {
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

/*
func (ls *LoginStruct) AuthVerify(w http.ResponseWriter, r *http.Request) (res bool, id int) {
	thecookie, err := r.Cookie("sessionid")
	if err == http.ErrNoCookie {
		if !ls.team.VerifyBruteForceCounter() {
			writeBruteForceStub(w)
			return false, 0
		}
		writeLoginPage(w, ls.lang, ls.template, false)
		return false, 0
	}
	allow, id := ls.team.CheckSession(thecookie.Value)
	if !allow {
		if !ls.team.VerifyBruteForceCounter() {
			writeBruteForceStub(w)
			return false, 0
		}
		writeLoginPage(w, ls.lang, ls.template, false)
		return false, 0
	}
	return true, id
}
*/

// LoginHandler is http handler for login page
func (lb *LoginBase) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if !lb.memorydb.VerifyBruteForceCounter() {
		writeBruteForceStub(w)
		return
	}
	if r.Method == "POST" && r.FormValue("loginName") != "" {
		var err error
		err = nil
		user := Profile{Login: r.FormValue("loginName")}
		err = user.loadByIDorLogin(lb.db, lb.dbType, "Login")
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
			lb.memorydb.Set(cookie.Value, user)
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusFound)
			lb.memorydb.ResetBruteForceCounterImmediately()
		} else {
			lb.memorydb.IncreaseBruteForceCounter(accs.GetIPAddr(r), r.FormValue("loginName"))
			writeLoginPage(w, lb.lang, lb.template, true)
			return
		}
	} else if r.Method == "POST" && r.FormValue("loginName") == "" {
		writeLoginPage(w, lb.lang, lb.template, true)
		return
	} else {
		writeLoginPage(w, lb.lang, lb.template, false)
		return
	}
}

// LogoutHandler is http handler for logout
func (lb *LoginBase) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	thecookie, err := r.Cookie("sessionid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	lb.memorydb.DelCookie(thecookie.Value)
	thecookie.Expires = time.Now().AddDate(-1, -1, -1)
	http.SetCookie(w, thecookie)
	http.Redirect(w, r, "/", http.StatusFound)
	writeLoginPage(w, lb.lang, lb.template, false)
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
