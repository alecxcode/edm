package projs

import (
	"edm/internal/core"
	"edm/pkg/accs"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type reqProjWS struct {
	Project int `json:"project"`
}

// WsHandler handles websockets connections to the Project base
func (pb *ProjsBase) WsHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	w.Header().Set("Content-Type", "application/json")
	allow, loggedinID := core.AuthVerifyAPI(w, r, pb.memorydb)
	if !allow {
		return
	}

	conn, err := pb.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(accs.CurrentFunction()+": loggedin ID:", loggedinID, " websocket upgrade:", err)
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	var reqObj reqProjWS
	var textID string
	var data []byte
	var prevData []byte
	var pingInterval = time.Duration(60) * time.Second
	var lastTime = time.Now()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println(accs.CurrentFunction()+": loggedin ID:", loggedinID, " websocket read:", err)
		return
	}

	err = json.Unmarshal(msg, &reqObj)
	if err != nil {
		log.Println(w, accs.CurrentFunction()+": decoding json request:", loggedinID, reqObj.Project)
		return
	}
	textID = strconv.Itoa(reqObj.Project)

	for {
		time.Sleep(time.Duration(125) * time.Millisecond)
		data = pb.memorydb.GetRaw("project:" + textID)
		if len(data) > 1 {
			if string(data) == string(prevData) {
				continue
			}
			prevData = data
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				if core.DEBUG {
					log.Println(accs.CurrentFunction()+": loggedin ID:", loggedinID, " websocket write:", err)
				}
				break
			}
		}
		if time.Since(lastTime) > pingInterval {
			lastTime = time.Now()
			if err = conn.WriteMessage(websocket.PongMessage, []byte{}); err != nil {
				break
			}
		}
	}
}
