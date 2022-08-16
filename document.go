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

// Approval is an approval item to a document
type Approval struct {
	//sql generate
	ID           int
	Written      DateTime `sql-gen:"bigint"`
	Approver     *Profile `sql-gen:"FK_NULL"`
	ApproverSign string
	DocID        int `sql-gen:"IDX,FK_CASCADE,fktable(documents)"`
	Approved     int
	Note         string `sql-gen:"varchar(max)"`
}

func (a *Approval) create(db *sql.DB, DBType byte) (lastid int, rowsaff int) {
	const NOACTION = 0
	var args sqla.AnyTslice
	if a.Approver != nil {
		args = args.AppendInt("Approver", a.Approver.ID)
	}
	args = args.AppendInt("DocID", a.DocID)
	args = args.AppendInt("Approved", NOACTION)
	lastid, rowsaff = sqla.InsertObject(db, DBType, "approvals", args)
	return lastid, rowsaff
}

func (a *Approval) sign(db *sql.DB, DBType byte) (rowsaff int) {
	var args sqla.AnyTslice
	if a.Written.Day != 0 {
		args = args.AppendInt64("Written", dateTimeToInt64(a.Written))
	}
	args = args.AppendStringOrNil("ApproverSign", a.ApproverSign)
	args = args.AppendInt("DocID", a.DocID)
	args = args.AppendInt("Approved", a.Approved)
	args = args.AppendStringOrNil("Note", a.Note)
	rowsaff = sqla.UpdateObject(db, DBType, "approvals", args, a.ID)
	return rowsaff
}

// GiveApproverID executes in a template to deliver the approver ID of this object
func (a Approval) GiveApproverID() int {
	if a.Approver == nil {
		return 0
	} else {
		return a.Approver.ID
	}
}

func (d *Document) loadApprovals(db *sql.DB, DBType byte) (ApprovalList []Approval, err error) {

	rows, err := db.Query(`SELECT a.ID, a.Written, a.Approver, a.ApproverSign, a.Approved, a.Note,
p.ID, p.FirstName, p.Surname, p.JobTitle
FROM approvals a
LEFT JOIN profiles p ON p.ID = a.Approver
WHERE DocID = `+sqla.MakeParam(DBType, 1)+` ORDER BY a.Written ASC, a.ID ASC`, d.ID)
	if err != nil {
		return ApprovalList, err
	}
	defer rows.Close()

	var ID sql.NullInt64
	var Written sql.NullInt64
	var Approver sql.NullInt64
	var ApproverSign sql.NullString
	var Approved sql.NullInt64
	var Note sql.NullString

	var ApproverID sql.NullInt64
	var ApproverFirstName sql.NullString
	var ApproverSurname sql.NullString
	var ApproverJobTitle sql.NullString

	for rows.Next() {
		err = rows.Scan(&ID, &Written, &Approver, &ApproverSign, &Approved, &Note,
			&ApproverID, &ApproverFirstName, &ApproverSurname, &ApproverJobTitle)
		if err != nil {
			return ApprovalList, err
		}
		a := Approval{
			ID:           int(ID.Int64),
			Written:      int64ToDateTime(Written.Int64),
			ApproverSign: ApproverSign.String,
			DocID:        d.ID,
			Approved:     int(Approved.Int64),
			Note:         Note.String,
		}
		if Approver.Valid {
			a.Approver = &Profile{
				ID:        int(ApproverID.Int64),
				FirstName: ApproverFirstName.String,
				Surname:   ApproverSurname.String,
				JobTitle:  ApproverJobTitle.String,
			}
		}
		ApprovalList = append(ApprovalList, a)
	}

	return ApprovalList, nil
}

// GiveDateTime executes in a template to deliver the queried date and time of an approval
func (a Approval) GiveDateTime(dateFmt string, timeFmt string, sep string) string {
	var dt = a.Written
	var rt string
	if timeFmt == "12h am/pm" {
		rt = timeToString12(dt.Hour, dt.Minute)
	} else if timeFmt == "24h" {
		rt = timeToString24(dt.Hour, dt.Minute)
	} else {
		rt = timeToString24(dt.Hour, dt.Minute)
	}
	if dt.Day == 0 {
		return ""
	}
	return dateToString(Date{dt.Year, dt.Month, dt.Day}, dateFmt) + sep + rt
}

type approvals []Approval

func (as approvals) getApprovalsIDsSlice() []int {
	appids := make([]int, len(as))
	for i := 0; i < len(as); i++ {
		appids[i] = as[i].ID
	}
	return appids
}

func (as approvals) getApprovalsIDsSliceApproved() []int {
	const APPROVED = 1
	appids := []int{}
	for i := 0; i < len(as); i++ {
		if as[i].Approved == APPROVED {
			appids = append(appids, as[i].ID)
		}
	}
	return appids
}

func (as approvals) getApproversIDsSlice() []int {
	cids := make([]int, len(as))
	for i := 0; i < len(as); i++ {
		cids[i] = as[i].GiveApproverID()
	}
	return cids
}

func (as approvals) getApprovalIDbyDocIDandApproverID(docID int, pID int) int {
	for _, a := range as {
		if a.Approver != nil && a.DocID == docID && a.Approver.ID == pID {
			return a.ID
		}
	}
	return 0
}

func (as approvals) getApprovalByID(aID int) Approval {
	for _, a := range as {
		if a.ID == aID {
			return a
		}
	}
	return Approval{ID: 0}
}

func (as approvals) approved(docID int, pID int) int {
	const NOACTION = 0
	for _, a := range as {
		if a.Approver != nil && a.DocID == docID && a.Approver.ID == pID {
			return a.Approved
		}
	}
	return NOACTION
}

func (as approvals) GetApprovalNote(docID int, pID int) string {
	for _, a := range as {
		if a.Approver != nil && a.DocID == docID && a.Approver.ID == pID {
			return a.Note
		}
	}
	return ""
}

// DocumentPage is passed into template
type DocumentPage struct {
	AppTitle      string
	PageTitle     string
	LoggedinID    int
	UserConfig    UserConfig
	Document      Document  //payload
	Approvals     approvals //payload
	Message       string
	RemoveAllowed bool
	Editable      bool
	IamApprover   bool
	YouApproved   int
	New           bool
	Categories    []string
	DocTypes      []string
	Currencies    map[int]string
	ApprovalSign  []string
	UserList      []UserListElem
}

func (bs *BaseStruct) documentHandler(w http.ResponseWriter, r *http.Request) {

	const NOACTION = 0
	const APPROVED = 1
	const REJECTED = 2
	const BROKEN = 3

	const ADMIN = 1

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
		AppTitle:     bs.text.AppTitle,
		LoggedinID:   id,
		Editable:     false,
		New:          false,
		Categories:   bs.text.Categories,
		DocTypes:     bs.text.DocTypes,
		Currencies:   bs.currencies,
		ApprovalSign: bs.text.ApprovalSign,
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
		Page.Approvals, err = Page.Document.loadApprovals(bs.db, bs.dbt)
		if err != nil {
			log.Println(currentFunction()+":", err)
			throwServerError(w, "loading document approvals", Page.LoggedinID, Page.Document.ID)
			return
		}
		Page.IamApprover = intToBool(Page.Approvals.getApprovalIDbyDocIDandApproverID(IntID, Page.LoggedinID))
		Page.YouApproved = Page.Approvals.approved(IntID, Page.LoggedinID)
	}

	user := bs.team.getByID(Page.LoggedinID)
	Page.UserConfig = user.UserConfig
	if user.UserRole == ADMIN || Page.New || Page.Document.GiveCreatorID() == Page.LoggedinID {
		Page.Editable = true
	}
	Page.RemoveAllowed, _ = strconv.ParseBool(bs.cfg.RemoveAllowed)
	if user.UserRole == ADMIN {
		Page.RemoveAllowed = true
	}

	var created int
	var updated int
	shallBreakApproval := false
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
				if Page.UserConfig.ReturnAfterCreation {
					http.Redirect(w, r, "/docs/", http.StatusSeeOther)
				} else {
					http.Redirect(w, r, fmt.Sprintf("/docs/document/%d", d.ID), http.StatusSeeOther)
				}
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

	// Delete files ==================================
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
				shallBreakApproval = true
				Page.Document.load(bs.db, bs.dbt)
			} else {
				Page.Message = "dataNotWritten"
			}
		} else {
			Page.Message = "removalError"
		}
	}

	// Add approval ===========================================
	if r.Method == "POST" && r.FormValue("approvalAdd") != "" {
		if Page.Editable || Page.IamApprover {
			pID := strToInt(r.FormValue("approvalAdd"))
			a := Approval{
				Approver: &Profile{ID: pID},
				DocID:    IntID,
			}
			if sliceContainsInt(Page.Approvals.getApproversIDsSlice(), pID) {
				Page.Message = "approverAlreadyInList"
			} else {
				id, res := a.create(bs.db, bs.dbt)
				if res > 0 {
					a.ID = id
					Page.Message = "dataWritten"
					p := Profile{ID: pID}
					p.preload(bs.db, bs.dbt)
					title := Page.Document.makeTitle(bs.i18n.Categories, bs.i18n.DocTypes, bs.i18n.DocCaption)
					mail := AnyMail{bs.i18n.Messages.Subj.NewApproval + ": " + title,
						bs.i18n.Messages.Cont.ApproveThis + ": " + title + ". " + bs.i18n.Messages.Cont.InfoLink + ": ",
						bs.systemURL + "/docs/document/" + strconv.Itoa(Page.Document.ID), bs.i18n.Messages.DoNotReply, bs.systemURL, bs.i18n.Messages.MailerName}
					mail.constructToChannel(bs.db, bs.dbt, bs.anymailtmpl, bs.mailchan, p)
				} else {
					Page.Message = "dataNotWritten"
				}
			}
		} else {
			throwAccessDenied(w, "adding approval", Page.LoggedinID, IntID)
			return
		}
	}

	// Approve (sign) ===========================================
	if r.Method == "POST" && r.FormValue("approvalSign") != "" {
		if Page.Editable || Page.IamApprover {
			a := Approval{
				ID:       Page.Approvals.getApprovalIDbyDocIDandApproverID(IntID, Page.LoggedinID),
				Written:  getCurrentDateTime(),
				Approver: &Profile{ID: Page.LoggedinID},
				DocID:    IntID,
				Approved: strToInt(r.FormValue("approvalSign")),
				Note:     r.FormValue("approvalNote"),
			}
			if Page.YouApproved == APPROVED {
				Page.Message = "dataNotWritten"
			} else {
				if a.Approved == APPROVED {
					a.Approver.load(bs.db, bs.dbt)
					a.Approver.Login = "no access"
					a.Approver.Passwd = "no access"
					a.ApproverSign = a.Approver.GiveSelfNameJob()
					if a.Approver.GiveUnitID() != 0 {
						a.ApproverSign += "; " + a.Approver.GiveUnitName()
					}
				}
				res := a.sign(bs.db, bs.dbt)
				if res > 0 {
					if Page.Document.Creator != nil && (a.Approved == APPROVED || a.Approved == REJECTED) {
						Page.Message = "dataWritten"
						p := Profile{ID: Page.Document.Creator.ID}
						p.preload(bs.db, bs.dbt)
						ResultSubj := map[int]string{APPROVED: bs.i18n.Messages.Subj.Approved, REJECTED: bs.i18n.Messages.Subj.Rejected}
						ResultCont := map[int]string{APPROVED: bs.i18n.Messages.Cont.Approved, REJECTED: bs.i18n.Messages.Cont.Rejected}
						title := Page.Document.makeTitle(bs.i18n.Categories, bs.i18n.DocTypes, bs.i18n.DocCaption)
						mail := AnyMail{ResultSubj[a.Approved] + ": " + title,
							bs.i18n.DocCaption + ": " + title + " " + ResultCont[a.Approved] + " " + bs.i18n.Messages.Cont.ByApprover + ": " + a.Approver.GiveSelfNameJob() + ". " + bs.i18n.Messages.Cont.InfoLink + ": ",
							bs.systemURL + "/docs/document/" + strconv.Itoa(Page.Document.ID), bs.i18n.Messages.DoNotReply, bs.systemURL, bs.i18n.Messages.MailerName}
						mail.constructToChannel(bs.db, bs.dbt, bs.anymailtmpl, bs.mailchan, p)
					}
				} else {
					Page.Message = "dataNotWritten"
				}
			}
		} else {
			throwAccessDenied(w, "signing approval", Page.LoggedinID, IntID)
			return
		}
	}

	// Remove approval ===========================================
	if r.Method == "POST" && r.FormValue("approvalRemove") != "" {
		aID := strToInt(r.FormValue("approvalRemove"))
		if Page.Editable && Page.Approvals.getApprovalByID(aID).Approved == NOACTION {
			res := sqla.DeleteObject(bs.db, bs.dbt, "approvals", "ID", aID)
			if res > 0 {
				Page.Message = "dataWritten"
			} else {
				Page.Message = "dataNotWritten"
			}
		} else {
			throwAccessDenied(w, "removing approval", Page.LoggedinID, IntID)
			return
		}
	}

	// Reset approvals ===========================================
	if shallBreakApproval {
		sqla.UpdateMultipleWithOneInt(bs.db, bs.dbt, "approvals", "Approved", BROKEN, "Written", dateTimeToInt64(getCurrentDateTime()), Page.Approvals.getApprovalsIDsSliceApproved())
	}

	// Other fields code ============================================
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = bs.text.NewDocument
	} else {
		Page.PageTitle = Page.Document.makeTitle(Page.Categories, Page.DocTypes, bs.text.Document)
		Page.Approvals, err = Page.Document.loadApprovals(bs.db, bs.dbt)
		if err != nil {
			log.Println(currentFunction()+":", err)
			throwServerError(w, "loading document approvals", Page.LoggedinID, Page.Document.ID)
			return
		}
		Page.IamApprover = intToBool(Page.Approvals.getApprovalIDbyDocIDandApproverID(IntID, Page.LoggedinID))
		Page.YouApproved = Page.Approvals.approved(IntID, Page.LoggedinID)
	}
	Page.UserList = bs.team.returnUserList()

	template := "document.tmpl"
	if strings.HasSuffix(r.URL.Path, "approval") || strings.HasSuffix(r.URL.Path, "approval/") {
		Page.PageTitle = "Approval list"
		template = "approval.tmpl"
		Page.Document.Creator.load(bs.db, bs.dbt)
		Page.Document.Creator.Login = "no access"
		Page.Document.Creator.Passwd = "no access"
		Page.Document.Creator.BirthDate.Year = 0
		Page.Document.Creator.UserConfig = UserConfig{}
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
	err = bs.templates.ExecuteTemplate(w, template, Page)
	if err != nil {
		log.Println(currentFunction()+":", err)
		throwServerError(w, "executing "+template[0:len(template)-5]+" template", Page.LoggedinID, Page.Document.ID)
		return
	}

}
