package tasks

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/mail"
	"edm/pkg/memdb"
	"html/template"
	"regexp"
)

// TasksBase is a struct which methods are http handlers
type TasksBase struct {
	cfg struct {
		serverRoot    string
		systemURL     string
		removeAllowed bool
	}
	text tasksText
	i18n struct {
		langCode       string
		taskCaption    string
		commentCaption string
		taskStatuses   []string
		messages       core.AppMessages
	}
	validURLs    *regexp.Regexp
	emailCont    *regexp.Regexp
	mailchan     chan mail.EmailMessage
	db           *sql.DB
	dbType       byte // DB type as defined by constants in sqla package
	memorydb     memdb.ObjectsInMemory
	templates    *template.Template
	taskmailtmpl *template.Template
	commmailtmpl *template.Template
}

type tasksText struct {
	AppTitle       string
	TasksPageTitle string
	Task           string
	NewTask        string
	Comment        string
	NewComment     string
	TaskStatuses   []string
}

// NewTasksBase is a constructor
func NewTasksBase(serverRoot string, systemURL string, removeAllowed bool,
	text core.Text, i18n core.Si18n,
	ch chan mail.EmailMessage,
	db *sql.DB, dbType byte,
	memorydb memdb.ObjectsInMemory,
	tmpl *template.Template) TasksBase {
	return TasksBase{
		cfg: struct {
			serverRoot    string
			systemURL     string
			removeAllowed bool
		}{
			serverRoot:    serverRoot,
			systemURL:     systemURL,
			removeAllowed: removeAllowed,
		},
		text: tasksText{
			AppTitle:       text.AppTitle,
			TasksPageTitle: text.TasksPageTitle,
			Task:           text.Task,
			NewTask:        text.NewTask,
			Comment:        text.Comment,
			NewComment:     text.NewComment,
			TaskStatuses:   text.TaskStatuses,
		},
		i18n: struct {
			langCode       string
			taskCaption    string
			commentCaption string
			taskStatuses   []string
			messages       core.AppMessages
		}{
			langCode:       i18n.LangCode,
			taskCaption:    i18n.TaskCaption,
			commentCaption: i18n.CommentCaption,
			taskStatuses:   i18n.TaskStatuses,
			messages:       i18n.Messages,
		},
		validURLs:    regexp.MustCompile("^/tasks(/task/([0-9]+|new))?/?$"),
		emailCont:    regexp.MustCompile(`(?sU)<div class="somemargins content">(.+)</div>`),
		mailchan:     ch,
		db:           db,
		dbType:       dbType,
		memorydb:     memorydb,
		templates:    tmpl,
		taskmailtmpl: getTaskMailTemplate(),
		commmailtmpl: getCommMailTemplate(),
	}
}
