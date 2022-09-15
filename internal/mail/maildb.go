package mail

import (
	"database/sql"
	"edm/pkg/accs"
	"encoding/json"
	"log"
	"time"

	"github.com/alecxcode/sqla"
)

func (em *EmailMessage) saveToDB(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	if len(em.SendTo) > 0 {
		json, err := json.Marshal(em.SendTo)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
		args = args.AppendNonEmptyString("SendTo", string(json))
	}
	if len(em.SendCc) > 0 {
		json, err := json.Marshal(em.SendCc)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
		args = args.AppendNonEmptyString("SendCc", string(json))
	}
	args = args.AppendNonEmptyString("Subj", em.Subj)
	args = args.AppendNonEmptyString("Cont", em.Cont)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "emailmessages", args)
	return lastid, rowsaff
}

func (em *EmailMessage) SaveToDBandLog(db *sql.DB, DBType byte) {
	log.Println("Saving mail message to DB.")
	_, ra := em.saveToDB(db, DBType)
	if ra < 1 {
		log.Println("Cannot save mail message to DB, DB error.")
	}
}

func ReadMailFromDB(ch chan EmailMessage, waitTimeSeconds int, db *sql.DB, DBType byte, DEBUG bool) {
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
				log.Println(accs.CurrentFunction()+":", err)
			}
			log.Println("Looking for unsent email, found emails in DB:", UnsentMailNum.Int64)
		}
		rows, err := db.Query(sq)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
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
				log.Println(accs.CurrentFunction()+":", err)
			}
			if SendTo.String != "" {
				err := json.Unmarshal([]byte(SendTo.String), &em.SendTo)
				if err != nil {
					log.Println(accs.CurrentFunction()+":", err)
				}
			}
			if SendCc.String != "" {
				err := json.Unmarshal([]byte(SendCc.String), &em.SendCc)
				if err != nil {
					log.Println(accs.CurrentFunction()+":", err)
				}
			}
			em.Subj = Subj.String
			em.Cont = Cont.String
			if len(em.SendTo) > 0 || len(em.SendCc) > 0 {
				select {
				case ch <- em:
				default:
					log.Println("Channel is full. DB Mailer cannot send message.")
				}
			}
			i++
		}
		if i == 0 {
			if DEBUG {
				log.Println("No mail in DB to send, quiting DB Mailer routine")
			}
			select {
			case ch <- EmailMessage{ID: 0, Subj: "DB_MAILER_GOROUTINE_QUITED"}:
			default:
				log.Println("Channel is full. DB Mailer cannot send quit message.")
			}
			return
		}
		time.Sleep(time.Duration(waitTimeSeconds) * time.Second)
	}
}
