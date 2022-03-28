package main

//go:generate python3 sql-gen.py

import (
	"database/sql"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/alecxcode/sqla"
)

func getSQLinitScript(DBType byte, scriptsPath string) string {
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

func postgresqlMakeDatabase(DSN string, DBName string, sqlStmt string) {
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		log.Fatal(currentFunction()+":", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatal(currentFunction()+":", err)
	}
	checkStmt := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '" + DBName + "');"
	var res sql.NullBool
	row := db.QueryRow(checkStmt)
	err = row.Scan(&res)
	if err != nil {
		log.Printf("%q: %s\n", err, checkStmt)
		return
	}

	if res.Bool == false && res.Valid == true {
		sqlStmt := "CREATE DATABASE edm;"
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return
		}
		db, err = sql.Open("postgres", DSN)
		if err != nil {
			log.Fatal(currentFunction()+":", err)
		}
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, "while creating tables")
			return
		}

	}
}
