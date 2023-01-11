package team

import (
	"database/sql"
	"edm/internal/mail"
)

// GetID returns Profile ID to satisfy some interfaces
func (p Profile) GetID() int {
	return p.ID
}

// GetUserToSend implements interface to get UserToSend for emails
func (p Profile) GetUserToSend() mail.UserToSend {
	if p.Contacts.Email == "" || p.UserLock != 0 {
		return mail.UserToSend{Name: "", Email: ""}
	}
	return mail.UserToSend{Name: p.FirstName + " " + p.Surname, Email: p.Contacts.Email}
}

// CreateFirstAdmin called after the program creates DB tables to create the first user profile
func CreateFirstAdmin(db *sql.DB, DBType byte, langCode string) {
	admin := Profile{
		UserRole:  1,
		Login:     "admin",
		FirstName: "Administrator",
		Passwd:    "",
		UserConfig: UserConfig{
			SystemTheme:          "dark",
			ElemsOnPage:          20,
			ElemsOnPageTeam:      500,
			DateFormat:           "dd.mm.yyyy",
			TimeFormat:           "24h",
			LangCode:             langCode,
			UseCalendarInConrols: true,
			CurrencyBeforeAmount: true,
			ShowFinishedTasks:    true,
			ReturnAfterCreation:  true,
		},
	}
	uniqueLogin, _ := admin.isLoginUniqueorBlank(db, DBType)
	if uniqueLogin {
		admin.Create(db, DBType)
	}
}
