package main

import (
	"database/sql"
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
	PageTitle     string
	LoggedinID    int
	LoggedinAdmin bool
	Message       string
	RemoveAllowed bool
	UserConfig    UserConfig
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
}

func (bs *BaseStruct) docsHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := bs.authVerify(w, r)
	if !allow {
		return
	}

	if bs.validURLs.Docs.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = DocsPage{
		AppTitle:   bs.lng.AppTitle,
		PageTitle:  bs.lng.DocsPageTitle,
		Categories: bs.lng.Categories,
		DocTypes:   bs.lng.DocTypes,
		Currencies: bs.currencies,
		SortedBy:   "RegDate",
		SortedHow:  0, // 0 - DESC, 1 - ASC
		Filters: sqla.Filter{
			ClassFilter: []sqla.ClassFilter{
				{Name: "categories", Column: "Category"},
				{Name: "doctypes", Column: "DocType"},
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

	Page.RemoveAllowed, _ = strconv.ParseBool(bs.cfg.RemoveAllowed)
	user := bs.team.getByID(Page.LoggedinID)
	Page.UserConfig = user.UserConfig
	if user.UserRole == 1 {
		Page.LoggedinAdmin = true
		Page.RemoveAllowed = true
	}

	Page.Filters.GetFilterFromForm(r,
		convDateStrToInt64, convDateTimeStrToInt64,
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
			allowedToRemove := sqla.VerifyRemovalPermissions(bs.db, bs.dbt, "documents", Page.LoggedinID, Page.LoggedinAdmin, Page.RemoveAllowed, ids)
			if allowedToRemove {
				removed := sqla.DeleteObjects(bs.db, bs.dbt, "documents", "ID", ids)
				if removed > 0 {
					removeUploadedDirs(filepath.Join(bs.cfg.ServerRoot, "files", "docs"), ids)
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

	sq, sqcount, args, argscount := sqla.ConstructSELECTquery(
		bs.dbt,
		"documents",
		"ID, RegNo, RegDate, IncNo, IncDate, Category, DocType, About, Authors, Addressee, DocSum, Currency, EndDate, Creator, Note, FileList",
		"*",
		"",
		Page.Filters,
		SQLSortedBy,
		Page.SortedHow,
		Page.UserConfig.ElemsOnPage,
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
			d.RegDate = getValidDateFromSQL(RegDate)
			d.IncNo = IncNo.String
			d.IncDate = getValidDateFromSQL(IncDate)
			d.Category = int(Category.Int64)
			d.DocType = int(DocType.Int64)
			d.About = About.String
			d.Authors = Authors.String
			d.Addressee = Addressee.String
			d.DocSum = int(DocSum.Int64)
			d.Currency = int(Currency.Int64)
			d.EndDate = getValidDateFromSQL(EndDate)
			if Creator.Valid == true {
				d.Creator = &Profile{ID: int(Creator.Int64)}
			}
			d.Note = Note.String
			d.FileList = sqla.UnmarshalNonEmptyJSONList(FileList.String)
			Page.Docs = append(Page.Docs, d)
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
	err = bs.templates.ExecuteTemplate(w, "docs.tmpl", Page)
	if err != nil {
		log.Println(currentFunction()+":", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		//http.Error(w, err.Error(), http.StatusInternalServerError) //Commented to not displayng error details to end user
		return
	}
}
