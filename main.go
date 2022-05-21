package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/alecxcode/sqla"
)

// // AppTitle is this application name (not used for now)
// const AppTitle = "EDM"

// Undefined is for any unknown category, type, etc.
const Undefined = 0

// DEBUG defines if verbose logging should be enabled
const DEBUG = false

// BaseStruct for handlers
type BaseStruct struct {
	cfg        *Config
	lng        Lang
	currencies map[int]string
	utcdiff    int64
	validURLs  struct {
		Indx   *regexp.Regexp
		Config *regexp.Regexp
		Docs   *regexp.Regexp
		Task   *regexp.Regexp
		Team   *regexp.Regexp
		Comp   *regexp.Regexp
		Port   *regexp.Regexp
	}
	regexes struct {
		emailCont *regexp.Regexp
	}
	systemURL    string
	mailchan     chan EmailMessage
	uploads      http.Handler
	db           *sql.DB
	dbt          byte // DB type as defined by constants in sqla package
	team         ProfilesInMemory
	templates    *template.Template
	logintmpl    *template.Template
	anymailtmpl  *template.Template
	taskmailtmpl *template.Template
	commmailtmpl *template.Template
	options      struct {
		Themes      []string
		DateFormats []string
		TimeFormats []string
	}
}

func main() {

	// CONFIGPATH is default path where config file is located.
	// If not specified, then the directory of user home + ".edm" (e.g. "~/.edm") will be used.
	// The name of the file which app will look for is: edm-system.cfg.
	// If path and (or) file do not exist the app will try to create them.
	const CONFIGPATH = ""
	// SERVERSYSTEM stores static files, templates, themes, etc. Not writable.
	const SERVERSYSTEM = "."
	// SERVERROOT is default path where server modifiable files (uploads, logs, etc.) are stored.
	// This directory should be writable for the app.
	// If not specified, then the directory of user home + ".edm" (e.g. "~/.edm") will be used.
	const SERVERROOT = ""

	var cfg = Config{
		ServerSystem:  SERVERSYSTEM,
		ServerRoot:    SERVERROOT,
		ServerHost:    "127.0.0.1",
		ServerPort:    "8090",
		DomainName:    "127.0.0.1",
		DefaultLang:   "en",
		StartPage:     "docs",
		RemoveAllowed: "true",
		RunBrowser:    "true",
		CreateDB:      "false",
		DBType:        "sqlite",
		DBName:        "edm.db",
		DBHost:        "127.0.0.1",
		UseTLS:        "false",
	}

	var (
		uploadPath     string
		logPath        string
		logFile        string
		assetsPath     string // default images, fonts, css, js, unmodifiable files
		templatesPath  string
		sqlscriptsPath string
		db             *sql.DB
		DBT            byte
		DSN            string // DB Path
	)

	err := cfg.readConfig(CONFIGPATH, cfg.ServerRoot)
	if err != nil {
		log.Println(currentFunction()+":", err)
	}

	// Reading command-line arguments
	consolelog := false
	filldb := false
	for _, a := range os.Args {
		if a == "--createdb" {
			cfg.CreateDB = "true"
		}
		if a == "--filldb" {
			filldb = true
		}
		if a == "--nobrowser" {
			cfg.RunBrowser = "false"
		}
		if a == "--consolelog" {
			consolelog = true
		}
	}

	// Server root path:
	if fileExists(cfg.ServerRoot) != true {
		os.Mkdir(cfg.ServerRoot, 0700)
	}
	// Upload files path:
	uploadPath = filepath.Join(cfg.ServerRoot, "files")
	if fileExists(uploadPath) != true {
		os.Mkdir(uploadPath, 0700)
	}
	// Logging to a file:
	logPath = filepath.Join(cfg.ServerRoot, "logs")
	if fileExists(logPath) != true {
		os.Mkdir(logPath, 0700)
	}
	logFile = filepath.Join(logPath, "edm-"+getCurrentYearMStr()+".log")
	applog, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(currentFunction()+":", err)
	}
	defer applog.Close()
	// Uncomment in Release; Comment in Development:
	if !consolelog {
		log.SetOutput(applog) // Setting logging to a file
	}

	// Creating DB connection string (for sqlite it is a path):
	DSN = sqla.BuildDSN(cfg.DBType, cfg.DBName, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword)
	if cfg.DBType == "sqlite" {
		DSN = filepath.Join(cfg.ServerRoot, cfg.DBName)
	}
	DBT = sqla.ReturnDBType(cfg.DBType)
	sqlscriptsPath = filepath.Join(cfg.ServerSystem, "sqlscripts")

	if cfg.CreateDB == "fulldatabase" && DBT == sqla.POSTGRESQL {
		log.Println("Creating Database...")
		postgresqlMakeDatabase(DSN, cfg.DBName, getSQLinitScript(DBT, sqlscriptsPath))
		return
	}

	if cfg.CreateDB == "true" {
		log.Println("Creating Schema...")
		sqla.CreateDB(DBT, DSN, getSQLinitScript(DBT, sqlscriptsPath))
		db = sqla.OpenSQLConnection(DBT, DSN)
		defer db.Close()
		createFirstAdmin(db, DBT)
		return
	}

	// Creating DB if nonexistent:
	sqliteFirstRun := false
	if DBT == sqla.SQLITE {
		if fileExists(DSN) != true {
			log.Println("Creating Database...")
			sqliteFirstRun = true
			sqla.CreateDB(DBT, DSN, getSQLinitScript(DBT, sqlscriptsPath))
		}
	}

	if DBT == sqla.SQLITE {
		DSN += "?_foreign_keys=on&_cache_size=-100000"
	}

	db = sqla.OpenSQLConnection(DBT, DSN)
	defer db.Close()
	if sqliteFirstRun {
		createFirstAdmin(db, DBT)
	}
	cfg.DBUser = "no access"     // clear DB login in memory
	cfg.DBPassword = "no access" // clear DB passwd in memory

	// Creating BaseStruct - requited for handlers
	bs := BaseStruct{
		cfg:        &cfg,
		db:         db,
		dbt:        DBT,
		currencies: getCurrencies(),
		//utcdiff:    getUTCDiff(), // difference between local time and UTC in minutes, reserved for future use
		team: ProfilesInMemory{
			Aarr: make(map[int]Profile),
			Cjar: make(map[string]int),
		},
		options: struct {
			Themes      []string
			DateFormats []string
			TimeFormats []string
		}{
			Themes:      []string{"dark", "light", "monochrome-dark", "monochrome-light"},
			DateFormats: []string{"yyyy-mm-dd", "yyyy.mm.dd", "dd.mm.yyyy", "dd/mm/yyyy", "Mon dd, yyyy", "mm/dd/yyyy"},
			TimeFormats: []string{"12h am/pm", "24h"},
		},
	}

	protocol := "http"
	if cfg.UseTLS == "true" {
		protocol = "https"
	}
	if (protocol == "http" && cfg.ServerPort == "80") || (protocol == "https" && cfg.ServerPort == "443") {
		bs.systemURL = protocol + "://" + cfg.DomainName
	} else {
		bs.systemURL = protocol + "://" + cfg.DomainName + ":" + cfg.ServerPort
	}

	// valid paths for handlers
	bs.validURLs.Indx = regexp.MustCompile("^/?$")
	bs.validURLs.Config = regexp.MustCompile("^/config$")
	bs.validURLs.Docs = regexp.MustCompile("^/docs(/document/([0-9]+|new))?/?$")
	bs.validURLs.Task = regexp.MustCompile("^/tasks(/task/([0-9]+|new))?/?$")
	bs.validURLs.Team = regexp.MustCompile("^/team(/profile/([0-9]+|new))?/?$")
	bs.validURLs.Comp = regexp.MustCompile("^/companies(/company/([0-9]+|new))?/?$")
	bs.validURLs.Port = regexp.MustCompile("^/portal/?$")

	bs.regexes.emailCont = regexp.MustCompile(`(?sU)<div class="somemargins content">(.+)</div>`)

	templatesPath = filepath.Join(cfg.ServerSystem, "templates", bs.cfg.DefaultLang)
	bs.templates = template.Must(template.ParseFiles(
		filepath.Join(templatesPath, "config.tmpl"),
		filepath.Join(templatesPath, "docs.tmpl"),
		filepath.Join(templatesPath, "document.tmpl"),
		filepath.Join(templatesPath, "tasks.tmpl"),
		filepath.Join(templatesPath, "task.tmpl"),
		filepath.Join(templatesPath, "team.tmpl"),
		filepath.Join(templatesPath, "profile.tmpl"),
		filepath.Join(templatesPath, "companies.tmpl"),
		filepath.Join(templatesPath, "company.tmpl"),
	))
	bs.logintmpl = getLoginTemplate()
	bs.anymailtmpl = getAnyMailTemplate()
	bs.taskmailtmpl = getTaskMailTemplate()
	bs.commmailtmpl = getCommMailTemplate()
	bs.lng = newLangStruct(filepath.Join(templatesPath, "lng.json"))

	err = bs.team.constructUserList(db, DBT)
	if err != nil {
		log.Println(currentFunction()+": constructUserList:", err)
	}
	err = bs.team.constructUnitList(db, DBT)
	if err != nil {
		log.Println(currentFunction()+": constructUnitList:", err)
	}
	err = bs.team.constructCorpList(db, DBT)
	if err != nil {
		log.Println(currentFunction()+": constructCorpList", err)
	}

	// Filling database with test data
	if filldb {
		log.Println("Populating DB with test data...")
		fillDBwithTestData(bs.db, bs.dbt)
		return
	}

	// Server code:
	if isPrevInstanceRunning(cfg.ServerHost + ":" + cfg.ServerPort) {
		log.Println("server is already running, quiting")
		if cfg.RunBrowser == "true" {
			runClient(cfg.UseTLS, "127.0.0.1:"+cfg.ServerPort, 1, 2)
		}
		return
	} else {
		log.Println("running server")
	}

	// Launching mailer monitor:
	bs.mailchan = make(chan EmailMessage, 1024)
	go mailerMonitor(bs.mailchan, bs.cfg.SMTPHost, strToInt(bs.cfg.SMTPPort), bs.cfg.SMTPUser, bs.cfg.SMTPPassword, bs.cfg.SMTPEmail, bs.db, bs.dbt)
	go readMailFromDB(bs.mailchan, 30, bs.db, bs.dbt)

	// Launching browser:
	if cfg.RunBrowser == "true" {
		go runClient(cfg.UseTLS, "127.0.0.1:"+cfg.ServerPort, 100, 2)
	}

	// Static files handler:
	assetsPath = filepath.Join(cfg.ServerSystem, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsPath))))

	// Uploads dir handler (func below: serveUploads):
	bs.uploads = http.StripPrefix("/files/", http.FileServer(http.Dir(uploadPath)))

	// Routing handlers:
	http.HandleFunc("/", bs.indexHandler)
	http.HandleFunc("/favicon.ico", bs.serveFavIcon)
	http.HandleFunc("/docs/", bs.docsHandler)
	http.HandleFunc("/docs/document/", bs.documentHandler)
	http.HandleFunc("/tasks/", bs.tasksHandler)
	http.HandleFunc("/tasks/task/", bs.taskHandler)
	http.HandleFunc("/team/", bs.teamHandler)
	http.HandleFunc("/team/profile/", bs.profileHandler)
	http.HandleFunc("/companies/", bs.companiesHandler)
	http.HandleFunc("/companies/company/", bs.companyHandler)
	http.HandleFunc("/portal/", bs.portalHandler)
	http.HandleFunc("/config", bs.userConfigHandler)
	http.HandleFunc("/login", bs.loginHandler)
	http.HandleFunc("/logout", bs.logoutHandler)
	http.HandleFunc("/files/", bs.serveUploads)
	if cfg.UseTLS == "true" {
		log.Fatal(http.ListenAndServeTLS(cfg.ServerHost+":"+cfg.ServerPort,
			filepath.Join(cfg.ServerRoot, cfg.SSLCertFile),
			filepath.Join(cfg.ServerRoot, cfg.SSLKeyFile),
			nil))
	} else {
		log.Fatal(http.ListenAndServe(cfg.ServerHost+":"+cfg.ServerPort, nil))
	}

	close(bs.mailchan)
	applog.Sync()
}

func (bs *BaseStruct) indexHandler(w http.ResponseWriter, r *http.Request) {
	allow, _ := bs.authVerify(w, r)
	if !allow {
		return
	}
	if bs.validURLs.Indx.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}
	switch bs.cfg.StartPage {
	case "docs":
		http.Redirect(w, r, "/docs/", http.StatusSeeOther)
	case "tasks":
		http.Redirect(w, r, "/tasks/", http.StatusSeeOther)
	case "team":
		http.Redirect(w, r, "/team/", http.StatusSeeOther)
	case "portal":
		http.Redirect(w, r, "/portal/", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/docs/", http.StatusSeeOther)
	}
}

func (bs *BaseStruct) serveFavIcon(w http.ResponseWriter, r *http.Request) {
	favIconPath := filepath.Join(bs.cfg.ServerSystem, "assets", "favicon.png")
	http.ServeFile(w, r, favIconPath)
}

func (bs *BaseStruct) serveUploads(w http.ResponseWriter, r *http.Request) {
	allow, _ := bs.authVerify(w, r)
	if !allow {
		return
	}
	bs.uploads.ServeHTTP(w, r)
}

func (bs *BaseStruct) portalHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This function is under development. Current URL: %s", r.URL.Path)
}
