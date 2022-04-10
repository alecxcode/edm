package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Lang contains all language-different strings
type Lang struct {
	LangNameInt        string
	LangName           string
	LangCode           string
	AppTitle           string
	LoginPageTitle     string
	LoninPrompt        string
	LoninFieldLabel    string
	PasswordFieldLabel string
	LoginButton        string
	WrongLoginMsg      string
	ConfigPageTitle    string
	DocsPageTitle      string
	Document           string
	NewDocument        string
	TeamPageTitle      string
	Profile            string
	NewProfile         string
	CompaniesPageTitle string
	Company            string
	NewCompany         string
	Unit               string
	TasksPageTitle     string
	Task               string
	NewTask            string
	Comment            string
	NewComment         string
	Categories         []string
	DocTypes           []string
	TaskStatuses       []string
	Messages           AppMessages
}

// AppMessages contains language-different strings for email messages
type AppMessages struct {
	MailerName string
	DoNotReply string
	Subj       struct {
		AssigneeSet        string
		AssigneeToSet      string
		NewTaskComment     string
		TaskEdited         string
		TaskStatusChanged  string
		ParticipantToAdded string
		ProfileRegistered  string
		SecurityAlert      string
	}
	Cont struct {
		ProfileRegistered  string
		LoginPasswdChanged string
	}
	Captions struct {
		CreatedTime      string
		Creator          string
		Assignee         string
		Open             string
		TaskStatus       string
		TaskStartDueTime string
		TaskAssignee     string
		FileList         string
	}
}

func newLangStruct(pathToLngFile string) Lang {

	lng := Lang{
		LangNameInt:        "English",
		LangName:           "English",
		LangCode:           "en",
		AppTitle:           "EDM",
		LoginPageTitle:     "System entrance",
		LoninPrompt:        "Please, enter your login and password",
		LoninFieldLabel:    "Login",
		PasswordFieldLabel: "Password",
		LoginButton:        "Log in",
		WrongLoginMsg:      "Wrong login data",
		ConfigPageTitle:    "Settings",
		DocsPageTitle:      "Documents",
		Document:           "Document",
		NewDocument:        "New document",
		TeamPageTitle:      "Team",
		Profile:            "User profile",
		NewProfile:         "New user profile",
		CompaniesPageTitle: "Companies",
		Company:            "Company",
		NewCompany:         "New company",
		Unit:               "Unit",
		TasksPageTitle:     "Tasks",
		Task:               "Task",
		NewTask:            "New task",
		Comment:            "Comment",
		NewComment:         "New comment",
		Categories:         []string{"None", "Incoming", "Outgoing", "Internal"},
		DocTypes: []string{
			"Other",
			"Letter",
			"Applicaton",
			"Claim",
			"Contract, addendum to it",
			"Accounting, financical",
			"Technical, engineering",
			"Order, regulation",
			"Corporate decision",
			"Minutes, meeting notes",
			"Power of Attorney",
			"Pleading",
			"Judicial (from a court)",
			"Template",
		},
		TaskStatuses: []string{
			"Created",
			"Assigned",
			"In progress",
			"Stuck",
			"Done",
			"Cancelled",
		},
		Messages: AppMessages{
			MailerName: "EDM System",
			DoNotReply: "Do not reply to this message, it was sent automatically",
			Subj: struct {
				AssigneeSet        string
				AssigneeToSet      string
				NewTaskComment     string
				TaskEdited         string
				TaskStatusChanged  string
				ParticipantToAdded string
				ProfileRegistered  string
				SecurityAlert      string
			}{
				AssigneeSet:        "New assignee to the task",
				AssigneeToSet:      "You have been assigned to do the task",
				NewTaskComment:     "New comment to the task",
				TaskEdited:         "Task has been edited",
				TaskStatusChanged:  "Task status has been changed",
				ParticipantToAdded: "You have been added to the task participants list",
				ProfileRegistered:  "EDM system access",
				SecurityAlert:      "Security alert",
			},
			Cont: struct {
				ProfileRegistered  string
				LoginPasswdChanged string
			}{
				ProfileRegistered:  "You can login to the system using the link below. Your login and password are: ",
				LoginPasswdChanged: "This is to notify you about your login or password change. If you know why is it happend, then no action required.",
			},
			Captions: struct {
				CreatedTime      string
				Creator          string
				Assignee         string
				Open             string
				TaskStatus       string
				TaskStartDueTime string
				TaskAssignee     string
				FileList         string
			}{
				CreatedTime:      "Created time",
				Creator:          "Creator",
				Assignee:         "Assignee",
				Open:             "Open this in the system",
				TaskStatus:       "Status",
				TaskStartDueTime: "Start and Due",
				TaskAssignee:     "Assignee",
				FileList:         "Files",
			},
		},
	}

	content, err := ioutil.ReadFile(pathToLngFile)
	if err != nil {
		log.Println(currentFunction()+":", err)
	}

	if len(content) > 1 {
		err := json.Unmarshal(content, &lng)
		if err != nil {
			log.Println(currentFunction()+":", err)
		}
	}
	return lng
}

func assignNonEmptyString(d string, s string) string {
	if len(s) == 0 {
		return d
	} else {
		return s
	}
}
