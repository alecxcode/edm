package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/alecxcode/sqla"
)

// TeamPage is passed into template
type TeamPage struct {
	AppTitle      string
	PageTitle     string
	LoggedinID    int
	LoggedinAdmin bool
	Message       string
	UserConfig    UserConfig
	Team          []Profile //payload
	SortedBy      string
	SortedHow     int
	Filters       sqla.Filter
	PageNumber    int
	FilteredNum   int
	RemovedNum    int
	UserList      []UserListElem
	UnitList      []UnitListElem
	CorpList      []CorpListElem
}

func (bs *BaseStruct) teamHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := bs.authVerify(w, r)
	if !allow {
		return
	}

	if bs.validURLs.Team.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	userFullName := "TRIM((COALESCE(profiles.Surname, '') || ' ' || COALESCE(profiles.FirstName, '') || ' ' || COALESCE(profiles.OtherName, '')))"
	if bs.dbt == sqla.MSSQL {
		userFullName = "TRIM((COALESCE(profiles.Surname, '') + ' ' + COALESCE(profiles.FirstName, '') + ' ' + COALESCE(profiles.OtherName, '')))"
	}

	var Page = TeamPage{
		AppTitle:  bs.lng.AppTitle,
		PageTitle: bs.lng.TeamPageTitle,
		SortedBy:  "FullName",
		SortedHow: 1, // 0 - DESC, 1 - ASC
		Filters: sqla.Filter{ClassFilter: []sqla.ClassFilter{
			{Name: "jobunits", Column: "units.ID"},
			{Name: "companies", Column: "companies.ID"},
			{Name: "userrole", Column: "profiles.UserRole"},
			{Name: "userlock", Column: "profiles.UserLock"}},
			TextFilterName:    "searchText",
			TextFilterColumns: []string{userFullName, "profiles.Contacts", "profiles.JobTitle", "units.UnitName", "companies.ShortName"},
		},
		PageNumber: 1,
		LoggedinID: id,
	}

	user := bs.team.getByID(Page.LoggedinID)
	Page.UserConfig = user.UserConfig
	if user.UserRole == 1 {
		Page.LoggedinAdmin = true
	}

	// Parsing Filters
	Page.Filters.GetFilterFromForm(r,
		convDateStrToInt64, convDateTimeStrToInt64,
		map[string]int{"my": Page.LoggedinID})

	// Parsing other fields
	if r.FormValue("sortedBy") != "" {
		Page.SortedBy = r.FormValue("sortedBy")
	}
	var SQLSortedBy string
	switch Page.SortedBy {
	case "JobTitle":
		SQLSortedBy = "profiles.JobTitle, units.UnitName"
	case "Unit":
		SQLSortedBy = "units.UnitName, profiles.JobTitle"
	default:
		SQLSortedBy = "FullName"
	}
	if r.FormValue("sortedHow") != "" {
		Page.SortedHow, _ = strconv.Atoi(r.FormValue("sortedHow"))
	}

	if r.FormValue("elemsOnPage") != "" {
		Page.UserConfig.ElemsOnPageTeam, _ = strconv.Atoi(r.FormValue("elemsOnPage"))
		if r.FormValue("elemsOnPageChanged") == "true" {
			p := Profile{ID: Page.LoggedinID, UserConfig: Page.UserConfig}
			updated := p.updateConfig(bs.db, bs.dbt)
			if updated > 0 {
				bs.team.updateConfig(p)
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
			allowedToRemove := false
			if Page.LoggedinAdmin {
				allowedToRemove = true
			}
			LastAdmins := false
			LastAdmins, err = areTheyLastAdmins(bs.db, bs.dbt, ids)
			if err != nil {
				log.Println(currentFunction()+":", err)
			}
			if allowedToRemove && !LastAdmins {
				removed := sqla.DeleteObjects(bs.db, bs.dbt, "profiles", "ID", ids)
				if removed > 0 {
					if bs.dbt == sqla.MSSQL {
						sqla.SetToNull(bs.db, bs.dbt, "profiles", "Boss", ids)
						sqla.SetToNull(bs.db, bs.dbt, "tasks", "Creator", ids)
						sqla.SetToNull(bs.db, bs.dbt, "tasks", "Assignee", ids)
					}
					Page.Message = "removedElems"
					Page.RemovedNum = removed
					for _, eachID := range ids {
						bs.team.delProfile(eachID)
					}
					bs.team.constructUserList(bs.db, bs.dbt)
					if removed >= elemsOnCurrentPage && Page.PageNumber > 1 {
						Page.PageNumber--
					}
				} else {
					Page.Message = "removalError"
					log.Println("Error removing profiles:", ids)
				}
			} else {
				if LastAdmins {
					Page.Message = "LastAdminRejection"
				} else {
					Page.Message = "notAllorSomeElemsAllowedtoRemove"
				}
				log.Println("Not allowed to or last admin remove attempt, LoggedinID and ids:", Page.LoggedinID, ids)
			}
		}
	}

	OFFSET := (Page.PageNumber - 1) * Page.UserConfig.ElemsOnPageTeam
	if OFFSET < 0 {
		OFFSET = 0
	}

	sq, sqcount, args, argscount := sqla.ConstructSELECTquery(
		bs.dbt,
		"profiles",
		`profiles.ID, profiles.FirstName, profiles.OtherName, profiles.Surname,`+userFullName+` as FullName,
profiles.JobTitle, profiles.JobUnit, profiles.Contacts,
units.ID, units.Company, units.UnitName,
companies.ID, companies.ShortName`,
		"*",
		`LEFT JOIN units ON units.ID = profiles.JobUnit
LEFT JOIN companies ON companies.ID = units.Company`,
		Page.Filters,
		SQLSortedBy,
		Page.SortedHow,
		Page.UserConfig.ElemsOnPageTeam,
		OFFSET,
		false,
		sqla.Seek{UseSeek: false})

	// Loading objects
	err = func() error {
		rows, err := bs.db.Query(sq, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		var FirstName sql.NullString
		var OtherName sql.NullString
		var Surname sql.NullString
		var FullName sql.NullString

		var JobTitle sql.NullString
		var JobUnit sql.NullInt64
		var Contacts sql.NullString

		var UnitID sql.NullInt64
		var UnitCompany sql.NullInt64
		var UnitName sql.NullString

		var CompanyID sql.NullInt64
		var ShortName sql.NullString

		for rows.Next() {
			var p Profile
			err = rows.Scan(&p.ID, &FirstName, &OtherName, &Surname, &FullName, &JobTitle, &JobUnit, &Contacts,
				&UnitID, &UnitCompany, &UnitName,
				&CompanyID, &ShortName)
			if err != nil {
				return err
			}
			p.FirstName = FirstName.String
			p.OtherName = OtherName.String
			p.Surname = Surname.String
			p.JobTitle = JobTitle.String
			if UnitID.Valid {
				p.JobUnit = &Unit{
					ID: int(UnitID.Int64),
					Company: &Company{
						ID:        int(CompanyID.Int64),
						ShortName: ShortName.String,
					},
					UnitName: UnitName.String,
				}
			}
			p.Contacts = unmarshalNonEmptyProfileContacts(Contacts.String)
			Page.Team = append(Page.Team, p)
		}
		var FilteredNum sql.NullInt64
		row := bs.db.QueryRow(sqcount, argscount...)
		err = row.Scan(&FilteredNum)
		if err != nil {
			return err
		}
		Page.FilteredNum = int(FilteredNum.Int64)
		return nil
	}()

	if err != nil {
		log.Println(currentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		//http.Error(w, err.Error(), http.StatusInternalServerError) //Commented to not displayng error details to end user
		return
	}

	Page.UserList = bs.team.returnUserList()
	Page.UnitList = bs.team.returnUnitList()
	Page.CorpList = bs.team.returnCorpList()

	// Attention! This removes db column lists in outputs like JSON.
	// Usually columns should not be available to a client.
	Page.Filters.ClearColumnsValues()

	// JSON output
	if r.URL.Query().Get("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		//jsonOut, _ := json.Marshal(Page)
		//fmt.Fprintln(w, string(jsonOut))
		return
	}

	// HTML output
	err = bs.templates.ExecuteTemplate(w, "team.tmpl", Page)
	if err != nil {
		log.Println(currentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		//http.Error(w, err.Error(), http.StatusInternalServerError) //Commented to not displayng error details to end user
		return
	}
}

func areTheyLastAdmins(db *sql.DB, DBType byte, ids []int) (res bool, err error) {
	const ADMIN = 1
	const LOCKED = 1
	counter, sq, args := sqla.BuildSQLINNOT(DBType, "SELECT COUNT(*) FROM profiles ", 0, "ID", ids)
	sq += " AND UserRole = " + sqla.MakeParam(DBType, counter+1) + " AND UserLock <> " + sqla.MakeParam(DBType, counter+2) + " AND Login IS NOT NULL"
	if DEBUG {
		log.Println(sq, args)
	}
	args = append(args, ADMIN, LOCKED)
	row := db.QueryRow(sq, args...)
	var counted sql.NullInt64
	err = row.Scan(&counted)
	if err != nil {
		log.Println(currentFunction()+":", err)
		return false, nil
	}
	AdminsRemains := int(counted.Int64)
	if AdminsRemains < 1 {
		return true, nil
	} else {
		return false, nil
	}
}
