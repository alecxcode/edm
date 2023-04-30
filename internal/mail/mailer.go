package mail

import (
	"database/sql"
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

// MailerMonitor reads from channel email messages and sends them using SMTP
// Undelivered messages are stored in DB
func MailerMonitor(ch chan EmailMessage, host string, port int, user string, passwd string, from string, fromName string, db *sql.DB, DBType byte, DEBUG bool) {
	if port == 0 {
		port = 25
	}
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
				if host != "" {
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
			writeMessage(em, gm, from, fromName)
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
					em.SaveToDBandLog(db, DBType)
				}
				if dbMailerQuited {
					if DEBUG {
						log.Println("Starting DB Mailer, as it is quited before.")
					}
					dbMailerQuited = false
					go ReadMailFromDB(ch, 60, db, DBType, DEBUG)
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
			time.Sleep(time.Duration(10) * time.Millisecond)
		}

	}
}

func writeMessage(appMsg EmailMessage, gomailMsg *gomail.Message, from string, fromName string) {
	emailsTo := make([]string, len(appMsg.SendTo))
	for i, user := range appMsg.SendTo {
		emailsTo[i] = gomailMsg.FormatAddress(user.Email, user.Name)
	}
	emailsCc := make([]string, len(appMsg.SendCc))
	for i, user := range appMsg.SendCc {
		emailsCc[i] = gomailMsg.FormatAddress(user.Email, user.Name)
	}
	gomailMsg.SetHeader("From", gomailMsg.FormatAddress(from, fromName))
	gomailMsg.SetHeader("To", emailsTo...)
	gomailMsg.SetHeader("Cc", emailsCc...)
	gomailMsg.SetHeader("Subject", appMsg.Subj)
	gomailMsg.SetBody("text/html", appMsg.Cont)
}
