package team

import (
	"database/sql"
	"edm/internal/core"
	"edm/pkg/accs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alecxcode/sqla"
)

// CompaniesPage is passed into template
type CompaniesPage struct {
	AppTitle      string
	AppVersion    string
	PageTitle     string
	LoggedinID    int
	LoggedinAdmin bool
	Message       string
	UserConfig    UserConfig
	Companies     []Company //payload
	FilteredNum   int
	RemovedNum    int
}

// CompaniesHandler is http handler for companies page
func (tb *TeamBase) CompaniesHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, tb.memorydb)
	if !allow {
		return
	}

	if tb.validURLs.comp.FindStringSubmatch(r.URL.Path) == nil {
		accs.ThrowObjectNotFound(w, r)
		return
	}

	var err error

	var Page = CompaniesPage{
		AppTitle:   tb.text.AppTitle,
		AppVersion: core.AppVersion,
		PageTitle:  tb.text.CompaniesPageTitle,
		LoggedinID: id,
	}

	user := UnmarshalToProfile(tb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == ADMIN {
		Page.LoggedinAdmin = true
	}

	// Processing removal
	if r.Method == "POST" && r.FormValue("deleteButton") != "" {
		r.ParseForm()
		ids := []int{}
		for _, v := range r.Form["ids"] {
			id, _ := strconv.Atoi(v)
			ids = append(ids, id)
		}
		if len(ids) > 0 {
			allowedToRemove := false
			if Page.LoggedinAdmin {
				allowedToRemove = true
			}
			if allowedToRemove {
				removed := sqla.DeleteObjects(tb.db, tb.dbType, "companies", "ID", ids)
				if removed > 0 {
					core.ConstructCorpList(tb.db, tb.dbType, tb.memorydb)
					core.ConstructUnitList(tb.db, tb.dbType, tb.memorydb)
					Page.Message = "removedElems"
					Page.RemovedNum = removed
				} else {
					log.Println("Error removing companies:", ids)
					Page.Message = "removalError"
				}
			} else {
				Page.Message = "notAllorSomeElemsAllowedtoRemove"
				log.Println("Not allowed to remove attempt, LoggedinID and ids:", Page.LoggedinID, ids)
			}
		}
	}

	// Loading objects
	err = func() error {
		rows, err := tb.db.Query(`SELECT 
c.ID, c.ShortName, c.FullName, c.ForeignName, 
c.Contacts, c.CompanyHead, c.RegNo, c.TaxNo, c.BankDetails, 
p.ID, p.FirstName, p.Surname,  p.JobTitle 
FROM companies c 
LEFT JOIN profiles p ON p.ID = c.CompanyHead 
ORDER BY c.ShortName ASC, c.FullName ASC, c.ForeignName ASC`)

		if err != nil {
			return err
		}
		defer rows.Close()

		var ShortName sql.NullString
		var FullName sql.NullString
		var ForeignName sql.NullString
		var Contacts sql.NullString
		var CompanyHead sql.NullInt64
		var RegNo sql.NullString
		var TaxNo sql.NullString
		var BankDetails sql.NullString
		var CompanyHeadID sql.NullInt64
		var CompanyHeadFirstName sql.NullString
		var CompanyHeadSurname sql.NullString
		var CompanyHeadJobTitle sql.NullString

		for rows.Next() {
			var c Company
			err = rows.Scan(&c.ID, &ShortName, &FullName, &ForeignName, &Contacts, &CompanyHead, &RegNo, &TaxNo, &BankDetails,
				&CompanyHeadID, &CompanyHeadFirstName, &CompanyHeadSurname, &CompanyHeadJobTitle)
			if err != nil {
				return err
			}
			c.ShortName = ShortName.String
			c.FullName = FullName.String
			c.ForeignName = ForeignName.String
			c.Contacts = unmarshalNonEmptyCompanyContacts(Contacts.String)
			if CompanyHeadID.Valid {
				c.CompanyHead = &Profile{
					ID:        int(CompanyHeadID.Int64),
					FirstName: CompanyHeadFirstName.String,
					Surname:   CompanyHeadSurname.String,
					JobTitle:  CompanyHeadJobTitle.String,
				}
			}
			c.RegNo = RegNo.String
			c.TaxNo = TaxNo.String
			c.BankDetails = BankDetails.String
			Page.Companies = append(Page.Companies, c)
			Page.FilteredNum++
		}
		return nil
	}()

	if err != nil {
		accs.ThrowServerError(w, fmt.Sprintf(accs.CurrentFunction()+":", err), Page.LoggedinID, 0)
		return
	}

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)

		return
	}

	// HTML output
	err = tb.templates.ExecuteTemplate(w, "companies.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
