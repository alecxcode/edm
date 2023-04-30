package core

//go:generate python3 sql-gen.py

import (
	"database/sql"
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/alecxcode/sqla"
)

// GetSQLinitScript loads SQL file to create tables
func GetSQLinitScript(DBType byte, scriptsPath string) string {
	fname := "-create.sql"
	switch DBType {
	case sqla.SQLITE:
		fname = "sqlite" + fname
	case sqla.MSSQL:
		fname = "mssql" + fname
	case sqla.MYSQL:
		fname = "mysql" + fname
	case sqla.ORACLE:
		fname = "oracle" + fname
	case sqla.POSTGRESQL:
		fname = "postgresql" + fname
	}
	fname = filepath.Join(scriptsPath, fname)
	content, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Println("Error opening SQL init script:", err)
	}
	return string(content)
}

// PostgresqlMakeDatabase should create a Postgresql DB
func PostgresqlMakeDatabase(DSN string, DBName string, sqlStmt string) {
	db, err := sql.Open("pgx", DSN)
	if err != nil {
		log.Fatal(accs.CurrentFunction()+":", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatal(accs.CurrentFunction()+":", err)
	}
	checkStmt := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '?');"
	var res sql.NullBool
	row := db.QueryRow(checkStmt, DBName)
	err = row.Scan(&res)
	if err != nil {
		log.Printf("%q: %s\n", err, checkStmt)
		return
	}

	if res.Bool == false && res.Valid == true {
		sqlStmt := "CREATE DATABASE ?;"
		_, err = db.Exec(sqlStmt, DBName)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return
		}
		db, err = sql.Open("pgx", DSN)
		if err != nil {
			log.Fatal(accs.CurrentFunction()+":", err)
		}
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, "while creating tables")
			return
		}
	}
}

func CheckOldVersionDB(db *sql.DB, dbType byte, memorydb memdb.ObjectsInMemory) {

	var sqck = "SELECT UserConfig FROM profiles ORDER BY ID ASC LIMIT 1"
	var sqexec = "UPDATE profiles SET UserConfig = REPLACE(UserConfig, 'UseCalendarInConrols', 'UseCalendarInControls')"
	var SomeUserConfig sql.NullString
	var err error
	if dbType == sqla.ORACLE {
		sqck = "SELECT UserConfig FROM profiles ORDER BY ID ASC FETCH FIRST 1 ROWS ONLY"
	} else if dbType == sqla.MSSQL {
		sqck = "SELECT UserConfig FROM profiles ORDER BY ID ASC"
	}
	if DEBUG {
		log.Println(sqck)
	}

	err = db.QueryRow(sqck).Scan(&SomeUserConfig)
	if err != nil && err != sql.ErrNoRows {
		log.Println(accs.CurrentFunction()+":", err)
		return
	} else if err == sql.ErrNoRows {
		return
	}

	if strings.Contains(SomeUserConfig.String, "UseCalendarInConrols") {
		log.Println("Flawed old UserConfig. Replacing...")
		if DEBUG {
			log.Println(sqexec)
		}
		if _, err = db.Exec(sqexec); err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
		memorydb.ReplaceRawMany("Aarr", "UseCalendarInConrols", "UseCalendarInControls")
	}

}
