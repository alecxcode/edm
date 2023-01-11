package docs

import (
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"edm/pkg/currencies"
	"edm/pkg/datetime"
	"edm/pkg/memdb"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecxcode/sqla"
)

// DocumentPage is passed into template
type DocumentPage struct {
	AppTitle      string
	AppVersion    string
	PageTitle     string
	LoggedinID    int
	UserConfig    team.UserConfig
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
	UserList      []memdb.ObjHasID
}

// DocumentHandler is http handler for document page
func (dd *DocsBase) DocumentHandler(w http.ResponseWriter, r *http.Request) {

	const ADMIN = 1

	allow, id := core.AuthVerify(w, r, dd.memorydb)
	if !allow {
		return
	}

	if dd.validURLs.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}

	var err error

	var Page = DocumentPage{
		AppTitle:     dd.text.AppTitle,
		AppVersion:   core.AppVersion,
		LoggedinID:   id,
		Editable:     false,
		New:          false,
		Categories:   dd.text.Categories,
		DocTypes:     dd.text.DocTypes,
		Currencies:   currencies.GetCurrencies(),
		ApprovalSign: dd.text.ApprovalSign,
	}

	TextID := accs.GetTextIDfromURL(r.URL.Path)
	IntID, _ := strconv.Atoi(TextID)
	if TextID == "new" {
		Page.New = true
	} else {
		Page.Document = Document{ID: IntID}
		err = Page.Document.load(dd.db, dd.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			http.NotFound(w, r)
			return
		}
		Page.Approvals, err = Page.Document.loadApprovals(dd.db, dd.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "loading document approvals", Page.LoggedinID, Page.Document.ID)
			return
		}
		Page.IamApprover = accs.IntToBool(Page.Approvals.getApprovalIDbyDocIDandApproverID(IntID, Page.LoggedinID))
		Page.YouApproved = Page.Approvals.approved(IntID, Page.LoggedinID)
	}

	user := team.UnmarshalToProfile(dd.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig
	if user.UserRole == team.ADMIN || Page.New || Page.Document.GiveCreatorID() == Page.LoggedinID {
		Page.Editable = true
	}
	Page.RemoveAllowed = dd.cfg.removeAllowed
	if user.UserRole == team.ADMIN {
		Page.RemoveAllowed = true
	}

	var created int
	var updated int
	shallBreakApproval := false
	defaultUploadPath := filepath.Join(dd.cfg.serverRoot, "files", "docs", TextID)

	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		r.Body = http.MaxBytesReader(w, r.Body, core.MAX_UPLOAD_SIZE+1048576)
	}

	// Create or update code ==========================================================================
	if r.Method == "POST" && (r.FormValue("createButton") != "" || r.FormValue("updateButton") != "") {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "writing document", Page.LoggedinID, IntID)
			return
		}
		d := Document{
			ID:        IntID,
			RegNo:     r.FormValue("regNo"),
			RegDate:   datetime.StringToDate(r.FormValue("regDate")),
			IncNo:     r.FormValue("incNo"),
			IncDate:   datetime.StringToDate(r.FormValue("incDate")),
			Category:  accs.StrToInt(r.FormValue("category")),
			DocType:   accs.StrToInt(r.FormValue("docType")),
			About:     r.FormValue("about"),
			Authors:   r.FormValue("authors"),
			Addressee: r.FormValue("addressee"),
			DocSum:    currencies.ProcessFormSumInt(r.FormValue("docSum")),
			Currency:  accs.StrToInt(r.FormValue("currencyCode")),
			EndDate:   datetime.StringToDate(r.FormValue("endDate")),
			Creator:   &team.Profile{ID: Page.LoggedinID},
			Note:      r.FormValue("note"),
		}
		if r.FormValue("docSum") != "" && d.Currency == 0 {
			d.Currency = -1
		} else if r.FormValue("docSum") == "" && d.Currency == -1 {
			d.Currency = 0
		}

		d.FileList, err = core.Uploader(r, defaultUploadPath, "fileList")
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			Page.Message = "uploadError"
		}

		if r.FormValue("createButton") != "" && !strings.Contains(Page.Message, "Error") {
			d.ID, created = d.Create(dd.db, dd.dbType)
			if created > 0 {
				core.MoveUploadedFilesToFinalDest(defaultUploadPath,
					filepath.Join(dd.cfg.serverRoot, "files", "docs", strconv.Itoa(d.ID)),
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
			updated = d.update(dd.db, dd.dbType)
			if updated > 0 {
				Page.Message = "dataWritten"
				Page.Document.load(dd.db, dd.dbType)
			} else {
				Page.Message = "dataNotWritten"
			}
		}

	}

	// Delete files ==================================
	if r.Method == "POST" && r.FormValue("deleteFiles") != "" {
		if Page.Editable == false {
			accs.ThrowAccessDenied(w, "writing document", Page.LoggedinID, IntID)
			return
		}
		r.ParseForm()
		filesToRemove := r.Form["filesToRemove"]
		err = core.RemoveUploadedFiles(defaultUploadPath, filesToRemove)
		if err == nil {
			FileList := accs.FilterSliceStrList(Page.Document.FileList, filesToRemove)
			updated = sqla.UpdateSingleJSONListStr(dd.db, dd.dbType, "documents", "FileList", FileList, IntID)
			if updated > 0 {
				Page.Message = "dataWritten"
				shallBreakApproval = true
				Page.Document.load(dd.db, dd.dbType)
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
			pID := accs.StrToInt(r.FormValue("approvalAdd"))
			a := Approval{
				Approver: &team.Profile{ID: pID},
				DocID:    IntID,
			}
			if accs.SliceContainsInt(Page.Approvals.getApproversIDsSlice(), pID) {
				Page.Message = "approverAlreadyInList"
			} else {
				id, res := a.create(dd.db, dd.dbType)
				if res > 0 {
					a.ID = id
					Page.Message = "dataWritten"
					p := team.Profile{ID: pID}
					p.Preload(dd.db, dd.dbType)
					title := Page.Document.makeTitle(dd.i18n.categories, dd.i18n.docTypes, dd.i18n.docCaption)
					mail := core.AnyMail{Subj: dd.i18n.messages.Subj.NewApproval + ": " + title,
						Text:       dd.i18n.messages.Cont.ApproveThis + ": " + title + ". " + dd.i18n.messages.Cont.InfoLink + ": ",
						SomeLink:   dd.cfg.systemURL + "/docs/document/" + strconv.Itoa(Page.Document.ID),
						DoNotReply: dd.i18n.messages.DoNotReply, SystemURL: dd.cfg.systemURL, MailerName: dd.i18n.messages.MailerName}
					mail.ConstructToChannel(dd.db, dd.dbType, dd.mailtmpl, dd.mailchan, p)
				} else {
					Page.Message = "dataNotWritten"
				}
			}
		} else {
			accs.ThrowAccessDenied(w, "adding approval", Page.LoggedinID, IntID)
			return
		}
	}

	// Approve (sign) ===========================================
	if r.Method == "POST" && r.FormValue("approvalSign") != "" {
		if Page.Editable || Page.IamApprover {
			a := Approval{
				ID:       Page.Approvals.getApprovalIDbyDocIDandApproverID(IntID, Page.LoggedinID),
				Written:  datetime.GetCurrentDateTime(),
				Approver: &team.Profile{ID: Page.LoggedinID},
				DocID:    IntID,
				Approved: accs.StrToInt(r.FormValue("approvalSign")),
				Note:     r.FormValue("approvalNote"),
			}
			if Page.YouApproved == APPROVED {
				Page.Message = "dataNotWritten"
			} else {
				if a.Approved == APPROVED {
					a.Approver.Load(dd.db, dd.dbType)
					a.Approver.Login = "no access"
					a.Approver.Passwd = "no access"
					a.ApproverSign = a.Approver.GiveSelfNameJob()
					if a.Approver.GiveUnitID() != 0 {
						a.ApproverSign += "; " + a.Approver.GiveUnitName()
					}
				}
				res := a.sign(dd.db, dd.dbType)
				if res > 0 {
					if Page.Document.Creator != nil && (a.Approved == APPROVED || a.Approved == REJECTED) {
						Page.Message = "dataWritten"
						p := team.Profile{ID: Page.Document.Creator.ID}
						p.Preload(dd.db, dd.dbType)
						ResultSubj := map[int]string{APPROVED: dd.i18n.messages.Subj.Approved, REJECTED: dd.i18n.messages.Subj.Rejected}
						ResultCont := map[int]string{APPROVED: dd.i18n.messages.Cont.Approved, REJECTED: dd.i18n.messages.Cont.Rejected}
						title := Page.Document.makeTitle(dd.i18n.categories, dd.i18n.docTypes, dd.i18n.docCaption)
						mail := core.AnyMail{Subj: ResultSubj[a.Approved] + ": " + title,
							Text:       dd.i18n.docCaption + ": " + title + " " + ResultCont[a.Approved] + " " + dd.i18n.messages.Cont.ByApprover + ": " + a.Approver.GiveSelfNameJob() + ". " + dd.i18n.messages.Cont.InfoLink + ": ",
							SomeLink:   dd.cfg.systemURL + "/docs/document/" + strconv.Itoa(Page.Document.ID),
							DoNotReply: dd.i18n.messages.DoNotReply, SystemURL: dd.cfg.systemURL, MailerName: dd.i18n.messages.MailerName}
						mail.ConstructToChannel(dd.db, dd.dbType, dd.mailtmpl, dd.mailchan, p)
					}
				} else {
					Page.Message = "dataNotWritten"
				}
			}
		} else {
			accs.ThrowAccessDenied(w, "signing approval", Page.LoggedinID, IntID)
			return
		}
	}

	// Remove approval ===========================================
	if r.Method == "POST" && r.FormValue("approvalRemove") != "" {
		aID := accs.StrToInt(r.FormValue("approvalRemove"))
		if Page.Editable && Page.Approvals.getApprovalByID(aID).Approved == NOACTION {
			res := sqla.DeleteObject(dd.db, dd.dbType, "approvals", "ID", aID)
			if res > 0 {
				Page.Message = "dataWritten"
			} else {
				Page.Message = "dataNotWritten"
			}
		} else {
			accs.ThrowAccessDenied(w, "removing approval", Page.LoggedinID, IntID)
			return
		}
	}

	// Reset approvals ===========================================
	if shallBreakApproval {
		sqla.UpdateMultipleWithOneInt(dd.db, dd.dbType, "approvals", "Approved", BROKEN, "Written", datetime.DateTimeToInt64(datetime.GetCurrentDateTime()), Page.Approvals.getApprovalsIDsSliceApproved())
	}

	// Other fields code ============================================
	if TextID == "new" {
		Page.New = true
		Page.PageTitle = dd.text.NewDocument
	} else {
		Page.PageTitle = Page.Document.makeTitle(Page.Categories, Page.DocTypes, dd.text.Document)
		Page.Approvals, err = Page.Document.loadApprovals(dd.db, dd.dbType)
		if err != nil {
			log.Println(accs.CurrentFunction()+":", err)
			accs.ThrowServerError(w, "loading document approvals", Page.LoggedinID, Page.Document.ID)
			return
		}
		Page.IamApprover = accs.IntToBool(Page.Approvals.getApprovalIDbyDocIDandApproverID(IntID, Page.LoggedinID))
		Page.YouApproved = Page.Approvals.approved(IntID, Page.LoggedinID)
	}
	Page.UserList = dd.memorydb.GetObjectArr("UserList")

	template := "document.tmpl"
	if strings.HasSuffix(r.URL.Path, "approval") || strings.HasSuffix(r.URL.Path, "approval/") {
		Page.PageTitle = "Approval list"
		template = "approval.tmpl"
		Page.Document.Creator.Load(dd.db, dd.dbType)
		Page.Document.Creator.Login = "no access"
		Page.Document.Creator.Passwd = "no access"
		Page.Document.Creator.BirthDate.Year = 0
		Page.Document.Creator.UserConfig = team.UserConfig{}
	}

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		return
	}

	// HTML output
	err = dd.templates.ExecuteTemplate(w, template, Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		accs.ThrowServerError(w, "executing "+template[0:len(template)-5]+" template", Page.LoggedinID, Page.Document.ID)
		return
	}

}
