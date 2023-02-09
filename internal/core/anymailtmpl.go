package core

import (
	"bytes"
	"database/sql"
	"edm/internal/mail"
	"html/template"
	"log"
)

// HasEmailAddress defines interface which realisation should return UserToSend type
type HasEmailAddress interface {
	GetUserToSend() mail.UserToSend
}

// AnyMail is a template for any message
type AnyMail struct {
	Subj       string
	Text       string
	SomeLink   string
	DoNotReply string
	SystemURL  string
	MailerName string
}

// GetTemplateHeader provides with email template header
func GetTemplateHeader() string {
	return `<!DOCTYPE html>
<head><meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="ie=edge">
<meta name="format-detection" content="telephone=no">
<title>{{.Subj}}</title>
<style>
:root{font-size:14px;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif;-webkit-font-smoothing:antialiased;-webkit-text-size-adjust:none}*{margin:0;box-sizing:border-box}body{font-size:14px;background-color:#fff;font-family:Roboto,"Segoe UI",Ubuntu,"-apple-system",BlinkMacSystemFont,Arial,sans-serif}
pre,code{font-family:'Courier New',Courier,monospace;overflow-x:auto}q{quotes:'"' '"' "'" "'"}#container{margin:0 auto;background-color:#fff;max-width:1140px;min-width:320px}#control{padding:40px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}#main{padding:10px 20px 0;color:#444;display:block;margin:0;font-size:14px;text-align:left}
a,a.msglnk{text-decoration:none;color:#048}a:visited,a.msglnk:visited{text-decoration:none;color:#048}a:hover,a:focus,a.msglnk:hover,a.msglnk:focus{text-decoration:underline;color:#065;outline:0}.txtred{color:red}.txtgreen{color:#0a0}.txtora{color:#c60}.txtblue{color:#55a}.txtbw{color:#000}.msgredfx{color:red}.msgred{color:red}.msgok{color:#0a0}
#bottom{padding:20px;color:#999;display:block;margin:0;font-size:14px;text-align:center}#bottom a,#bottom a:visited{text-decoration:none;color:#89a}#bottom a:hover,#bottom a:focus{text-decoration:underline;color:#456;outline:0}
h1{color:#706a65;font-size:22px;font-weight:700;margin-bottom:10px}h2{color:#555;font-size:18px;font-weight:700;margin-top:6px;margin-bottom:4px}.afile,#control .afile,#main .afile{color:#100;margin:1px;border-radius:2px;border:1px solid #eeb}.center{display:block;text-align:center}.somemargins{margin-top:6px;margin-bottom:6px}.margintop{margin-top:6px}.marginbottom{margin-bottom:6px}
.sbut{cursor:pointer;text-decoration:none;margin:2px 0;display:inline-block;color:#eee;background-color:#00609e;border:1px solid transparent;border-radius:3px;padding:5px 8px}.sbut:hover,.sbut:focus{color:#fff;background-color:#004c80;outline:0}a.sbut,a.sbut:hover,a.sbut:active,a.sbut:visited,a.sbut:focus{text-decoration:none;color:#fff;display:inline-block;box-sizing:border-box;-webkit-box-sizing:border-box;-moz-box-sizing:border-box;-ms-box-sizing:border-box}
.smaller{padding:2px 4px;border-radius:2px}.nowrap{white-space:nowrap}.inline-block{display:inline-block}#stat{font-size:12px;color:#777}.highlight{color:#000;background-color:#fe0}
</style>
</head>`
}

// GetAnyMailTemplate provides with a complete simple template for mail
func GetAnyMailTemplate() *template.Template {
	tmpl := GetTemplateHeader() + `
<body><div id="container"><div id="main">
<h1>{{.Subj}}</h1>
<div>{{.Text}}</div>
<div class="margintop">{{if .SomeLink}}<a class="msglnk" href="{{.SomeLink}}" target="_blank">{{.SomeLink}}</a>{{end}}</div>
</div><div class="bottom" id="bottom">{{.DoNotReply}}<br>
<a href="{{.SystemURL}}" target="_blank">{{.MailerName}}</a></div></div></body></html>`
	return template.Must(template.New("anymail").Parse(tmpl))
}

// ConstructToChannel writes email to the channel
func (m AnyMail) ConstructToChannel(db *sql.DB, DBType byte, mailtmlp *template.Template, mailchan chan mail.EmailMessage, recepient HasEmailAddress) {
	userToSend := recepient.GetUserToSend()
	if userToSend.Email == "" {
		return
	}
	email := mail.EmailMessage{Subj: m.Subj, SendTo: []mail.UserToSend{userToSend}}
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
