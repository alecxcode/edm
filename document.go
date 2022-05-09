package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecxcode/sqla"
)

// Document is related to any document object
type Document struct {
	//sql generate
	ID        int
	RegNo     string `sql-gen:"varchar(255)"`
	RegDate   Date   `sql-gen:"bigint,IDX"`
	IncNo     string `sql-gen:"varchar(255)"`
	IncDate   Date   `sql-gen:"bigint,IDX"`
	Category  int    // See lang.go
	DocType   int    // See lang.go
	About     string `sql-gen:"varchar(4000)"`
	Authors   string `sql-gen:"varchar(2000)"`
	Addressee string `sql-gen:"varchar(2000)"`
	DocSum    int    `sql-gen:"bigint"`
	Currency  int
	EndDate   Date     `sql-gen:"bigint"`
	Creator   *Profile `sql-gen:"FK_NULL"`
	Note      string   `sql-gen:"varchar(4000)"`
	FileList  []string `sql-gen:"varchar(max)"`
}

func (d Document) print() {
	fmt.Printf("%#v\n", d)
}

// GiveCategory executes in a template to deliver the category of a document
func (d Document) GiveCategory(catslice []string, unknown string) string {
	if d.Category < len(catslice) && d.Category >= Undefined {
		return catslice[d.Category]
	} else {
		return unknown
	}
}

// GiveType executes in a template to deliver the type of a document
func (d Document) GiveType(typslice []string, unknown string) string {
	if d.DocType < len(typslice) && d.DocType >= Undefined {
		return typslice[d.DocType]
	} else {
		return unknown
	}
}

// GiveCreatorID executes in a template to deliver the creator ID of this object
func (d Document) GiveCreatorID() int {
	if d.Creator == nil {
		return 0
	} else {
		return d.Creator.ID
	}
}

// GiveDate executes in a template to deliver the queried date of a document
func (d Document) GiveDate(dateWhat string, dateFmt string) string {
	var datetoconv Date
	switch dateWhat {
	case "Reg":
		datetoconv = d.RegDate
	case "Inc":
		datetoconv = d.IncDate
	case "End":
		datetoconv = d.EndDate
	default:
		return "wrong arg"
	}
	return dateToString(datetoconv, dateFmt)
}

// GiveShortFileName executes in a template to deliver the shortened filelist
func (d Document) GiveShortFileName(index int) string {
	if index >= len(d.FileList) {
		return ""
	} else {
		return string([]rune(d.FileList[index])[0]) + ".." + filepath.Ext(d.FileList[index])
	}
}

// GiveSum executes in a template to deliver the sum of a document
func (d Document) GiveSum() string {
	if d.DocSum == 0 && d.Currency == Undefined {
		return ""
	} else {
		return toDecimalStr(strconv.Itoa(d.DocSum))
	}
}

func (d *Document) create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendNonEmptyString("RegNo", d.RegNo)
	if d.RegDate.Day != 0 {
		args = args.AppendInt64("RegDate", dateToInt64(d.RegDate))
	}
	args = args.AppendNonEmptyString("IncNo", d.IncNo)
	if d.IncDate.Day != 0 {
		args = args.AppendInt64("IncDate", dateToInt64(d.IncDate))
	}
	args = args.AppendInt("Category", d.Category) // Mandatory
	args = args.AppendInt("DocType", d.DocType)   // Mandatory
	args = args.AppendNonEmptyString("About", d.About)
	args = args.AppendNonEmptyString("Authors", d.Authors)
	args = args.AppendNonEmptyString("Addressee", d.Addressee)
	if d.DocSum != 0 || d.Currency != Undefined {
		args = args.AppendInt("DocSum", d.DocSum)
	}
	args = args.AppendInt("Currency", d.Currency)
	if d.EndDate.Day != 0 {
		args = args.AppendInt64("EndDate", dateToInt64(d.EndDate))
	}
	if d.Creator != nil {
		args = args.AppendInt("Creator", d.Creator.ID)
	}
	args = args.AppendNonEmptyString("Note", d.Note)
	args = args.AppendJSONList("FileList", d.FileList)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "documents", args)
	return lastid, rowsaff
}

func (d *Document) update(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendStringOrNil("RegNo", d.RegNo)
	if d.RegDate.Day != 0 {
		args = args.AppendInt64("RegDate", dateToInt64(d.RegDate))
	} else {
		args = args.AppendNil("RegDate")
	}
	args = args.AppendStringOrNil("IncNo", d.IncNo)
	if d.IncDate.Day != 0 {
		args = args.AppendInt64("IncDate", dateToInt64(d.IncDate))
	} else {
		args = args.AppendNil("IncDate")
	}
	args = args.AppendInt("Category", d.Category) // Mandatory
	args = args.AppendInt("DocType", d.DocType)   // Mandatory
	args = args.AppendStringOrNil("About", d.About)
	args = args.AppendStringOrNil("Authors", d.Authors)
	args = args.AppendStringOrNil("Addressee", d.Addressee)
	if d.DocSum != 0 || d.Currency != Undefined {
		args = args.AppendInt("DocSum", d.DocSum)
	} else {
		args = args.AppendNil("DocSum")
	}
	args = args.AppendInt("Currency", d.Currency)
	if d.EndDate.Day != 0 {
		args = args.AppendInt64("EndDate", dateToInt64(d.EndDate))
	} else {
		args = args.AppendNil("EndDate")
	}
	args = args.AppendStringOrNil("Note", d.Note)
	args = args.AppendJSONList("FileList", d.FileList)
	rowsaff = sqla.UpdateObject(db, DBType, "documents", args, d.ID)
	return rowsaff
}

func (d *Document) load(db *sql.DB, DBType byte) error {

	row := db.QueryRow(`SELECT
documents.ID, RegNo, RegDate, IncNo, IncDate, Category, DocType, About, Authors, Addressee, DocSum, Currency, EndDate, Creator, Note, FileList,
creator.ID, creator.FirstName, creator.Surname, creator.JobTitle
FROM documents
LEFT JOIN profiles creator ON creator.ID = Creator
WHERE documents.ID = `+sqla.MakeParam(DBType, 1), d.ID)

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

	var CreatorID sql.NullInt64
	var CreatorFirstName sql.NullString
	var CreatorSurname sql.NullString
	var CreatorJobTitle sql.NullString

	err := row.Scan(&d.ID, &RegNo, &RegDate, &IncNo, &IncDate, &Category, &DocType, &About, &Authors, &Addressee, &DocSum, &Currency, &EndDate, &Creator, &Note, &FileList,
		&CreatorID, &CreatorFirstName, &CreatorSurname, &CreatorJobTitle)
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
	if CreatorID.Valid == true {
		d.Creator = &Profile{
			ID:        int(CreatorID.Int64),
			FirstName: CreatorFirstName.String,
			Surname:   CreatorSurname.String,
			JobTitle:  CreatorJobTitle.String,
		}
	}
	d.Note = Note.String
	d.FileList = sqla.UnmarshalNonEmptyJSONList(FileList.String)

	return nil
}

// DocumentPage is passed into template
type DocumentPage struct {
	AppTitle      string
	PageTitle     string
	LoggedinID    int
	UserConfig    UserConfig
	Document      Document //payload
	Message       string
	RemoveAllowed bool
	Editable      bool
	New           bool
	Categories    []string
	DocTypes      []string
	Currencies    map[int]string
}

func (bs *BaseStruct) documentHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := bs.authVerify(w, r)
	if !allow {
		return
	}

	if bs.validURLs.Docs.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = DocumentPage{
		AppTitle:   bs.lng.AppTitle,
		LoggedinID: id,
		Editable:   false,
		New:        false,
		Categories: bs.lng.Categories,
		DocTypes:   bs.lng.DocTypes,
		Currencies: bs.currencies,
	}

	TextID := getTextIDfromURL(r.URL.Path)
	IntID, _ := strconv.Atoi(TextID)
	if TextID == "new" {
		Page.New = true
	} else {
		Page.Document = Document{ID: IntID}
		err = Page.Document.load(bs.db, bs.dbt)
		if err != nil {
			log.Println(currentFunction()+":", err)
			http.NotFound(w, r)
			return
		}
	}

	user := bs.team.getByID(Page.LoggedinID)
	Page.UserConfig = user.UserConfig
	if user.UserRole == 1 || Page.New || Page.Document.GiveCreatorID() == Page.LoggedinID {
		Page.Editable = true
	}
	Page.RemoveAllowed, _ = strconv.ParseBool(bs.cfg.RemoveAllowed)
	if user.UserRole == 1 {
		Page.RemoveAllowed = true
	}

	var created int
	var updated int
	defaultUploadPath := filepath.Join(bs.cfg.ServerRoot, "files", "docs", TextID)

	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE+1048576)
	}

	// Create or update code ==========================================================================
	if r.Method == "POST" && (r.FormValue("createButton") != "" || r.FormValue("updateButton") != "") {
		if Page.Editable == false {
			throwAccessDenied(w, "writing document", Page.LoggedinID, IntID)
			return
		}
		d := Document{
			ID:        IntID,
			RegNo:     r.FormValue("regNo"),
			RegDate:   stringToDate(r.FormValue("regDate")),
			IncNo:     r.FormValue("incNo"),
			IncDate:   stringToDate(r.FormValue("incDate")),
			Category:  strToInt(r.FormValue("category")),
			DocType:   strToInt(r.FormValue("docType")),
			About:     r.FormValue("about"),
			Authors:   r.FormValue("authors"),
			Addressee: r.FormValue("addressee"),
			DocSum:    processFormSumInt(r.FormValue("docSum")),
			Currency:  strToInt(r.FormValue("currencyCode")),
			EndDate:   stringToDate(r.FormValue("endDate")),
			Creator:   &Profile{ID: Page.LoggedinID},
			Note:      r.FormValue("note"),
		}
		if r.FormValue("docSum") != "" && d.Currency == 0 {
			d.Currency = -1
		} else if r.FormValue("docSum") == "" && d.Currency == -1 {
			d.Currency = 0
		}

		d.FileList, err = uploader(r, defaultUploadPath, "fileList")
		if err != nil {
			log.Println(currentFunction()+":", err)
			Page.Message = "uploadError"
		}

		if r.FormValue("createButton") != "" && !strings.Contains(Page.Message, "Error") {
			d.ID, created = d.create(bs.db, bs.dbt)
			if created > 0 {
				moveUploadedFilesToFinalDest(defaultUploadPath,
					filepath.Join(bs.cfg.ServerRoot, "files", "docs", strconv.Itoa(d.ID)),
					d.FileList)
				http.Redirect(w, r, fmt.Sprintf("/docs/document/%d", d.ID), http.StatusSeeOther)
				return
			} else {
				Page.Message = "dataNotWritten"
			}
		}

		if r.FormValue("updateButton") != "" && !strings.Contains(Page.Message, "Error") {
			d.FileList = append(Page.Document.FileList, d.FileList...)
			updated = d.update(bs.db, bs.dbt)
			if updated > 0 {
				Page.Message = "dataWritten"
				Page.Document.load(bs.db, bs.dbt)
			} else {
				Page.Message = "dataNotWritten"
			}
		}

	}

	// Create or update code ==================================
	if r.Method == "POST" && r.FormValue("deleteFiles") != "" {
		if Page.Editable == false {
			throwAccessDenied(w, "writing document", Page.LoggedinID, IntID)
			return
		}
		r.ParseForm()
		filesToRemove := r.Form["filesToRemove"]
		err = removeUploadedFiles(defaultUploadPath, filesToRemove)
		if err == nil {
			FileList := filterSliceStr(Page.Document.FileList, filesToRemove)
			updated = sqla.UpdateSingleJSONListStr(bs.db, bs.dbt, "documents", "FileList", FileList, IntID)
			if updated > 0 {
				Page.Message = "dataWritten"
				Page.Document.load(bs.db, bs.dbt)
			} else {
				Page.Message = "dataNotWritten"
			}
		} else {
			Page.Message = "removalError"
		}
	}

	// Other fields code ============================================
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = bs.lng.NewDocument
	} else {
		Page.PageTitle = Page.Document.GiveType(Page.DocTypes, bs.lng.Document)
		if Page.Document.DocType == 0 {
			Page.PageTitle = bs.lng.Document + " " + Page.PageTitle
		}
		if Page.Document.Category != 0 {
			Page.PageTitle += " (" + Page.Document.GiveCategory(Page.Categories, "") + ")"
		}
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
	err = bs.templates.ExecuteTemplate(w, "document.tmpl", Page)
	if err != nil {
		log.Println(currentFunction()+":", err)
		throwServerError(w, "executing document template", Page.LoggedinID, Page.Document.ID)
		return
	}

}
