package core

import (
	"edm/pkg/accs"
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
)

// Text contains only default English strings
type Text struct {
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
	ProjsPageTitle     string
	Project            string
	NewProject         string
	Categories         []string
	DocTypes           []string
	ApprovalSign       []string
	TaskStatuses       []string
	ProjStatuses       []string
}

// Si18n contains server laguage-specific strings
type Si18n struct {
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

// LoginLang applies to a login template
type LoginLang struct {
	AppTitle           string
	LoginPageTitle     string
	LoninPrompt        string
	LoninFieldLabel    string
	PasswordFieldLabel string
	LoginButton        string
	WrongLoginMsg      string
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

// NewTextStruct is a constructor for Text struct
func NewTextStruct() Text {
	lng := Text{
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
		ProjsPageTitle:     "Projects",
		Project:            "Project",
		NewProject:         "New Project",
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
			"Canceled",
			"In review",
		},
		ProjStatuses: []string{
			"Active",
			"Done",
			"Canceled",
		},
	}
	return lng
}

// Newi18nStruct is a constructor for Si18n struct
func Newi18nStruct(pathToLngFile string) Si18n {
	lng := Si18n{}
	content, err := ioutil.ReadFile(pathToLngFile)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
	}
	if len(content) > 1 {
		err := json.Unmarshal(content, &lng)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
		}
	}
	return lng
}

// AssignNonEmptyString returns new string if it is now empty, otherwise returns old string
func AssignNonEmptyString(d string, s string) string {
	if len(s) == 0 {
		return d
	}
	return s
}

// GetLangList loads a list of available frontend languages
func GetLangList(serverSystem string) []string {
	files, err := ioutil.ReadDir(filepath.Join(serverSystem, "static", "i18n"))
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		return []string{}
	}
	var res []string
	var fname string
	for _, file := range files {
		fname = file.Name()
		if ext := filepath.Ext(fname); ext == ".json" {
			fname = fname[0 : len(fname)-len(ext)]
			res = append(res, fname)
		}
	}
	return res
}
