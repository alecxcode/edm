package core

import (
	"edm/pkg/accs"
	"edm/pkg/memdb"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
)

// CoreBase is a struct which methods are http handlers
type CoreBase struct {
	cfg struct {
		startPage    string
		serverSystem string
	}
	validURLs *regexp.Regexp
	memorydb  memdb.ObjectsInMemory
	uploads   http.Handler
}

// NewCoreBase is a constructor
func NewCoreBase(startPage string, serverSystem string, uploadPath string,
	memorydb memdb.ObjectsInMemory) CoreBase {
	return CoreBase{
		cfg: struct {
			startPage    string
			serverSystem string
		}{
			startPage:    startPage,
			serverSystem: serverSystem,
		},
		validURLs: regexp.MustCompile("^/?$"),
		memorydb:  memorydb,
		uploads:   http.StripPrefix("/files/", http.FileServer(http.Dir(uploadPath))),
	}
}

// AuthVerify checks if a user is logged in
func AuthVerify(w http.ResponseWriter, r *http.Request, memorydb memdb.ObjectsInMemory) (res bool, id int) {
	thecookie, err := r.Cookie("sessionid")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false, 0
	}
	allow, id := memorydb.CheckSession(thecookie.Value)
	if !allow {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false, 0
	}
	return true, id
}

// AuthVerifyAPI checks if a user is logged in
func AuthVerifyAPI(w http.ResponseWriter, r *http.Request, memorydb memdb.ObjectsInMemory) (res bool, id int) {
	thecookie, err := r.Cookie("sessionid")
	if err == http.ErrNoCookie {
		accs.ThrowAccessDeniedAPI(w, r.URL.Path, 0)
		return
	}
	allow, id := memorydb.CheckSession(thecookie.Value)
	if !allow {
		accs.ThrowAccessDeniedAPI(w, r.URL.Path, id)
		return
	}
	return true, id
}

// IndexHandler handles the root page
func (cb *CoreBase) IndexHandler(w http.ResponseWriter, r *http.Request) {
	allow, _ := AuthVerify(w, r, cb.memorydb)
	if !allow {
		return
	}
	if cb.validURLs.FindStringSubmatch(r.URL.Path) == nil {
		http.NotFound(w, r)
		return
	}
	switch cb.cfg.startPage {
	case "docs":
		http.Redirect(w, r, "/docs/", http.StatusSeeOther)
	case "tasks":
		http.Redirect(w, r, "/tasks/", http.StatusSeeOther)
	case "team":
		http.Redirect(w, r, "/team/", http.StatusSeeOther)
	case "portal":
		http.Redirect(w, r, "/portal/", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/docs/", http.StatusSeeOther)
	}
}

// PortalHandler is a stub for now
func PortalHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This function is under development. Current URL: %s", r.URL.Path)
}

// GetAppVersion returns app version in JSON
func GetAppVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Ver string `json:"ver"`
	}{AppVersion})
}

// ServeFavIcon serve favicon image
func (cb *CoreBase) ServeFavIcon(w http.ResponseWriter, r *http.Request) {
	favIconPath := filepath.Join(cb.cfg.serverSystem, "static", "favicon.png")
	http.ServeFile(w, r, favIconPath)
}

// ServeUploads serve uploaded files
func (cb *CoreBase) ServeUploads(w http.ResponseWriter, r *http.Request) {
	allow, _ := AuthVerify(w, r, cb.memorydb)
	if !allow {
		return
	}
	cb.uploads.ServeHTTP(w, r)
}
