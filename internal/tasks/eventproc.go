package tasks

import (
	"edm/internal/mail"
	"edm/internal/team"
	"strconv"
)

func (tb *TasksBase) taskEventProcessing(
	t Task,
	participants []team.Profile,
	comments []Comment,
	eventAssigneeSet bool,
	eventTaskEdited bool,
	eventTaskStatusChanged bool,
	eventParticipantToAdded bool,
	newParticipantID int,
	eventNewTaskComment bool,
	newCommentID int) {
	// Email messages ================================================
	if eventTaskEdited || eventTaskStatusChanged || eventAssigneeSet {
		email := mail.EmailMessage{}
		if eventTaskEdited {
			email.Subj = tb.i18n.messages.Subj.TaskEdited + " [" + tb.i18n.taskCaption + " #" + strconv.Itoa(t.ID) + "]"
		} else if eventAssigneeSet {
			email.Subj = tb.i18n.messages.Subj.AssigneeSet + " [" + tb.i18n.taskCaption + " #" + strconv.Itoa(t.ID) + "]"
		} else if eventTaskStatusChanged {
			email.Subj = tb.i18n.messages.Subj.TaskStatusChanged + " [" + tb.i18n.taskCaption + " #" + strconv.Itoa(t.ID) + "]"
		}
		if t.Creator != nil && t.Creator.Contacts.Email != "" && t.Creator.UserLock == 0 {
			email.SendTo = append(email.SendTo, mail.UserToSend{Name: t.Creator.FirstName + " " + t.Creator.Surname, Email: t.Creator.Contacts.Email})
		}
		if t.Assignee != nil && t.Assignee.Contacts.Email != "" && !eventAssigneeSet && t.Assignee.UserLock == 0 {
			email.SendTo = append(email.SendTo, mail.UserToSend{Name: t.Assignee.FirstName + " " + t.Assignee.Surname, Email: t.Assignee.Contacts.Email})
		}
		for i := 0; i < len(participants); i++ {
			if participants[i].Contacts.Email != "" && participants[i].UserLock == 0 {
				email.SendCc = append(email.SendCc, mail.UserToSend{Name: participants[i].FirstName + " " + participants[i].Surname, Email: participants[i].Contacts.Email})
			}
		}
		taskMail := TaskMail{email.Subj, t, tb.i18n.messages, tb.i18n.taskCaption, tb.i18n.taskStatuses, tb.cfg.systemURL}
		taskMail.constructToChannel(tb.db, tb.dbType, tb.taskmailtmpl, tb.mailchan, email, tb.emailCont)
	}

	if eventAssigneeSet {
		email := mail.EmailMessage{Subj: tb.i18n.messages.Subj.AssigneeToSet + " [" + tb.i18n.taskCaption + " #" + strconv.Itoa(t.ID) + "]"}
		if t.Assignee != nil && t.Assignee.Contacts.Email != "" && t.Assignee.UserLock == 0 {
			email.SendTo = append(email.SendTo, mail.UserToSend{Name: t.Assignee.FirstName + " " + t.Assignee.Surname, Email: t.Assignee.Contacts.Email})
		}
		taskMail := TaskMail{email.Subj, t, tb.i18n.messages, tb.i18n.taskCaption, tb.i18n.taskStatuses, tb.cfg.systemURL}
		taskMail.constructToChannel(tb.db, tb.dbType, tb.taskmailtmpl, tb.mailchan, email, tb.emailCont)
	}

	if eventParticipantToAdded {
		email := mail.EmailMessage{Subj: tb.i18n.messages.Subj.ParticipantToAdded + " [" + tb.i18n.taskCaption + " #" + strconv.Itoa(t.ID) + "]"}
		for i := 0; i < len(participants); i++ {
			if participants[i].ID == newParticipantID && participants[i].Contacts.Email != "" && participants[i].UserLock == 0 {
				email.SendTo = append(email.SendTo, mail.UserToSend{Name: participants[i].FirstName + " " + participants[i].Surname, Email: participants[i].Contacts.Email})
				break
			}
		}
		taskMail := TaskMail{email.Subj, t, tb.i18n.messages, tb.i18n.taskCaption, tb.i18n.taskStatuses, tb.cfg.systemURL}
		taskMail.constructToChannel(tb.db, tb.dbType, tb.taskmailtmpl, tb.mailchan, email, tb.emailCont)
	}

	if eventNewTaskComment {
		email := mail.EmailMessage{Subj: tb.i18n.messages.Subj.NewTaskComment + " [" + tb.i18n.taskCaption + " #" + strconv.Itoa(t.ID) + "]"}
		if t.Creator != nil && t.Creator.Contacts.Email != "" && t.Creator.UserLock == 0 {
			email.SendTo = append(email.SendTo, mail.UserToSend{Name: t.Creator.FirstName + " " + t.Creator.Surname, Email: t.Creator.Contacts.Email})
		}
		if t.Assignee != nil && t.Assignee.Contacts.Email != "" && t.Assignee.UserLock == 0 {
			email.SendTo = append(email.SendTo, mail.UserToSend{Name: t.Assignee.FirstName + " " + t.Assignee.Surname, Email: t.Assignee.Contacts.Email})
		}
		for i := 0; i < len(participants); i++ {
			if participants[i].Contacts.Email != "" && participants[i].UserLock == 0 {
				email.SendCc = append(email.SendCc, mail.UserToSend{Name: participants[i].FirstName + " " + participants[i].Surname, Email: participants[i].Contacts.Email})
			}
		}
		if len(email.SendTo) > 0 || len(email.SendCc) > 0 {
			commMail := CommMail{email.Subj, t.ID, t.Topic, Comment{ID: 0}, 0, tb.i18n.messages, tb.i18n.taskCaption, tb.i18n.commentCaption, tb.cfg.systemURL}
			for i := 0; i < len(comments); i++ {
				if comments[i].ID == newCommentID {
					commMail.Comment = comments[i]
					commMail.CommentIndex = i + 1
					break
				}
			}
			commMail.constructToChannel(tb.db, tb.dbType, tb.commmailtmpl, tb.mailchan, email, tb.emailCont)
		}
	}
}
