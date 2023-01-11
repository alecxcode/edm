package docs

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/mail"
	"edm/pkg/memdb"
	"html/template"
	"regexp"
)

// DocsBase is a struct which methods are http handlers
type DocsBase struct {
	cfg struct {
		serverRoot    string
		systemURL     string
		removeAllowed bool
	}
	text docsText
	i18n struct {
		langCode   string
		docCaption string
		categories []string
		docTypes   []string
		messages   core.AppMessages
	}
	validURLs *regexp.Regexp
	mailchan  chan mail.EmailMessage
	db        *sql.DB
	dbType    byte // DB type as defined by constants in sqla package
	memorydb  memdb.ObjectsInMemory
	templates *template.Template
	mailtmpl  *template.Template
}

type docsText struct {
	AppTitle      string
	DocsPageTitle string
	Document      string
	NewDocument   string
	Categories    []string
	DocTypes      []string
	ApprovalSign  []string
}

// NewDocsBase is a constructor
func NewDocsBase(serverRoot string, systemURL string, removeAllowed bool,
	text core.Text, i18n core.Si18n,
	ch chan mail.EmailMessage,
	db *sql.DB, dbType byte,
	memorydb memdb.ObjectsInMemory,
	tmpl *template.Template,
	mailtmpl *template.Template) DocsBase {
	return DocsBase{
		cfg: struct {
			serverRoot    string
			systemURL     string
			removeAllowed bool
		}{
			serverRoot:    serverRoot,
			systemURL:     systemURL,
			removeAllowed: removeAllowed,
		},
		text: docsText{
			AppTitle:      text.AppTitle,
			DocsPageTitle: text.DocsPageTitle,
			Document:      text.Document,
			NewDocument:   text.NewDocument,
			Categories:    text.Categories,
			DocTypes:      text.DocTypes,
			ApprovalSign:  text.ApprovalSign,
		},
		i18n: struct {
			langCode   string
			docCaption string
			categories []string
			docTypes   []string
			messages   core.AppMessages
		}{
			langCode:   i18n.LangCode,
			docCaption: i18n.DocCaption,
			categories: i18n.Categories,
			docTypes:   i18n.DocTypes,
			messages:   i18n.Messages,
		},
		validURLs: regexp.MustCompile("^/docs(/document/([0-9]+|[0-9]+/approval|new))?/?$"),
		mailchan:  ch,
		db:        db,
		dbType:    dbType,
		memorydb:  memorydb,
		templates: tmpl,
		mailtmpl:  mailtmpl,
	}
}
