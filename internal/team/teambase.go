package team

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/mail"
	"edm/pkg/memdb"
	"html/template"
	"regexp"
)

// TeamBase is a struct which methods are http handlers
type TeamBase struct {
	cfg struct {
		defaultLang string
		systemURL   string
	}
	text teamText
	i18n struct {
		langCode string
		messages core.AppMessages
	}
	validURLs struct {
		team *regexp.Regexp
		comp *regexp.Regexp
		conf *regexp.Regexp
	}
	mailchan  chan mail.EmailMessage
	db        *sql.DB
	dbType    byte // DB type as defined by constants in sqla package
	memorydb  memdb.ObjectsInMemory
	templates *template.Template
	mailtmpl  *template.Template
	options   struct {
		Themes      []string
		DateFormats []string
		TimeFormats []string
		LangCodes   []string
	}
}

type teamText struct {
	AppTitle           string
	ConfigPageTitle    string
	AppVersion         string
	TeamPageTitle      string
	Profile            string
	NewProfile         string
	CompaniesPageTitle string
	Company            string
	NewCompany         string
	Unit               string
}

// NewTeamBase is a constructor
func NewTeamBase(defaultLang string, systemURL string,
	text core.Text, i18n core.Si18n,
	ch chan mail.EmailMessage,
	db *sql.DB, dbType byte,
	memorydb memdb.ObjectsInMemory,
	tmpl *template.Template,
	mailtmpl *template.Template,
	themes []string,
	dateFormats []string,
	timeFormats []string,
	langCodes []string) TeamBase {
	return TeamBase{
		cfg: struct {
			defaultLang string
			systemURL   string
		}{
			defaultLang: defaultLang,
			systemURL:   systemURL,
		},
		text: teamText{
			AppTitle:           text.AppTitle,
			ConfigPageTitle:    text.ConfigPageTitle,
			AppVersion:         core.AppVersion,
			TeamPageTitle:      text.TeamPageTitle,
			Profile:            text.Profile,
			NewProfile:         text.NewProfile,
			CompaniesPageTitle: text.CompaniesPageTitle,
			Company:            text.Company,
			NewCompany:         text.NewCompany,
			Unit:               text.Unit,
		},
		i18n: struct {
			langCode string
			messages core.AppMessages
		}{
			langCode: i18n.LangCode,
			messages: i18n.Messages,
		},
		validURLs: struct {
			team *regexp.Regexp
			comp *regexp.Regexp
			conf *regexp.Regexp
		}{
			team: regexp.MustCompile("^/team(/profile/([0-9]+|new))?/?$"),
			comp: regexp.MustCompile("^/companies(/company/([0-9]+|new))?/?$"),
			conf: regexp.MustCompile("^/config$"),
		},
		mailchan:  ch,
		db:        db,
		dbType:    dbType,
		memorydb:  memorydb,
		templates: tmpl,
		mailtmpl:  mailtmpl,
		options: struct {
			Themes      []string
			DateFormats []string
			TimeFormats []string
			LangCodes   []string
		}{
			Themes:      themes,
			DateFormats: dateFormats,
			TimeFormats: timeFormats,
			LangCodes:   langCodes,
		},
	}
}
