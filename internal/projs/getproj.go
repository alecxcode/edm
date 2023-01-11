package projs

import (
	"database/sql"
	"edm/internal/core"
	"edm/pkg/accs"
	"encoding/json"
	"fmt"
	"net/http"
)

type reqProj struct {
	Proj int `json:"proj"`
}

// GetProjAPI returns project JSON by project ID
func (pb *ProjsBase) GetProjAPI(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")
	allow, loggedinID := core.AuthVerifyAPI(w, r, pb.memorydb)
	if !allow {
		return
	}
	var reqObj reqProj
	err = json.NewDecoder(r.Body).Decode(&reqObj)
	if err != nil {
		accs.ThrowServerErrorAPI(w, accs.CurrentFunction()+": decoding json request", loggedinID, reqObj.Proj)
		return
	}
	proj := Project{ID: reqObj.Proj}
	err = proj.load(pb.db, pb.dbType)
	if err == sql.ErrNoRows {
		fmt.Fprint(w, `{"error":804,"description":"object not found"}`)
		return
	} else if err != nil {
		accs.ThrowServerErrorAPI(w, accs.CurrentFunction()+": loading project", loggedinID, reqObj.Proj)
		return
	}
	json.NewEncoder(w).Encode(proj)
	return
}
