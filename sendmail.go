package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"time"

	"github.com/alecxcode/sqla"

	"github.com/go-gomail/gomail"
)

// EmailMessage is processed by mailer subsystem
type EmailMessage struct {
	//sql generate
	ID     int
	SendTo []UserToSend `sql-gen:"varchar(max)"`
	SendCc []UserToSend `sql-gen:"varchar(max)"`
	Subj   string       `sql-gen:"varchar(4000)"`
	Cont   string       `sql-gen:"varchar(max)"`
}

// UserToSend is an addressee of an email
type UserToSend struct {
	Name  string
	Email string
}

func (em *EmailMessage) saveToDB(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	if len(em.SendTo) > 0 {
		json, err := json.Marshal(em.SendTo)
		if err != nil {
			log.Println(currentFunction()+":", err)
		}
		args = args.AppendNonEmptyString("SendTo", string(json))
	}
	if len(em.SendCc) > 0 {
		json, err := json.Marshal(em.SendCc)
		if err != nil {
			log.Println(currentFunction()+":", err)
		}
		args = args.AppendNonEmptyString("SendCc", string(json))
	}
	args = args.AppendNonEmptyString("Subj", em.Subj)
	args = args.AppendNonEmptyString("Cont", em.Cont)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "emailmessages", args)
	return lastid, rowsaff
}

func readMailFromDB(ch chan EmailMessage, waitTimeSeconds int, db *sql.DB, DBType byte) {
	for {
		sq := "SELECT ID, SendTo, SendCc, Subj, Cont FROM emailmessages ORDER BY ID ASC LIMIT 100"
		if DBType == sqla.ORACLE {
			sq = "SELECT ID, SendTo, SendCc, Subj, Cont FROM emailmessages ORDER BY ID ASC FETCH FIRST 100 ROWS ONLY"
		} else if DBType == sqla.MSSQL {
			sq = "SELECT TOP 100 ID, SendTo, SendCc, Subj, Cont FROM emailmessages ORDER BY ID ASC"
		}
		sqcount := "SELECT COUNT(*) FROM emailmessages"
		if DEBUG {
			log.Println(sq, sqcount)
			var UnsentMailNum sql.NullInt64
			row := db.QueryRow(sqcount)
			err := row.Scan(&UnsentMailNum)
			if err != nil {
				log.Println(currentFunction()+":", err)
			}
			log.Println("Looking for unsent email, found emails in DB:", UnsentMailNum.Int64)
		}
		rows, err := db.Query(sq)
		if err != nil {
			log.Println(currentFunction()+":", err)
		}
		defer rows.Close()
		var SendTo sql.NullString
		var SendCc sql.NullString
		var Subj sql.NullString
		var Cont sql.NullString
		i := 0
		for rows.Next() {
			var em EmailMessage
			err = rows.Scan(&em.ID, &SendTo, &SendCc, &Subj, &Cont)
			if err != nil {
				log.Println(currentFunction()+":", err)
			}
			if SendTo.String != "" {
				err := json.Unmarshal([]byte(SendTo.String), &em.SendTo)
				if err != nil {
					log.Println(currentFunction()+":", err)
				}
			}
			if SendCc.String != "" {
				err := json.Unmarshal([]byte(SendCc.String), &em.SendCc)
				if err != nil {
					log.Println(currentFunction()+":", err)
				}
			}
			em.Subj = Subj.String
			em.Cont = Cont.String
			if len(em.SendTo) > 0 || len(em.SendCc) > 0 {
				ch <- em
			}
			i++
		}
		if i == 0 {
			if DEBUG {
				log.Println("No mail in DB to send, quiting DB Mailer routine")
			}
			ch <- EmailMessage{ID: 0, Subj: "DB_MAILER_GOROUTINE_QUITED"}
			return
		}
		time.Sleep(time.Duration(waitTimeSeconds) * time.Second)
	}
}

func mailerMonitor(ch chan EmailMessage, host string, port int, user string, passwd string, from string, db *sql.DB, DBType byte) {
	d := gomail.NewDialer(host, port, user, passwd)
	gm := gomail.NewMessage()
	var sc gomail.SendCloser
	var err error
	open := false
	dbMailerQuited := false
	timer := time.Now()
	for {
		errsending := false
		select {
		case em, ok := <-ch:
			if !ok {
				if open {
					if err := sc.Close(); err != nil {
						log.Println("Error closing SMTP connection:", err)
					}
				}
				return
			}
			if DEBUG {
				log.Printf("Email to send: To:%+v, Cc:%+v, Subject:%+v\n", em.SendTo, em.SendCc, em.Subj)
			}
			if em.ID == 0 && em.Subj == "DB_MAILER_GOROUTINE_QUITED" {
				dbMailerQuited = true
				continue
			}
			if !open {
				if host != "" || user != "" || passwd != "" {
					if DEBUG {
						log.Println("Dialing SMTP connection...")
					}
					if sc, err = d.Dial(); err != nil {
						log.Println("Error opening SMTP connection:", err)
					}
					if err == nil {
						open = true
					}
				} else {
					if DEBUG {
						log.Println("SMTP not configured, cannot send mail, dropping message.")
					}
					continue
				}
			}
			writeMessage(em, gm, from)
			if open {
				err = gomail.Send(sc, gm)
				if err != nil {
					errsending = true
					log.Println("Error sending mail message:", err)
				} else {
					if em.ID != 0 {
						sqla.DeleteObject(db, DBType, "emailmessages", "ID", em.ID)
					}
				}
			} else {
				log.Println("Cannot send mail message: connection is closed.")
			}
			if !open || errsending {
				if em.ID == 0 {
					log.Println("Saving mail message to DB.")
					_, ra := em.saveToDB(db, DBType)
					if ra < 1 {
						log.Println("Cannot save mail message to DB, DB error.")
					}
				}
				if dbMailerQuited {
					if DEBUG {
						log.Println("Starting DB Mailer, as it is quited before.")
					}
					dbMailerQuited = false
					go readMailFromDB(ch, 60, db, DBType)
				}
			}
			gm.Reset()
			timer = time.Now()
		default:
			// Close the SMTP server connection if no email was sent in the last x seconds.
			if open && time.Now().Sub(timer).Seconds() > 10 {
				if DEBUG {
					log.Println("Closing SMTP connection...")
				}
				if err := sc.Close(); err != nil {
					log.Println("Error closing SMTP connection:", err)
				}
				open = false
			}
			// Wait a bit.
			time.Sleep(time.Duration(100) * time.Millisecond)
		}

	}
}

func writeMessage(appMsg EmailMessage, gomailMsg *gomail.Message, from string) {
	emailsTo := make([]string, len(appMsg.SendTo))
	for i, user := range appMsg.SendTo {
		emailsTo[i] = gomailMsg.FormatAddress(user.Email, user.Name)
	}
	emailsCc := make([]string, len(appMsg.SendCc))
	for i, user := range appMsg.SendCc {
		emailsCc[i] = gomailMsg.FormatAddress(user.Email, user.Name)
	}
	gomailMsg.SetHeader("From", gomailMsg.FormatAddress(from, "EDM System"))
	gomailMsg.SetHeader("To", emailsTo...)
	gomailMsg.SetHeader("Cc", emailsCc...)
	gomailMsg.SetHeader("Subject", appMsg.Subj)
	gomailMsg.SetBody("text/html", appMsg.Cont)
}

/*
func sendmail(host string, port int, user string, passwd string, from string, to []string, message string) {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(from, "EDM System"))
	m.SetHeader("To", to...)
	//m.SetAddressHeader("Cc", "user@example.com", "User")
	m.SetHeader("Subject", "EDM System message")
	m.SetBody("text/html", message)
	//m.Attach("/home/user/testfile.txt")
	d := gomail.NewDialer(host, port, user, passwd)
	if err := d.DialAndSend(m); err != nil {
		log.Println("Sending single mail:", err)
	}
}
*/

// AnyMail is a template for any message
type AnyMail struct {
	Subj       string
	Text       string
	SomeLink   string
	DoNotReply string
	SystemURL  string
	MailerName string
}

func getAnyMailTemplate() *template.Template {
	tmpl := `<!DOCTYPE html>
<head><meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="ie=edge">
<meta name="format-detection" content="telephone=no">
<title>{{.Subj}}</title>
<style>
:root{font-size:14px;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif;-webkit-font-smoothing:antialiased;-webkit-text-size-adjust:none}*{margin:0;box-sizing:border-box}body{font-size:14px;background-color:#fff;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif}pre,code{font-family:'Courier New',Courier,monospace;overflow-x:auto}#container{margin:0 auto;background-color:#fff;max-width:1140px;min-width:320px}#control{padding:40px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}#main{padding:10px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}a,a.msglnk{text-decoration:none;color:#048}a:visited,a.msglnk:visited{text-decoration:none;color:#048}a:hover,a:focus,a.msglnk:hover,a.msglnk:focus{text-decoration:underline;color:#065;outline:0}.txtred{color:red}.txtgreen{color:#0a0}.txtblue{color:#55a}.txtbw{color:#000}.msgredfx{color:red}.msgred{color:red}.msgok{color:#0a0}
#bottom{padding:20px;color:#999;display:block;margin:0;font-size:14px;text-align:center}#bottom a,#bottom a:visited{text-decoration:none;color:#89a}#bottom a:hover,#bottom a:focus{text-decoration:underline;color:#456;outline:0}h1{color:#706a65;font-size:22px;font-weight:700;margin-bottom:10px}h2{color:#555;font-size:18px;font-weight:700;margin-top:6px;margin-bottom:4px}.afile,#control .afile,#main .afile{color:#100;margin:1px;border-radius:2px;border:1px solid #eeb}.center{display:block;text-align:center}.somemargins{margin-top:6px;margin-bottom:6px}.margintop{margin-top:6px}.marginbottom{margin-top:6px}.sbut{cursor:pointer;text-decoration:none;margin:2px 0;display:inline-block;color:#eee;background-color:#258;border:1px solid #333;border-radius:2px;padding:5px 8px;box-shadow:1px 1px 2px 0px #777}.sbut:hover,.sbut:focus{color:#fff;background-color:#48e;box-shadow:1px 1px 2px 0px #888;border:1px solid #004;outline:0}a.sbut,a.sbut:hover,a.sbut:active,a.sbut:visited,a.sbut:focus{text-decoration:none;color:#fff;display:inline-block;box-sizing:border-box;-webkit-box-sizing:border-box;-moz-box-sizing:border-box;-ms-box-sizing:border-box}
.smaller{padding:2px 4px}.nowrap{white-space:nowrap}.inline-block{display:inline-block}#stat{font-size:12px;color:#777}.highlight{color:#000;background-color:#fe0}
</style>
</head>
<body><div id="container"><div id="main">
<h1>{{.Subj}}</h1>
<div>{{.Text}}</div>
<div class="margintop">{{if .SomeLink}}<a class="msglnk" href="{{.SomeLink}}" target="_blank">{{.SomeLink}}</a>{{end}}</div>
</div><div class="bottom" id="bottom">{{.DoNotReply}}<br>
<a href="{{.SystemURL}}" target="_blank">{{.MailerName}}</a></div></div></body></html>`
	return template.Must(template.New("anymail").Parse(tmpl))
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
	tmpl := `<!DOCTYPE html>
<head><meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="ie=edge">
<meta name="format-detection" content="telephone=no">
<title>{{.Subj}}</title>
<style>
:root{font-size:14px;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif;-webkit-font-smoothing:antialiased;-webkit-text-size-adjust:none}*{margin:0;box-sizing:border-box}body{font-size:14px;background-color:#fff;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif}pre,code{font-family:'Courier New',Courier,monospace;overflow-x:auto}#container{margin:0 auto;background-color:#fff;max-width:1140px;min-width:320px}#control{padding:40px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}#main{padding:10px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}a,a.msglnk{text-decoration:none;color:#048}a:visited,a.msglnk:visited{text-decoration:none;color:#048}a:hover,a:focus,a.msglnk:hover,a.msglnk:focus{text-decoration:underline;color:#065;outline:0}.txtred{color:red}.txtgreen{color:#0a0}.txtblue{color:#55a}.txtbw{color:#000}.msgredfx{color:red}.msgred{color:red}.msgok{color:#0a0}
#bottom{padding:20px;color:#999;display:block;margin:0;font-size:14px;text-align:center}#bottom a,#bottom a:visited{text-decoration:none;color:#89a}#bottom a:hover,#bottom a:focus{text-decoration:underline;color:#456;outline:0}h1{color:#706a65;font-size:22px;font-weight:700;margin-bottom:10px}h2{color:#555;font-size:18px;font-weight:700;margin-top:6px;margin-bottom:4px}.afile,#control .afile,#main .afile{color:#100;margin:1px;border-radius:2px;border:1px solid #eeb}.center{display:block;text-align:center}.somemargins{margin-top:6px;margin-bottom:6px}.margintop{margin-top:6px}.marginbottom{margin-top:6px}.sbut{cursor:pointer;text-decoration:none;margin:2px 0;display:inline-block;color:#eee;background-color:#258;border:1px solid #333;border-radius:2px;padding:5px 8px;box-shadow:1px 1px 2px 0px #777}.sbut:hover,.sbut:focus{color:#fff;background-color:#48e;box-shadow:1px 1px 2px 0px #888;border:1px solid #004;outline:0}a.sbut,a.sbut:hover,a.sbut:active,a.sbut:visited,a.sbut:focus{text-decoration:none;color:#fff;display:inline-block;box-sizing:border-box;-webkit-box-sizing:border-box;-moz-box-sizing:border-box;-ms-box-sizing:border-box}
.smaller{padding:2px 4px}.nowrap{white-space:nowrap}.inline-block{display:inline-block}#stat{font-size:12px;color:#777}.highlight{color:#000;background-color:#fe0}
</style>
</head>
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
<div><a class="sbut" href="{{$.SystemURL}}/tasks/task/{{.Task.ID}}" target="_blank">{{.AppMessages.Captions.Open}}</a></div>
</div><div class="bottom" id="bottom">{{.AppMessages.DoNotReply}}<br>
<a href="{{.SystemURL}}" target="_blank">{{.AppMessages.MailerName}}</a></div></div></body></html>`
	return template.Must(template.New("taskmail").Parse(tmpl))
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
	tmpl := `<!DOCTYPE html>
<head><meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="ie=edge">
<meta name="format-detection" content="telephone=no">
<title>{{.Subj}}</title>
<style>
:root{font-size:14px;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif;-webkit-font-smoothing:antialiased;-webkit-text-size-adjust:none}*{margin:0;box-sizing:border-box}body{font-size:14px;background-color:#fff;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif}pre,code{font-family:'Courier New',Courier,monospace;overflow-x:auto}#container{margin:0 auto;background-color:#fff;max-width:1140px;min-width:320px}#control{padding:40px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}#main{padding:10px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}a,a.msglnk{text-decoration:none;color:#048}a:visited,a.msglnk:visited{text-decoration:none;color:#048}a:hover,a:focus,a.msglnk:hover,a.msglnk:focus{text-decoration:underline;color:#065;outline:0}.txtred{color:red}.txtgreen{color:#0a0}.txtblue{color:#55a}.txtbw{color:#000}.msgredfx{color:red}.msgred{color:red}.msgok{color:#0a0}
#bottom{padding:20px;color:#999;display:block;margin:0;font-size:14px;text-align:center}#bottom a,#bottom a:visited{text-decoration:none;color:#89a}#bottom a:hover,#bottom a:focus{text-decoration:underline;color:#456;outline:0}h1{color:#706a65;font-size:22px;font-weight:700;margin-bottom:10px}h2{color:#555;font-size:18px;font-weight:700;margin-top:6px;margin-bottom:4px}.afile,#control .afile,#main .afile{color:#100;margin:1px;border-radius:2px;border:1px solid #eeb}.center{display:block;text-align:center}.somemargins{margin-top:6px;margin-bottom:6px}.margintop{margin-top:6px}.marginbottom{margin-top:6px}.sbut{cursor:pointer;text-decoration:none;margin:2px 0;display:inline-block;color:#eee;background-color:#258;border:1px solid #333;border-radius:2px;padding:5px 8px;box-shadow:1px 1px 2px 0px #777}.sbut:hover,.sbut:focus{color:#fff;background-color:#48e;box-shadow:1px 1px 2px 0px #888;border:1px solid #004;outline:0}a.sbut,a.sbut:hover,a.sbut:active,a.sbut:visited,a.sbut:focus{text-decoration:none;color:#fff;display:inline-block;box-sizing:border-box;-webkit-box-sizing:border-box;-moz-box-sizing:border-box;-ms-box-sizing:border-box}
.smaller{padding:2px 4px}.nowrap{white-space:nowrap}.inline-block{display:inline-block}#stat{font-size:12px;color:#777}.highlight{color:#000;background-color:#fe0}
</style>
</head>
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
<div><a class="sbut" href="{{$.SystemURL}}/tasks/task/{{.TaskID}}#comment{{.Comment.ID}}" target="_blank">{{.AppMessages.Captions.Open}}</a></div>
</div><div class="bottom" id="bottom">{{.AppMessages.DoNotReply}}<br>
<a href="{{.SystemURL}}" target="_blank">{{.AppMessages.MailerName}}</a></div></div></body></html>`
	return template.Must(template.New("commmail").Parse(tmpl))
}
