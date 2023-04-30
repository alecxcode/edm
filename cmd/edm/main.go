package main

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/docs"
	"edm/internal/filldata"
	"edm/internal/mail"
	"edm/internal/projs"
	"edm/internal/tasks"
	"edm/internal/team"
	"edm/pkg/accs"
	"edm/pkg/datetime"
	"edm/pkg/memdb"
	"edm/pkg/ramdb"
	"edm/pkg/redisdb"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alecxcode/sqla"
	"github.com/gorilla/websocket"
)

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

	var cfg = defaultConfig(SERVERSYSTEM, SERVERROOT)

	var (
		uploadPath     string
		logPath        string
		logFile        string
		staticPath     string // default images, fonts, css, js, unmodifiable files
		templatesPath  string
		i18nPath       string
		sqlscriptsPath string
		systemURL      string
		DSN            string // DB Path or URI
		db             *sql.DB
		dbType         byte
	)

	var err error

	// Reading command-line arguments
	filldb := false
	consolelog := false
	createdb := ""
	runbrowser := ""
	onlybrowser := false
	config := CONFIGPATH
	createdb, filldb, runbrowser, onlybrowser, consolelog, config, err = processCmdLineArgs(createdb, filldb, runbrowser, onlybrowser, consolelog)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = cfg.ReadConfig(config, cfg.ServerRoot)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
	}
	if createdb != "" {
		cfg.CreateDB = createdb
	}
	if runbrowser != "" {
		cfg.RunBrowser = runbrowser
	}

	// Server root path:
	if accs.FileExists(cfg.ServerRoot) != true {
		os.Mkdir(cfg.ServerRoot, 0700)
	}
	// Upload files path:
	uploadPath = filepath.Join(cfg.ServerRoot, "files")
	if accs.FileExists(uploadPath) != true {
		os.Mkdir(uploadPath, 0700)
	}
	// Logging to a file:
	logPath = filepath.Join(cfg.ServerRoot, "logs")
	if accs.FileExists(logPath) != true {
		os.Mkdir(logPath, 0700)
	}
	logFile = filepath.Join(logPath, "edm-"+datetime.GetCurrentYearMStr()+".log")
	applog, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
	}
	defer applog.Close()
	defer applog.Sync()
	if !consolelog {
		log.SetOutput(applog) // Setting logging to a file
	}

	// Creating DB connection string (for sqlite it is a path):
	DSN = sqla.BuildDSN(cfg.DBType, cfg.DBName, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword)
	if cfg.DBType == "sqlite" {
		DSN = filepath.Join(cfg.ServerRoot, cfg.DBName)
	}
	dbType = sqla.ReturnDBType(cfg.DBType)
	sqlscriptsPath = filepath.Join(cfg.ServerSystem, "sqlscripts")

	if cfg.CreateDB == "fulldatabase" && dbType == sqla.POSTGRESQL {
		log.Println("Creating Database...")
		core.PostgresqlMakeDatabase(DSN, cfg.DBName, core.GetSQLinitScript(dbType, sqlscriptsPath))
		return
	}

	if accs.StrToBool(cfg.CreateDB) {
		log.Println("Creating Schema...")
		sqla.CreateDB(dbType, DSN, core.GetSQLinitScript(dbType, sqlscriptsPath))
		db = sqla.OpenSQLConnection(dbType, DSN)
		defer db.Close()
		//core.CheckOldVersionTables(db, dbType)
		team.CreateFirstAdmin(db, dbType, cfg.DefaultLang)
		return
	}

	// Creating DB if nonexistent:
	sqliteFirstRun := false
	if dbType == sqla.SQLITE {
		if accs.FileExists(DSN) != true {
			log.Println("Creating Database...")
			sqliteFirstRun = true
			sqla.CreateDB(dbType, DSN, core.GetSQLinitScript(dbType, sqlscriptsPath))
		}
	}

	if dbType == sqla.SQLITE {
		DSN += "?_foreign_keys=on&_cache_size=-100000"
	}

	db = sqla.OpenSQLConnection(dbType, DSN)
	defer db.Close()
	if sqliteFirstRun {
		team.CreateFirstAdmin(db, dbType, cfg.DefaultLang)
	}
	cfg.DBUser = "no access"     // clear DB login in memory
	cfg.DBPassword = "no access" // clear DB passwd in memory

	themes := core.GetThemeList(cfg.ServerSystem)
	dateFormats := []string{"yyyy-mm-dd", "yyyy.mm.dd", "dd.mm.yyyy", "dd/mm/yyyy", "Mon dd, yyyy", "mm/dd/yyyy"}
	timeFormats := []string{"12h am/pm", "24h"}
	langCodes := core.GetLangList(cfg.ServerSystem)

	var memorydb memdb.ObjectsInMemory
	if cfg.REDISConnect != "" {
		memorydb = redisdb.NewReidsConnection([]string{"UserList", "UnitList", "CorpList"}, cfg.REDISConnect, cfg.REDISPassword, accs.StrToBool(cfg.REDISFlush))
	} else {
		memorydb = ramdb.NewObjectsInMemory([]string{"UserList", "UnitList", "CorpList"})
	}
	defer memorydb.Close()

	core.CheckOldVersionDB(db, dbType, memorydb)

	protocol := "http"
	if accs.StrToBool(cfg.UseTLS) {
		protocol = "https"
	}

	if cfg.DomainName == "" {
		cfg.DomainName = cfg.ServerHost
	}
	if (protocol == "http" && cfg.ServerPort == "80") || (protocol == "https" && cfg.ServerPort == "443") {
		systemURL = protocol + "://" + cfg.DomainName
	} else {
		systemURL = protocol + "://" + cfg.DomainName + ":" + cfg.ServerPort
	}

	i18nPath = filepath.Join(cfg.ServerSystem, "i18nserver")
	text := core.NewTextStruct()
	i18n := core.Newi18nStruct(filepath.Join(i18nPath, cfg.DefaultLang+".json"))

	templatesPath = filepath.Join(cfg.ServerSystem, "templates")
	templates := getTemplates(templatesPath)
	anymailtmpl := core.GetAnyMailTemplate()

	err = core.ConstructUserList(db, dbType, memorydb)
	if err != nil {
		log.Println(accs.CurrentFunction()+": constructUserList:", err)
	}
	err = core.ConstructUnitList(db, dbType, memorydb)
	if err != nil {
		log.Println(accs.CurrentFunction()+": constructUnitList:", err)
	}
	err = core.ConstructCorpList(db, dbType, memorydb)
	if err != nil {
		log.Println(accs.CurrentFunction()+": constructCorpList", err)
	}

	// Filling database with test data
	if filldb {
		filldata.FillDBwithTestData(db, dbType)
		return
	}

	// Launching browser:
	if onlybrowser {
		accs.RunClient(cfg.UseTLS, "127.0.0.1:"+cfg.ServerPort, 1, 2)
		return
	}
	if accs.IsPrevInstanceRunning(cfg.ServerHost + ":" + cfg.ServerPort) {
		log.Println("server is already running, quiting")
		if accs.StrToBool(cfg.RunBrowser) {
			accs.RunClient(cfg.UseTLS, "127.0.0.1:"+cfg.ServerPort, 1, 2)
		}
		return
	}
	if accs.StrToBool(cfg.RunBrowser) && !onlybrowser {
		go accs.RunClient(cfg.UseTLS, "127.0.0.1:"+cfg.ServerPort, 100, 2)
	}

	// Server code:
	log.Println("running server")

	// Launching mailer monitor:
	mailchan := make(chan mail.EmailMessage, 1024)
	defer close(mailchan)
	go mail.MailerMonitor(mailchan, cfg.SMTPHost, accs.StrToInt(cfg.SMTPPort), cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPEmail, i18n.Messages.MailerName, db, dbType, core.DEBUG)
	if cfg.SMTPHost != "" {
		go mail.ReadMailFromDB(mailchan, 30, db, dbType, core.DEBUG)
	}

	// Static files handler:
	staticPath = filepath.Join(cfg.ServerSystem, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	removeAllowed := accs.StrToBool(cfg.RemoveAllowed)

	// Routing handlers:
	http.HandleFunc("/appversion", core.GetAppVersion)
	http.HandleFunc("/portal/", core.PortalHandler)

	cr := core.NewCoreBase(cfg.StartPage, cfg.ServerSystem, uploadPath, memorydb)
	http.HandleFunc("/", cr.IndexHandler)
	http.HandleFunc("/favicon.ico", cr.ServeFavIcon)
	http.HandleFunc("/files/", cr.ServeUploads)

	lg := team.NewLoginBase(i18n.LoginLang, db, dbType, memorydb)
	http.HandleFunc("/login", lg.LoginHandler)
	http.HandleFunc("/logout", lg.LogoutHandler)

	org := team.NewTeamBase(cfg.DefaultLang, systemURL, text, i18n,
		mailchan, db, dbType, memorydb, templates, anymailtmpl,
		themes, dateFormats, timeFormats, langCodes)
	http.HandleFunc("/team/", org.TeamHandler)
	http.HandleFunc("/team/profile/", org.ProfileHandler)
	http.HandleFunc("/config", org.UserConfigHandler)
	http.HandleFunc("/companies/", org.CompaniesHandler)
	http.HandleFunc("/companies/company/", org.CompanyHandler)

	doc := docs.NewDocsBase(cfg.ServerRoot, systemURL, removeAllowed, text, i18n,
		mailchan, db, dbType, memorydb, templates, anymailtmpl)
	http.HandleFunc("/docs/", doc.DocsHandler)
	http.HandleFunc("/docs/document/", doc.DocumentHandler)

	tsk := tasks.NewTasksBase(cfg.ServerRoot, systemURL, removeAllowed, text, i18n,
		mailchan, db, dbType, memorydb, templates)
	http.HandleFunc("/tasks/", tsk.TasksHandler)
	http.HandleFunc("/tasks/task/", tsk.TaskHandler)
	http.HandleFunc("/tasks/loadtasks", tsk.LoadTasksAPI)
	http.HandleFunc("/tasks/assigntask", tsk.AssignTaskAPI)
	http.HandleFunc("/tasks/updateproj", tsk.UpdateTaskProjectAPI)

	pro := projs.NewProjsBase(removeAllowed, text, db, dbType, memorydb, templates, websocket.Upgrader{})
	http.HandleFunc("/projs/", pro.ProjsHandler)
	http.HandleFunc("/projs/project/", pro.ProjectHandler)
	http.HandleFunc("/projs/getproj", pro.GetProjAPI)
	http.HandleFunc("/projs/setstatus", pro.SetProjStatusAPI)
	http.HandleFunc("/projs/ws", pro.WsHandler)

	if accs.StrToBool(cfg.UseTLS) {
		log.Fatal(http.ListenAndServeTLS(cfg.ServerHost+":"+cfg.ServerPort,
			accs.GetAbsoluteOrRelativePath(cfg.ServerRoot, cfg.SSLCertFile),
			accs.GetAbsoluteOrRelativePath(cfg.ServerRoot, cfg.SSLKeyFile),
			nil))
	} else {
		log.Fatal(http.ListenAndServe(cfg.ServerHost+":"+cfg.ServerPort, nil))
	}

}
