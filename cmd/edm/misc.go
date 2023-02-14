package main

import (
	"edm/internal/config"
	"edm/pkg/accs"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

func defaultConfig(ServerSystem string, ServerRoot string) config.Config {
	return config.Config{
		ServerSystem:  ServerSystem,
		ServerRoot:    ServerRoot,
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
		REDISFlush:    "false",
		UseTLS:        "false",
	}
}

func getTemplates(templatesPath string) *template.Template {
	return template.Must(template.New("edm").
		Funcs(template.FuncMap{
			"returnFilterRender": accs.ReturnFilterRender,
			"returnHeadRender":   accs.ReturnHeadRender,
			"isThemeSystem":      accs.IsThemeSystem,
		}).ParseFiles(
		filepath.Join(templatesPath, "blocks.tmpl"),
		filepath.Join(templatesPath, "config.tmpl"),
		filepath.Join(templatesPath, "docs.tmpl"),
		filepath.Join(templatesPath, "document.tmpl"),
		filepath.Join(templatesPath, "approval.tmpl"),
		filepath.Join(templatesPath, "tasks.tmpl"),
		filepath.Join(templatesPath, "task.tmpl"),
		filepath.Join(templatesPath, "projs.tmpl"),
		filepath.Join(templatesPath, "project.tmpl"),
		filepath.Join(templatesPath, "team.tmpl"),
		filepath.Join(templatesPath, "profile.tmpl"),
		filepath.Join(templatesPath, "companies.tmpl"),
		filepath.Join(templatesPath, "company.tmpl"),
	))
}

func processCmdLineArgs(createdb string, filldb bool, runbrowser string, onlybrowser bool, consolelog bool) (string, bool, string, bool, bool, string, error) {
	validArgs := []string{"--createdb", "--filldb", "--nobrowser", "--onlybrowser", "--consolelog", "--config"}
	var err error = nil
	var preva string
	var config string
	var useconfig bool
	for i, a := range os.Args {
		if a == "--createdb" {
			createdb = "true"
		}
		if a == "--filldb" {
			filldb = true
		}
		if a == "--nobrowser" {
			runbrowser = "false"
		}
		if a == "--onlybrowser" {
			onlybrowser = true
		}
		if a == "--consolelog" {
			consolelog = true
		}
		if a == "--config" {
			useconfig = true
		}
		if preva == "--config" && !accs.SliceContainsStr(validArgs, a) && !strings.HasPrefix(a, "--") {
			config = a
		} else if i > 0 && !accs.SliceContainsStr(validArgs, a) {
			err = errors.New("wrong command line argument: the program finished with no action")
		}
		preva = a
	}
	if useconfig && (config == "" || config == " ") {
		err = errors.New("wrong command line argument: config path is not specified")
	}
	return createdb, filldb, runbrowser, onlybrowser, consolelog, config, err
}
