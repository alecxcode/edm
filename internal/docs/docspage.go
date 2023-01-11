package docs

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"edm/pkg/currencies"
	"edm/pkg/datetime"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/alecxcode/sqla"
)

// DocsPage is passed into template
type DocsPage struct {
	AppTitle      string
	AppVersion    string
	PageTitle     string
	LoggedinID    int
	LoggedinAdmin bool
	Message       string
	RemoveAllowed bool
	UserConfig    team.UserConfig
	Docs          []Document //payload
	SortedBy      string
	SortedHow     int
	ShowIncNo     bool
	Filters       sqla.Filter
	PageNumber    int
	FilteredNum   int
	RemovedNum    int
	Categories    []string
	DocTypes      []string
	Currencies    map[int]string
	ApprovalSign  []string
}

// DocsHandler is http handler for docs page
func (dd *DocsBase) DocsHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, dd.memorydb)
	if !allow {
		return
	}

	if dd.validURLs.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = DocsPage{
		AppTitle:     dd.text.AppTitle,
		AppVersion:   core.AppVersion,
		PageTitle:    dd.text.DocsPageTitle,
		Categories:   dd.text.Categories,
		DocTypes:     dd.text.DocTypes,
		Currencies:   currencies.GetCurrencies(),
		ApprovalSign: dd.text.ApprovalSign,
		SortedBy:     "RegDate",
		SortedHow:    0, // 0 - DESC, 1 - ASC
		Filters: sqla.Filter{
			ClassFilter: []sqla.ClassFilter{
				{Name: "categories", Column: "Category"},
				{Name: "doctypes", Column: "DocType"},
				{Name: "approver", Column: "Approver"},
				{Name: "approved", Column: "Approved"},
				{Name: "creator", Column: "Creator"},
			},
			DateFilter: []sqla.DateFilter{
				{Name: "regDates", Column: "RegDate"},
				{Name: "incDates", Column: "IncDate"},
				{Name: "endDates", Column: "EndDate"},
			},
			SumFilter: []sqla.SumFilter{
				{Name: "sums", Column: "DocSum", CurrencyColumn: "Currency"},
			},
			TextFilterName:    "searchText",
			TextFilterColumns: []string{"RegNo", "IncNo", "About", "Authors", "Addressee", "Note", "FileList"},
		},
		PageNumber: 1,
		LoggedinID: id,
	}

	Page.RemoveAllowed = dd.cfg.removeAllowed
	user := team.UnmarshalToProfile(dd.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == team.ADMIN {
		Page.LoggedinAdmin = true
		Page.RemoveAllowed = true
	}

	Page.Filters.GetFilterFromForm(r,
		datetime.ConvDateStrToInt64, datetime.ConvDateTimeStrToInt64,
		map[string]int{"my": Page.LoggedinID})

	// Parsing other fields
	if r.FormValue("sortedBy") != "" {
		Page.SortedBy = r.FormValue("sortedBy")
	}
	var SQLSortedBy string
	switch Page.SortedBy {
	case "IncNo":
		SQLSortedBy = "IncNo"
	case "IncDate":
		SQLSortedBy = "IncDate"
	case "RegNo":
		SQLSortedBy = "RegNo"
	case "RegDate":
		SQLSortedBy = "RegDate"
	case "About":
		SQLSortedBy = "About"
	case "EndDate":
		SQLSortedBy = "EndDate"
	case "DocSum":
		SQLSortedBy = "DocSum"
	default:
		SQLSortedBy = "RegDate"
	}
	if r.FormValue("sortedHow") != "" {
		Page.SortedHow, _ = strconv.Atoi(r.FormValue("sortedHow"))
	}

	if r.FormValue("showIncNo") == "true" {
		Page.ShowIncNo = true
	}

	if r.FormValue("elemsOnPage") != "" {
		Page.UserConfig.ElemsOnPage, _ = strconv.Atoi(r.FormValue("elemsOnPage"))
		if r.FormValue("elemsOnPageChanged") == "true" {
			p := team.Profile{ID: Page.LoggedinID, UserConfig: Page.UserConfig}
			updated := p.UpdateConfig(dd.db, dd.dbType)
			if updated > 0 {
				team.MemoryUpdateProfile(dd.db, dd.dbType, dd.memorydb, p.ID)
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
			allowedToRemove := sqla.VerifyRemovalPermissions(dd.db, dd.dbType, "documents", Page.LoggedinID, Page.LoggedinAdmin, Page.RemoveAllowed, ids)
			if allowedToRemove {
				removed := sqla.DeleteObjects(dd.db, dd.dbType, "documents", "ID", ids)
				if removed > 0 {
					core.RemoveUploadedDirs(filepath.Join(dd.cfg.serverRoot, "files", "docs"), ids)
					Page.Message = "removedElems"
					Page.RemovedNum = removed
					if removed >= elemsOnCurrentPage && Page.PageNumber > 1 {
						Page.PageNumber--
					}
				} else {
					Page.Message = "removalError"
					log.Println("Error removing documens:", ids)
				}
			} else {
				Page.Message = "notAllorSomeElemsAllowedtoRemove"
				log.Println("Not allowed to remove attempt, LoggedinID and ids:", Page.LoggedinID, ids)
			}
		}
	}

	OFFSET := (Page.PageNumber - 1) * Page.UserConfig.ElemsOnPage
	if OFFSET < 0 {
		OFFSET = 0
	}

	columns := "documents.ID, RegNo, RegDate, IncNo, IncDate, Category, DocType, About, Authors, Addressee, DocSum, Currency, EndDate, Creator, documents.Note, FileList"
	joins := ""

	if r.URL.Query().Get("approver") != "" || r.URL.Query().Get("approved") != "" {
		joins = "LEFT JOIN approvals ON approvals.DocID = documents.ID"
	}

	distinct := false
	columnsToCount := "*"
	if r.URL.Query().Get("approver") == "" && r.URL.Query().Get("approved") != "" {
		distinct = true
		columnsToCount = "documents.ID"
	}

	sq, sqcount, args, argscount := sqla.ConstructSELECTquery(
		dd.dbType,
		"documents",
		columns,
		columnsToCount,
		joins,
		Page.Filters,
		SQLSortedBy,
		Page.SortedHow,
		Page.UserConfig.ElemsOnPage,
		OFFSET,
		distinct,
		sqla.Seek{UseSeek: false})

	// Loading objects
	err = func() error {
		rows, err := dd.db.Query(sq, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		var RegNo sql.NullString
		var RegDate sql.NullInt64
		var IncNo sql.NullString
		var IncDate sql.NullInt64
		var Category sql.NullInt64
		var DocType sql.NullInt64
		var About sql.NullString
		var Authors sql.NullString
		var Addressee sql.NullString
		var DocSum sql.NullInt64
		var Currency sql.NullInt64
		var EndDate sql.NullInt64
		var Creator sql.NullInt64
		var Note sql.NullString
		var FileList sql.NullString

		for rows.Next() {
			var d Document
			err = rows.Scan(&d.ID, &RegNo, &RegDate, &IncNo, &IncDate, &Category, &DocType, &About, &Authors, &Addressee, &DocSum, &Currency, &EndDate, &Creator, &Note, &FileList)
			if err != nil {
				return err
			}

			d.RegNo = RegNo.String
			d.RegDate = datetime.GetValidDateFromSQL(RegDate)
			d.IncNo = IncNo.String
			d.IncDate = datetime.GetValidDateFromSQL(IncDate)
			d.Category = int(Category.Int64)
			d.DocType = int(DocType.Int64)
			d.About = About.String
			d.Authors = Authors.String
			d.Addressee = Addressee.String
			d.DocSum = int(DocSum.Int64)
			d.Currency = int(Currency.Int64)
			d.EndDate = datetime.GetValidDateFromSQL(EndDate)
			if Creator.Valid == true {
				d.Creator = &team.Profile{ID: int(Creator.Int64)}
			}
			d.Note = Note.String
			d.FileList = sqla.UnmarshalNonEmptyJSONList(FileList.String)
			Page.Docs = append(Page.Docs, d)
		}

		var FilteredNum sql.NullInt64
		row := dd.db.QueryRow(sqcount, argscount...)
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
	err = dd.templates.ExecuteTemplate(w, "docs.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
