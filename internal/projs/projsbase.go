package projs

import (
	"database/sql"
	"edm/internal/core"
	"edm/pkg/memdb"
	"html/template"
	"regexp"

	"github.com/gorilla/websocket"
)

// ProjsBase is a struct which methods are http handlers
type ProjsBase struct {
	text          projsText
	removeAllowed bool
	validURLs     *regexp.Regexp
	db            *sql.DB
	dbType        byte // DB type as defined by constants in sqla package
	memorydb      memdb.ObjectsInMemory
	templates     *template.Template
	upgrader      websocket.Upgrader
}

type projsText struct {
	AppTitle       string
	ProjsPageTitle string
	Project        string
	NewProject     string
	ProjStatuses   []string
	TaskStatuses   []string
}

// NewProjsBase is a constructor
func NewProjsBase(removeAllowed bool, text core.Text, db *sql.DB, dbType byte,
	memorydb memdb.ObjectsInMemory, tmpl *template.Template, upgrader websocket.Upgrader) ProjsBase {
	return ProjsBase{
		removeAllowed: removeAllowed,
		text: projsText{
			AppTitle:       text.AppTitle,
			ProjsPageTitle: text.ProjsPageTitle,
			Project:        text.Project,
			NewProject:     text.NewProject,
			ProjStatuses:   text.ProjStatuses,
			TaskStatuses:   text.TaskStatuses,
		},
		validURLs: regexp.MustCompile("^/projs(/project/([0-9]+|new))?/?$"),
		db:        db,
		dbType:    dbType,
		memorydb:  memorydb,
		templates: tmpl,
		upgrader:  upgrader,
	}
}
