package main

import (
	"bytes"
	"database/sql"
	"edm/internal/mail"
	"edm/pkg/accs"
	"html/template"
	"log"
	"regexp"
	"strings"
)

// AnyMail is a template for any message
type AnyMail struct {
	Subj       string
	Text       string
	SomeLink   string
	DoNotReply string
	SystemURL  string
	MailerName string
}

func getTemplateHeader() string {
	return `<!DOCTYPE html>
<head><meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="ie=edge">
<meta name="format-detection" content="telephone=no">
<title>{{.Subj}}</title>
<style>
:root{font-size:14px;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif;-webkit-font-smoothing:antialiased;-webkit-text-size-adjust:none}*{margin:0;box-sizing:border-box}body{font-size:14px;background-color:#fff;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif}
pre,code{font-family:'Courier New',Courier,monospace;overflow-x:auto}q{quotes:'"' '"' "'" "'"}#container{margin:0 auto;background-color:#fff;max-width:1140px;min-width:320px}#control{padding:40px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}#main{padding:10px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}
a,a.msglnk{text-decoration:none;color:#048}a:visited,a.msglnk:visited{text-decoration:none;color:#048}a:hover,a:focus,a.msglnk:hover,a.msglnk:focus{text-decoration:underline;color:#065;outline:0}.txtred{color:red}.txtgreen{color:#0a0}.txtblue{color:#55a}.txtbw{color:#000}.msgredfx{color:red}.msgred{color:red}.msgok{color:#0a0}
#bottom{padding:20px;color:#999;display:block;margin:0;font-size:14px;text-align:center}#bottom a,#bottom a:visited{text-decoration:none;color:#89a}#bottom a:hover,#bottom a:focus{text-decoration:underline;color:#456;outline:0}
h1{color:#706a65;font-size:22px;font-weight:700;margin-bottom:10px}h2{color:#555;font-size:18px;font-weight:700;margin-top:6px;margin-bottom:4px}.afile,#control .afile,#main .afile{color:#100;margin:1px;border-radius:2px;border:1px solid #eeb}.center{display:block;text-align:center}.somemargins{margin-top:6px;margin-bottom:6px}.margintop{margin-top:6px}.marginbottom{margin-bottom:6px}
.sbut{cursor:pointer;text-decoration:none;margin:2px 0;display:inline-block;color:#eee;background-color:#258;border:1px solid #333;border-radius:2px;padding:5px 8px;box-shadow:1px 1px 2px 0px #777}.sbut:hover,.sbut:focus{color:#fff;background-color:#48e;box-shadow:1px 1px 2px 0px #888;border:1px solid #004;outline:0}a.sbut,a.sbut:hover,a.sbut:active,a.sbut:visited,a.sbut:focus{text-decoration:none;color:#fff;display:inline-block;box-sizing:border-box;-webkit-box-sizing:border-box;-moz-box-sizing:border-box;-ms-box-sizing:border-box}
.smaller{padding:2px 4px}.nowrap{white-space:nowrap}.inline-block{display:inline-block}#stat{font-size:12px;color:#777}.highlight{color:#000;background-color:#fe0}
</style>
</head>`
}

func getAnyMailTemplate() *template.Template {
	tmpl := getTemplateHeader() + `
<body><div id="container"><div id="main">
<h1>{{.Subj}}</h1>
<div>{{.Text}}</div>
<div class="margintop">{{if .SomeLink}}<a class="msglnk" href="{{.SomeLink}}" target="_blank">{{.SomeLink}}</a>{{end}}</div>
</div><div class="bottom" id="bottom">{{.DoNotReply}}<br>
<a href="{{.SystemURL}}" target="_blank">{{.MailerName}}</a></div></div></body></html>`
	return template.Must(template.New("anymail").Parse(tmpl))
}

func (m AnyMail) constructToChannel(db *sql.DB, DBType byte, mailtmlp *template.Template, mailchan chan mail.EmailMessage, recepient Profile) {
	if recepient.Contacts.Email == "" || recepient.UserLock != 0 {
		return
	}
	email := mail.EmailMessage{Subj: m.Subj, SendTo: []mail.UserToSend{{Name: recepient.FirstName + " " + recepient.Surname, Email: recepient.Contacts.Email}}}
	if len(email.SendTo) > 0 || len(email.SendCc) > 0 {
		newMail := AnyMail{m.Subj, m.Text, m.SomeLink, m.DoNotReply, m.SystemURL, m.MailerName}
		var tmpl bytes.Buffer
		if err := mailtmlp.Execute(&tmpl, newMail); err != nil {
			log.Println("executing any mail template: "+m.Subj+":", err)
		}
		email.Cont = tmpl.String()
		select {
		case mailchan <- email:
		default:
			log.Println("Channel is full. While submitting: " + m.Subj + ". Cannot submit message.")
			email.SaveToDBandLog(db, DBType)
		}
	}
}

type bbcodedMail interface {
	getSubj() string
	constructToChannel(db *sql.DB, DBType byte, mailtmlp *template.Template, mailchan chan mail.EmailMessage, email mail.EmailMessage, emailCont *regexp.Regexp)
}

// TaskMail is a template for a task-related message
type TaskMail struct {
	Subj         string
	Task         Task
	AppMessages  AppMessages
	TaskCaption  string
	TaskStatuses []string
	SystemURL    string
}

func getTaskMailTemplate() *template.Template {
	tmpl := getTemplateHeader() + `
<body><div id="container"><div id="main">
<h1>{{.Subj}}</h1>{{$ci := .Task.GiveCreatorID}}{{$ai := .Task.GiveAssigneeID}}{{$statusindicator := "txtblue"}}
{{if eq .Task.TaskStatus 3}}{{$statusindicator = "txtred"}}{{else if eq .Task.TaskStatus 4}}{{$statusindicator = "txtgreen"}}{{else if eq .Task.TaskStatus 5}}{{$statusindicator = "txtbw"}}{{end}}
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
	AppMessages    AppMessages
	TaskCaption    string
	CommentCaption string
	SystemURL      string
}

func getCommMailTemplate() *template.Template {
	tmpl := getTemplateHeader() + `
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
