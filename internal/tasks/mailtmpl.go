package tasks

import (
	"bytes"
	"database/sql"
	"edm/internal/core"
	"edm/internal/mail"
	"edm/pkg/accs"
	"html/template"
	"log"
	"regexp"
	"strings"
)

type bbcodedMail interface {
	getSubj() string
	constructToChannel(db *sql.DB, DBType byte, mailtmlp *template.Template, mailchan chan mail.EmailMessage, email mail.EmailMessage, emailCont *regexp.Regexp)
}

// TaskMail is a template for a task-related message
type TaskMail struct {
	Subj         string
	Task         Task
	AppMessages  core.AppMessages
	TaskCaption  string
	TaskStatuses []string
	SystemURL    string
}

func getTaskMailTemplate() *template.Template {
	tmpl := core.GetTemplateHeader() + `
<body><div id="container"><div id="main">
<h1>{{.Subj}}</h1>{{$ci := .Task.GiveCreatorID}}{{$ai := .Task.GiveAssigneeID}}{{$statusindicator := "txtblue"}}
{{if eq .Task.TaskStatus 3}}{{$statusindicator = "txtred"}}
{{else if eq .Task.TaskStatus 4}}{{$statusindicator = "txtgreen"}}
{{else if eq .Task.TaskStatus 5}}{{$statusindicator = "txtbw"}}
{{else if eq .Task.TaskStatus 6}}{{$statusindicator = "txtora"}}{{end}}
<h2>{{.TaskCaption}} #{{.Task.ID}}: {{.Task.Topic}}</h2>
{{if $ci}}<div>{{.AppMessages.Captions.Creator}}: {{.Task.Creator.GiveSelfNameJob}}</div>{{end}}
<div>{{.AppMessages.Captions.CreatedTime}}: {{.Task.GiveDateTime "Created" "" "" " "}}</div>
<div>{{.AppMessages.Captions.TaskStatus}}: <span class="{{$statusindicator}}">{{.Task.GiveStatus .TaskStatuses "Unknown"}}</span></p>
<div>{{.AppMessages.Captions.TaskStartDueTime}}:
  {{if .Task.PlanStart.Day}}{{.Task.GiveDateTime "PlanStart" "" "" " "}}{{else}}...{{end}} - 
  {{if .Task.PlanDue.Day}}{{.Task.GiveDateTime "PlanDue" "" "" " "}}{{else}}...{{end}}</div>
{{if $ai}}<div>{{.AppMessages.Captions.Assignee}}: {{.Task.Assignee.GiveSelfNameJob}}</div>{{end}}
<div class="somemargins content">{{.Task.Content}}</div>
{{if .Task.FileList}} <div>{{.AppMessages.Captions.FileList}}:<br>
{{range .Task.FileList}}
  <a class="msglnk" href="{{$.SystemURL}}/files/tasks/{{$.Task.ID}}/{{.}}" target="_blank">{{.}}</a>
<br>{{end}}</div>
{{end}}
<div class="margintop"><a class="sbut" href="{{$.SystemURL}}/tasks/task/{{.Task.ID}}" target="_blank">{{.AppMessages.Captions.Open}}</a></div>
</div><div class="bottom" id="bottom">{{.AppMessages.DoNotReply}}<br>
<a href="{{.SystemURL}}" target="_blank">{{.AppMessages.MailerName}}</a></div></div></body></html>`
	return template.Must(template.New("taskmail").Parse(tmpl))
}

func (m TaskMail) getSubj() string {
	return m.Subj
}
func (m TaskMail) constructToChannel(db *sql.DB, DBType byte, mailtmpl *template.Template, mailchan chan mail.EmailMessage, email mail.EmailMessage, emailCont *regexp.Regexp) {
	constructToChannel(db, DBType, mailtmpl, mailchan, email, emailCont, m)
}

// CommMail is a template for a comment-related message
type CommMail struct {
	Subj           string
	TaskID         int
	TaskTopic      string
	Comment        Comment
	CommentIndex   int
	AppMessages    core.AppMessages
	TaskCaption    string
	CommentCaption string
	SystemURL      string
}

func getCommMailTemplate() *template.Template {
	tmpl := core.GetTemplateHeader() + `
<body><div id="container"><div id="main">
<h1>{{.Subj}}</h1>{{$ci := .Comment.GiveCreatorID}}
<div>{{.TaskCaption}} #{{.TaskID}}: {{.TaskTopic}}</div>
<h2>{{.CommentCaption}} #{{.CommentIndex}}: {{.Comment.GiveDateTime "" "" " "}}</h2>
{{if $ci}}<div>{{.AppMessages.Captions.Creator}}: {{.Comment.Creator.GiveSelfNameJob}}</div>{{end}}
<div class="somemargins content">{{.Comment.Content}}</div>
{{if .Comment.FileList}} <div>{{.AppMessages.Captions.FileList}}:<br>
{{range .Comment.FileList}}
  <a class="msglnk" href="{{$.SystemURL}}/files/tasks/{{$.TaskID}}/{{$.Comment.ID}}/{{.}}" target="_blank">{{.}}</a>
<br>{{end}}</div>
{{end}}
<div class="margintop"><a class="sbut" href="{{$.SystemURL}}/tasks/task/{{.TaskID}}#comment{{.Comment.ID}}" target="_blank">{{.AppMessages.Captions.Open}}</a></div>
</div><div class="bottom" id="bottom">{{.AppMessages.DoNotReply}}<br>
<a href="{{.SystemURL}}" target="_blank">{{.AppMessages.MailerName}}</a></div></div></body></html>`
	return template.Must(template.New("commmail").Parse(tmpl))
}

func (m CommMail) getSubj() string {
	return m.Subj
}
func (m CommMail) constructToChannel(db *sql.DB, DBType byte, mailtmpl *template.Template, mailchan chan mail.EmailMessage, email mail.EmailMessage, emailCont *regexp.Regexp) {
	constructToChannel(db, DBType, mailtmpl, mailchan, email, emailCont, m)
}

func constructToChannel(db *sql.DB, DBType byte, mailtmpl *template.Template, mailchan chan mail.EmailMessage, email mail.EmailMessage, emailCont *regexp.Regexp, m bbcodedMail) {
	if len(email.SendTo) > 0 || len(email.SendCc) > 0 {
		var tmpl bytes.Buffer
		if err := mailtmpl.Execute(&tmpl, m); err != nil {
			log.Println("executing any mail template: "+m.getSubj()+":", err)
		}
		email.Cont = tmpl.String()
		cont := emailCont.FindStringSubmatch(email.Cont)
		if cont != nil && len(cont) >= 1 {
			email.Cont = strings.Replace(email.Cont, cont[1], accs.ReplaceBBCodeWithHTML(cont[1]), 1)
		}
		select {
		case mailchan <- email:
		default:
			log.Println("Channel is full. While submitting: " + m.getSubj() + ". Cannot submit message.")
			email.SaveToDBandLog(db, DBType)
		}
	}
}
