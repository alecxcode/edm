package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// text contains only default English strings
type text struct {
	AppTitle           string
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
	ApprovalSign       []string
	TaskStatuses       []string
}

// i18n contains laguage-specific strings
type i18n struct {
	LangCode       string
	LoginLang      LoginLang
	DocCaption     string
	TaskCaption    string
	CommentCaption string
	Categories     []string
	DocTypes       []string
	TaskStatuses   []string
	Messages       AppMessages
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
		NewApproval        string
		Approved           string
		Rejected           string
	}
	Cont struct {
		ProfileRegistered  string
		LoginPasswdChanged string
		ApproveThis        string
		ByApprover         string
		Approved           string
		Rejected           string
		InfoLink           string
	}
	Captions struct {
		CreatedTime      string
		Creator          string
		Assignee         string
		Open             string
		TaskStatus       string
		TaskStartDueTime string
		FileList         string
	}
}

func newTextStruct() text {
	lng := text{
		AppTitle:           "EDM",
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
		Categories: []string{
			"None",
			"Incoming",
			"Outgoing",
			"Internal",
		},
		DocTypes: []string{
			"Other",
			"Letter",
			"Application",
			"Claim",
			"Contract, addendum to it",
			"Accounting, financial",
			"Technical, engineering",
			"Order, regulation",
			"Corporate decision",
			"Minutes, meeting notes",
			"Power of Attorney",
			"Pleading",
			"Judicial (from a court)",
			"Template",
		},
		ApprovalSign: []string{
			"Pending",
			"Approved",
			"Rejected",
			"Broken",
		},
		TaskStatuses: []string{
			"Created",
			"Assigned",
			"In progress",
			"Stuck",
			"Done",
			"Cancelled",
		},
	}
	return lng
}

func newi18nStruct(pathToLngFile string) i18n {
	lng := i18n{}
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
