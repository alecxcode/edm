package projs

import (
	"database/sql"
	"edm/internal/core"
	"edm/internal/team"
	"edm/pkg/accs"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alecxcode/sqla"
)

type reqProjStatus struct {
	Proj   int `json:"proj"`
	Status int `json:"status"`
}

// SetProjStatusAPI updates project status
func (pb *ProjsBase) SetProjStatusAPI(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "application/json")
	allow, loggedinID := core.AuthVerifyAPI(w, r, pb.memorydb)
	if !allow {
		return
	}
	var reqObj reqProjStatus
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
	user := team.UnmarshalToProfile(pb.memorydb.GetByID(loggedinID))
	var res int
	if user.UserRole == team.ADMIN || proj.GiveCreatorID() == loggedinID {
		res = sqla.UpdateSingleInt(pb.db, pb.dbType, "projects", "ProjStatus", reqObj.Status, proj.ID)
	} else {
		accs.ThrowAccessDeniedAPI(w, r.URL.Path, loggedinID)
		return
	}
	if res > 0 {
		proj.ProjStatus = reqObj.Status
		json.NewEncoder(w).Encode(proj)
		return
	}
	accs.ThrowServerErrorAPI(w, accs.CurrentFunction()+": updating project", loggedinID, proj.ID)
	return
}
