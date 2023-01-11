package projs

import (
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// ProjectPage is passed into template
type ProjectPage struct {
	AppTitle      string
	AppVersion    string
	PageTitle     string
	LoggedinID    int
	UserConfig    team.UserConfig
	Project       Project //payload
	Message       string
	Editable      bool
	LoggedinAdmin bool
	RemoveAllowed bool
	New           bool
	ProjStatuses  []string
	TaskStatuses  []string
	UserList      []memdb.ObjHasID
}

// ProjectHandler is http handler for project page
func (pb *ProjsBase) ProjectHandler(w http.ResponseWriter, r *http.Request) {
	const ADMIN = 1

	allow, id := core.AuthVerify(w, r, pb.memorydb)
	if !allow {
		return
	}

	if pb.validURLs.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = ProjectPage{
		AppTitle:     pb.text.AppTitle,
		AppVersion:   core.AppVersion,
		LoggedinID:   id,
		Editable:     false,
		New:          false,
		ProjStatuses: pb.text.ProjStatuses,
		TaskStatuses: pb.text.TaskStatuses,
	}

	TextID := accs.GetTextIDfromURL(r.URL.Path)
	IntID, _ := strconv.Atoi(TextID)
	if TextID == "new" {
		Page.New = true
	} else {
		Page.Project = Project{ID: IntID}
		err = Page.Project.load(pb.db, pb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "loading project tasks", Page.LoggedinID, Page.Project.ID)
			return
		}
	}

	user := team.UnmarshalToProfile(pb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == team.ADMIN || Page.New || Page.Project.GiveCreatorID() == Page.LoggedinID {
		Page.Editable = true
	}
	if user.UserRole == team.ADMIN || (Page.Project.GiveCreatorID() == Page.LoggedinID && pb.removeAllowed) {
		Page.RemoveAllowed = true
	}
	if user.UserRole == team.ADMIN {
		Page.LoggedinAdmin = true
	}

	var created int
	var updated int

	// Create or update code ==========================================================================
	if r.Method == "POST" && (r.FormValue("createButton") != "" || r.FormValue("updateButton") != "") {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "writing project", Page.LoggedinID, IntID)
			return
		}
		p := Project{
			ID:          IntID,
			ProjName:    r.FormValue("projName"),
			Description: r.FormValue("description"),
		}
		if accs.StrToInt(r.FormValue("owner")) != core.Undefined {
			p.Creator = &team.Profile{ID: accs.StrToInt(r.FormValue("owner"))}
		}
		if r.FormValue("createButton") != "" {
			p.Creator = &team.Profile{ID: Page.LoggedinID}
			p.ID, created = p.Create(pb.db, pb.dbType)
			if created > 0 {
				if Page.UserConfig.ReturnAfterCreation {
					http.Redirect(w, r, "/projs/", http.StatusSeeOther)
				} else {
					http.Redirect(w, r, fmt.Sprintf("/projs/project/%d", p.ID), http.StatusSeeOther)
				}
				return
			} else {
				Page.Message = "dataNotWritten"
			}
		}
		if r.FormValue("updateButton") != "" && !strings.Contains(Page.Message, "Error") {
			updated = p.update(pb.db, pb.dbType)
			if updated > 0 {
				Page.Message = "dataWritten"
				Page.Project.load(pb.db, pb.dbType)
			} else {
				Page.Message = "dataNotWritten"
			}
		}
	}

	// Other fields code ============================================
	Page.UserList = pb.memorydb.GetObjectArr("UserList")
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = pb.text.NewProject
	} else {
		if Page.Project.ProjName != "" {
			Page.PageTitle += pb.text.Project + ": " + Page.Project.ProjName
		} else {
			Page.PageTitle = pb.text.Project + " #" + TextID
		}
	}

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		return
	}

	// HTML output
	err = pb.templates.ExecuteTemplate(w, "project.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		accs.ThrowServerError(w, "executing project template", Page.LoggedinID, Page.Project.ID)
		return
	}

}
