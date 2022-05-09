package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/alecxcode/sqla"
)

// CompaniesPage is passed into template
type CompaniesPage struct {
	AppTitle      string
	PageTitle     string
	LoggedinID    int
	LoggedinAdmin bool
	Message       string
	UserConfig    UserConfig
	Companies     []Company //payload
	FilteredNum   int
	RemovedNum    int
}

func (bs *BaseStruct) companiesHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := bs.authVerify(w, r)
	if !allow {
		return
	}

	if bs.validURLs.Comp.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = CompaniesPage{
		AppTitle:   bs.lng.AppTitle,
		PageTitle:  bs.lng.CompaniesPageTitle,
		LoggedinID: id,
	}

	user := bs.team.getByID(Page.LoggedinID)
	Page.UserConfig = user.UserConfig
	if user.UserRole == 1 {
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
				removed := sqla.DeleteObjects(bs.db, bs.dbt, "companies", "ID", ids)
				if removed > 0 {
					bs.team.constructCorpList(bs.db, bs.dbt)
					bs.team.constructUnitList(bs.db, bs.dbt)
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
		rows, err := bs.db.Query(`SELECT 
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
		log.Println(currentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		//http.Error(w, err.Error(), http.StatusInternalServerError) //Commented to not displayng error details to end user
		return
	}

	// JSON output
	if r.URL.Query().Get("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		//jsonOut, _ := json.Marshal(Page)
		//fmt.Fprintln(w, string(jsonOut))
		return
	}

	// HTML output
	err = bs.templates.ExecuteTemplate(w, "companies.tmpl", Page)
	if err != nil {
		log.Println(currentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		//http.Error(w, err.Error(), http.StatusInternalServerError) //Commented to not displayng error details to end user
		return
	}
}
