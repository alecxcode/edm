package docs

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/currencies"
	"edm/pkg/datetime"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/alecxcode/sqla"
)

// Document is related to any document object
type Document struct {
	//sql generate
	ID        int
	RegNo     string        `sql-gen:"varchar(255)"`
	RegDate   datetime.Date `sql-gen:"bigint,IDX"`
	IncNo     string        `sql-gen:"varchar(255)"`
	IncDate   datetime.Date `sql-gen:"bigint,IDX"`
	Category  int           // See lang.go
	DocType   int           // See lang.go
	About     string        `sql-gen:"varchar(4000)"`
	Authors   string        `sql-gen:"varchar(2000)"`
	Addressee string        `sql-gen:"varchar(2000)"`
	DocSum    int           `sql-gen:"bigint"`
	Currency  int
	EndDate   datetime.Date `sql-gen:"bigint"`
	Creator   *team.Profile `sql-gen:"FK_NULL"`
	Note      string        `sql-gen:"varchar(4000)"`
	FileList  []string      `sql-gen:"varchar(max)"`
}

func (d Document) print() {
	fmt.Printf("%#v\n", d)
}

// GiveCategory executes in a template to deliver the category of a document
func (d Document) GiveCategory(catslice []string, unknown string) string {
	if d.Category < len(catslice) && d.Category >= core.Undefined {
		return catslice[d.Category]
	} else {
		return unknown
	}
}

// GiveType executes in a template to deliver the type of a document
func (d Document) GiveType(typslice []string, unknown string) string {
	if d.DocType < len(typslice) && d.DocType >= core.Undefined {
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
	var datetoconv datetime.Date
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
	return datetime.DateToString(datetoconv, dateFmt)
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
	if d.DocSum == 0 && d.Currency == core.Undefined {
		return ""
	} else {
		return currencies.ToDecimalStr(strconv.Itoa(d.DocSum))
	}
}

func (d Document) makeTitle(categories []string, docTypes []string, document string) (fullTitle string) {
	fullTitle = d.GiveType(docTypes, "Unknown")
	if d.DocType == 0 {
		fullTitle = document + " " + fullTitle
	}
	if d.Category != 0 {
		fullTitle += " (" + d.GiveCategory(categories, "") + ")"
	}
	if d.RegNo != "" {
		fullTitle += " No. " + d.RegNo
	}
	return fullTitle
}

// Create creates a document in DB
func (d *Document) Create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	var args sqla.AnyTslice
	args = args.AppendNonEmptyString("RegNo", d.RegNo)
	if d.RegDate.Day != 0 {
		args = args.AppendInt64("RegDate", datetime.DateToInt64(d.RegDate))
	}
	args = args.AppendNonEmptyString("IncNo", d.IncNo)
	if d.IncDate.Day != 0 {
		args = args.AppendInt64("IncDate", datetime.DateToInt64(d.IncDate))
	}
	args = args.AppendInt("Category", d.Category) // Mandatory
	args = args.AppendInt("DocType", d.DocType)   // Mandatory
	args = args.AppendNonEmptyString("About", d.About)
	args = args.AppendNonEmptyString("Authors", d.Authors)
	args = args.AppendNonEmptyString("Addressee", d.Addressee)
	if d.DocSum != 0 || d.Currency != core.Undefined {
		args = args.AppendInt("DocSum", d.DocSum)
	}
	args = args.AppendInt("Currency", d.Currency)
	if d.EndDate.Day != 0 {
		args = args.AppendInt64("EndDate", datetime.DateToInt64(d.EndDate))
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
		args = args.AppendInt64("RegDate", datetime.DateToInt64(d.RegDate))
	} else {
		args = args.AppendNil("RegDate")
	}
	args = args.AppendStringOrNil("IncNo", d.IncNo)
	if d.IncDate.Day != 0 {
		args = args.AppendInt64("IncDate", datetime.DateToInt64(d.IncDate))
	} else {
		args = args.AppendNil("IncDate")
	}
	args = args.AppendInt("Category", d.Category) // Mandatory
	args = args.AppendInt("DocType", d.DocType)   // Mandatory
	args = args.AppendStringOrNil("About", d.About)
	args = args.AppendStringOrNil("Authors", d.Authors)
	args = args.AppendStringOrNil("Addressee", d.Addressee)
	if d.DocSum != 0 || d.Currency != core.Undefined {
		args = args.AppendInt("DocSum", d.DocSum)
	} else {
		args = args.AppendNil("DocSum")
	}
	args = args.AppendInt("Currency", d.Currency)
	if d.EndDate.Day != 0 {
		args = args.AppendInt64("EndDate", datetime.DateToInt64(d.EndDate))
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
	if CreatorID.Valid == true {
		d.Creator = &team.Profile{
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
