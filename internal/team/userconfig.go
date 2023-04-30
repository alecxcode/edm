package team

import (
	"edm/internal/core"
	"edm/pkg/accs"
	"encoding/json"
	"log"
	"net/http"
)

// UserConfigPage is passed into template
type UserConfigPage struct {
	AppTitle    string
	AppVersion  string
	PageTitle   string
	LoggedinID  int
	Message     string
	UserConfig  UserConfig
	Themes      []string
	DateFormats []string
	TimeFormats []string
	LangCodes   []string
}

// UserConfigHandler is http handler for config page
func (tb *TeamBase) UserConfigHandler(w http.ResponseWriter, r *http.Request) {

	allow, id := core.AuthVerify(w, r, tb.memorydb)
	if !allow {
		return
	}

	var Page = UserConfigPage{
		AppTitle:   tb.text.AppTitle,
		AppVersion: core.AppVersion,
		LoggedinID: id,
	}

	var err error
	var updated int

	// Update code =============================================
	if r.Method == "POST" && r.FormValue("updateButton") != "" {
		p := Profile{ID: Page.LoggedinID}
		p.UserConfig = UserConfig{
			SystemTheme:           r.FormValue("systemTheme"),
			ElemsOnPage:           accs.StrToInt(r.FormValue("elemsOnPage")),
			ElemsOnPageTeam:       accs.StrToInt(r.FormValue("elemsOnPageTeam")),
			DateFormat:            r.FormValue("dateFormat"),
			TimeFormat:            r.FormValue("timeFormat"),
			LangCode:              r.FormValue("langCode"),
			UseCalendarInControls: accs.StrToBool(r.FormValue("useCalendarInControls")),
			CurrencyBeforeAmount:  accs.StrToBool(r.FormValue("currencyBeforeAmount")),
			ShowFinishedTasks:     accs.StrToBool(r.FormValue("showFinishedTasks")),
			ReturnAfterCreation:   accs.StrToBool(r.FormValue("returnAfterCreation")),
		}
		updated = p.UpdateConfig(tb.db, tb.dbType)
		if updated > 0 {
			MemoryUpdateProfile(tb.db, tb.dbType, tb.memorydb, p.ID)
			Page.Message = "configSaved"
		}
	}

	// Loading code ============================================
	Page.PageTitle = tb.text.ConfigPageTitle
	Page.Themes = tb.options.Themes
	Page.DateFormats = tb.options.DateFormats
	Page.TimeFormats = tb.options.TimeFormats
	Page.LangCodes = tb.options.LangCodes

	user := UnmarshalToProfile(tb.memorydb.GetByID(Page.LoggedinID))
	Page.UserConfig = user.UserConfig

	// JSON output
	if r.URL.Query().Get("api") == "json" || r.FormValue("api") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page)
		return
	}

	// HTML output
	err = tb.templates.ExecuteTemplate(w, "config.tmpl", Page)
	if err != nil {
		log.Println(accs.CurrentFunction()+":", err)
		accs.ThrowServerError(w, "executing config template", Page.LoggedinID, Page.LoggedinID)
		return
	}

}
