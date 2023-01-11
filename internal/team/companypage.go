package team

import (
	"edm/internal/core"
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alecxcode/sqla"
)

// CompanyPage is passed into template
type CompanyPage struct {
	AppTitle   string
	AppVersion string
	PageTitle  string
	LoggedinID int
	UserConfig UserConfig
	Company    Company //payload
	Units      []Unit  //payload+
	Message    string
	Editable   bool
	New        bool
	UserList   []memdb.ObjHasID
}

// CompanyHandler is http handler for company page
func (tb *TeamBase) CompanyHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, tb.memorydb)
	if !allow {
		return
	}

	if tb.validURLs.comp.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = CompanyPage{
		AppTitle:   tb.text.AppTitle,
		AppVersion: core.AppVersion,
		LoggedinID: id,
		Editable:   false,
		New:        false,
	}

	user := UnmarshalToProfile(tb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == ADMIN {
		Page.Editable = true
	}

	TextID := accs.GetTextIDfromURL(r.URL.Path)
	IntID, _ := strconv.Atoi(TextID)

	var created int
	var updated int

	// Create or update code ==========================================================================
	if r.Method == "POST" && (r.FormValue("createButton") != "" || r.FormValue("updateButton") != "") {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "writing company", Page.LoggedinID, IntID)
			return
		}
		c := Company{
			ID:          IntID,
			ShortName:   r.FormValue("shortName"),
			FullName:    r.FormValue("fullName"),
			ForeignName: r.FormValue("foreignName"),
			Contacts: CompanyContacts{
				AddressReg:  r.FormValue("addressReg"),
				AddressFact: r.FormValue("addressFact"),
				Phone:       r.FormValue("phone"),
				Email:       r.FormValue("email"),
				WebSite:     r.FormValue("website"),
				Other:       r.FormValue("otherContacts"),
			},
			RegNo:       r.FormValue("regNo"),
			TaxNo:       r.FormValue("taxNo"),
			BankDetails: r.FormValue("bankDetails"),
		}
		if r.FormValue("companyHead") != "" && r.FormValue("companyHead") != "0" {
			c.CompanyHead = &Profile{ID: accs.StrToInt(r.FormValue("companyHead"))}
		}

		if r.FormValue("createButton") != "" {
			c.ID, created = c.Create(tb.db, tb.dbType)
			if created > 0 {
				core.ConstructCorpList(tb.db, tb.dbType, tb.memorydb)
				if Page.UserConfig.ReturnAfterCreation {
					http.Redirect(w, r, "/companies/", http.StatusSeeOther)
				} else {
					http.Redirect(w, r, fmt.Sprintf("/companies/company/%d", c.ID), http.StatusSeeOther)
				}
				return
			} else {
				Page.Message = "dataNotWritten"
			}
		}

		if r.FormValue("updateButton") != "" {
			updated = c.update(tb.db, tb.dbType)
			if updated > 0 {
				core.ConstructCorpList(tb.db, tb.dbType, tb.memorydb)
				core.ConstructUnitList(tb.db, tb.dbType, tb.memorydb)
				Page.Message = "dataWritten"
			} else {
				Page.Message = "dataNotWritten"
			}
		}

	}

	// Create or update Units =====================================================================
	if r.Method == "POST" && (r.FormValue("createUnit") != "" || r.FormValue("updateUnit") != "") {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "writing unit", Page.LoggedinID, IntID)
			return
		}
		UnitIntID, _ := strconv.Atoi(r.FormValue("unitID"))
		u := Unit{
			ID:       UnitIntID,
			UnitName: r.FormValue("unitName"),
			Company:  &Company{ID: IntID},
		}
		if r.FormValue("unitHead") != "" && r.FormValue("unitHead") != "0" {
			u.UnitHead = &Profile{ID: accs.StrToInt(r.FormValue("unitHead"))}
		}
		var unitaff int
		if r.FormValue("createUnit") != "" {
			_, unitaff = u.Create(tb.db, tb.dbType)
		}
		if r.FormValue("updateUnit") != "" {
			unitaff = u.update(tb.db, tb.dbType)
		}
		if unitaff > 0 {
			core.ConstructUnitList(tb.db, tb.dbType, tb.memorydb)
			Page.Message = "dataWritten"
		} else {
			Page.Message = "dataNotWritten"
		}
	}

	// Delete Unit ================+==========================
	if r.Method == "POST" && r.FormValue("deleteUnit") != "" {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "writing unit", Page.LoggedinID, IntID)
			return
		}
		UnitIntID, _ := strconv.Atoi(r.FormValue("unitID"))
		unitaff := sqla.DeleteObjects(tb.db, tb.dbType, "units", "ID", []int{UnitIntID})
		if unitaff > 0 {
			core.ConstructUnitList(tb.db, tb.dbType, tb.memorydb)
			Page.Message = "dataWritten"
		} else {
			Page.Message = "dataNotWritten"
		}
	}

	// Loading code ============================================
	Page.UserList = tb.memorydb.GetObjectArr("UserList")
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = tb.text.NewCompany
		if Page.Message == "" {
			Page.Message = "onlyAdminCanCreate"
		}
	} else {
		Page.Company.ID = IntID
		err = Page.Company.load(tb.db, tb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			http.NotFound(w, r)
			return
		}
		Page.Units, err = Page.Company.loadUnits(tb.db, tb.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			http.NotFound(w, r)
			return
		}
		Page.PageTitle = Page.Company.ShortName
		if Page.PageTitle == "" {
			Page.PageTitle = Page.Company.FullName
		}
		if Page.PageTitle == "" {
			Page.PageTitle = Page.Company.ForeignName
		}
		if Page.PageTitle == "" {
			Page.PageTitle = tb.text.Company + " ID: " + strconv.Itoa(Page.Company.ID)
		}
	}

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		return
	}

	// HTML output
	err = tb.templates.ExecuteTemplate(w, "company.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		accs.ThrowServerError(w, "executing company template", Page.LoggedinID, Page.Company.ID)
		return
	}

}
