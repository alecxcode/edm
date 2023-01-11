package team

import (
	"database/sql"
	"edm/internal/core"
	"edm/pkg/accs"
	"edm/pkg/datetime"
	"edm/pkg/memdb"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/alecxcode/sqla"
)

// TeamPage is passed into template
type TeamPage struct {
	AppTitle      string
	AppVersion    string
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
	UserList      []memdb.ObjHasID
	UnitList      []memdb.ObjHasID
	CorpList      []memdb.ObjHasID
}

// TeamHandler is http handler for docs page
func (tb *TeamBase) TeamHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, tb.memorydb)
	if !allow {
		return
	}

	if tb.validURLs.team.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	userFullName := "TRIM((COALESCE(profiles.Surname, '') || ' ' || COALESCE(profiles.FirstName, '') || ' ' || COALESCE(profiles.OtherName, '')))"
	if tb.dbType == sqla.MSSQL {
		userFullName = "TRIM((COALESCE(profiles.Surname, '') + ' ' + COALESCE(profiles.FirstName, '') + ' ' + COALESCE(profiles.OtherName, '')))"
	}

	var Page = TeamPage{
		AppTitle:   tb.text.AppTitle,
		AppVersion: core.AppVersion,
		PageTitle:  tb.text.TeamPageTitle,
		SortedBy:   "FullName",
		SortedHow:  1, // 0 - DESC, 1 - ASC
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

	user := UnmarshalToProfile(tb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == ADMIN {
		Page.LoggedinAdmin = true
	}

	// Parsing Filters
	Page.Filters.GetFilterFromForm(r,
		datetime.ConvDateStrToInt64, datetime.ConvDateTimeStrToInt64,
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
			updated := p.UpdateConfig(tb.db, tb.dbType)
			if updated > 0 {
				MemoryUpdateProfile(tb.db, tb.dbType, tb.memorydb, p.ID)
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
			LastAdmins, err = areTheyLastAdmins(tb.db, tb.dbType, ids)
			if err != nil {
				log.Println(accs.CurrentFunction()+":", err)
			}
			if allowedToRemove && !LastAdmins {
				removed := sqla.DeleteObjects(tb.db, tb.dbType, "profiles", "ID", ids)
				if removed > 0 {
					if tb.dbType == sqla.MSSQL {
						sqla.SetToNull(tb.db, tb.dbType, "profiles", "Boss", ids)
						sqla.SetToNull(tb.db, tb.dbType, "tasks", "Creator", ids)
						sqla.SetToNull(tb.db, tb.dbType, "tasks", "Assignee", ids)
					}
					Page.Message = "removedElems"
					Page.RemovedNum = removed
					for _, eachID := range ids {
						tb.memorydb.DelObject(eachID)
					}
					core.ConstructUserList(tb.db, tb.dbType, tb.memorydb)
					if removed >= elemsOnCurrentPage && Page.PageNumber > 1 {
						Page.PageNumber--
					}
				} else {
					Page.Message = "removalError"
					log.Println("Error removing profiles:", ids)
				}
			} else {
				if LastAdmins {
					Page.Message = "lastAdminRejection"
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
		tb.dbType,
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
		rows, err := tb.db.Query(sq, args...)
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
			p.Contacts = UnmarshalNonEmptyProfileContacts(Contacts.String)
			Page.Team = append(Page.Team, p)
		}
		var FilteredNum sql.NullInt64
		row := tb.db.QueryRow(sqcount, argscount...)
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

	Page.UserList = tb.memorydb.GetObjectArr("UserList")
	Page.UnitList = tb.memorydb.GetObjectArr("UnitList")
	Page.CorpList = tb.memorydb.GetObjectArr("CorpList")

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
	err = tb.templates.ExecuteTemplate(w, "team.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func areTheyLastAdmins(db *sql.DB, DBType byte, ids []int) (res bool, err error) {
	const ADMIN = 1
	const LOCKED = 1
	counter, sq, args := sqla.BuildSQLINNOT(DBType, "SELECT COUNT(*) FROM profiles ", 0, "ID", ids)
	sq += " AND UserRole = " + sqla.MakeParam(DBType, counter+1) + " AND UserLock <> " + sqla.MakeParam(DBType, counter+2) + " AND Login IS NOT NULL"
	if core.DEBUG {
		log.Println(sq, args)
	}
	args = append(args, ADMIN, LOCKED)
	row := db.QueryRow(sq, args...)
	var counted sql.NullInt64
	err = row.Scan(&counted)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		return false, nil
	}
	AdminsRemains := int(counted.Int64)
	if AdminsRemains < 1 {
		return true, nil
	} else {
		return false, nil
	}
}
