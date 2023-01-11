package projs

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"edm/pkg/datetime"
	"edm/pkg/memdb"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/alecxcode/sqla"
)

// ProjsPage is passed into template
type ProjsPage struct {
	AppTitle      string
	AppVersion    string
	PageTitle     string
	LoggedinID    int
	LoggedinAdmin bool
	RemoveAllowed bool
	Message       string
	UserConfig    team.UserConfig
	Projects      []Project //payload
	Filters       sqla.Filter
	PageNumber    int
	FilteredNum   int
	RemovedNum    int
	ProjStatuses  []string
	UserList      []memdb.ObjHasID
}

// ProjsHandler is http handler for projects page
func (pb *ProjsBase) ProjsHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, pb.memorydb)
	if !allow {
		return
	}

	if pb.validURLs.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = ProjsPage{
		AppTitle:     pb.text.AppTitle,
		AppVersion:   core.AppVersion,
		PageTitle:    pb.text.ProjsPageTitle,
		ProjStatuses: pb.text.ProjStatuses,
		UserList:     pb.memorydb.GetObjectArr("UserList"),
		Filters: sqla.Filter{
			ClassFilter: []sqla.ClassFilter{
				{Name: "creators", Column: "Creator"},
				{Name: "projstatuses", Column: "ProjStatus"},
			},
			TextFilterName:    "searchText",
			TextFilterColumns: []string{"ProjName", "Description"},
		},
		PageNumber: 1,
		LoggedinID: id,
	}

	user := team.UnmarshalToProfile(pb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == team.ADMIN || pb.removeAllowed {
		Page.RemoveAllowed = true
	}
	if user.UserRole == team.ADMIN {
		Page.LoggedinAdmin = true
	}

	Page.Filters.GetFilterFromForm(r,
		datetime.ConvDateStrToInt64, datetime.ConvDateTimeStrToInt64,
		map[string]int{"my": Page.LoggedinID})

	// Pagination
	if r.FormValue("elemsOnPage") != "" {
		Page.UserConfig.ElemsOnPage, _ = strconv.Atoi(r.FormValue("elemsOnPage"))
		if r.FormValue("elemsOnPageChanged") == "true" {
			p := team.Profile{ID: Page.LoggedinID, UserConfig: Page.UserConfig}
			updated := p.UpdateConfig(pb.db, pb.dbType)
			if updated > 0 {
				team.MemoryUpdateProfile(pb.db, pb.dbType, pb.memorydb, p.ID)
			}
		}
	}
	if r.FormValue("pageNumber") != "" {
		Page.PageNumber, _ = strconv.Atoi(r.FormValue("pageNumber"))
	}

	// Processing removal
	if r.Method == "POST" && r.FormValue("deleteButton") != "" {
		elemsOnCurrentPage, _ := strconv.Atoi(r.FormValue("elemsOnCurrentPage"))
		r.ParseForm()
		ids := []int{}
		for _, v := range r.Form["ids"] {
			id, _ := strconv.Atoi(v)
			ids = append(ids, id)
		}
		if len(ids) > 0 {
			if Page.RemoveAllowed && sqla.VerifyRemovalPermissions(pb.db, pb.dbType, "projects", Page.LoggedinID, Page.LoggedinAdmin, pb.removeAllowed, ids) {
				removed := sqla.DeleteObjects(pb.db, pb.dbType, "projects", "ID", ids)
				if removed > 0 {
					Page.Message = "removedElems"
					Page.RemovedNum = removed
					if removed >= elemsOnCurrentPage && Page.PageNumber > 1 {
						Page.PageNumber--
					}
				} else {
					Page.Message = "removalError"
					log.Println("Error removing projects:", ids)
				}
			} else {
				Page.Message = "notAllorSomeElemsAllowedtoRemove"
				log.Println("Not allowed to remove attempt, LoggedinID and ids:", Page.LoggedinID, ids)
			}
		}
	}

	offset := (Page.PageNumber - 1) * Page.UserConfig.ElemsOnPage
	if offset < 0 {
		offset = 0
	}

	sq, sqcount, args, argscount := sqla.ConstructSELECTquery(pb.dbType, "projects",
		"projects.ID, ProjName, Description, Creator, ProjStatus, creator.ID, creator.FirstName, creator.Surname, creator.JobTitle",
		"*", "LEFT JOIN profiles creator ON creator.ID = Creator",
		Page.Filters, "ProjName", 1, Page.UserConfig.ElemsOnPage, offset, false,
		sqla.Seek{UseSeek: false})

	// Loading objects
	err = func() error {
		rows, err := pb.db.Query(sq, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		var ProjName sql.NullString
		var Description sql.NullString
		var Creator sql.NullInt64
		var ProjStatus sql.NullInt64

		var CreatorID sql.NullInt64
		var CreatorFirstName sql.NullString
		var CreatorSurname sql.NullString
		var CreatorJobTitle sql.NullString

		for rows.Next() {
			var p Project
			err = rows.Scan(&p.ID, &ProjName, &Description, &Creator, &ProjStatus,
				&CreatorID, &CreatorFirstName, &CreatorSurname, &CreatorJobTitle)
			if err != nil {
				return err
			}
			p.ProjName = ProjName.String
			p.Description = Description.String
			p.ProjStatus = int(ProjStatus.Int64)
			if CreatorID.Valid == true {
				p.Creator = &team.Profile{
					ID:        int(CreatorID.Int64),
					FirstName: CreatorFirstName.String,
					Surname:   CreatorSurname.String,
					JobTitle:  CreatorJobTitle.String,
				}
			} else {
				p.Creator = nil
			}
			Page.Projects = append(Page.Projects, p)
		}

		var FilteredNum sql.NullInt64
		row := pb.db.QueryRow(sqcount, argscount...)
		err = row.Scan(&FilteredNum)
		if err != nil {
			return err
		}
		Page.FilteredNum = int(FilteredNum.Int64)

		return nil
	}()

	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Attention! This removes db column lists in outputs like JSON.
	// Usually columns should not be available to a client.
	Page.Filters.ClearColumnsValues()

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		return
	}

	// HTML output
	err = pb.templates.ExecuteTemplate(w, "projs.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}
