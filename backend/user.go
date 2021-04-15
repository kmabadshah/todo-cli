package backend

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	// get the req body
	reqBody, _ := ioutil.ReadAll(r.Body)
	// unmarshall
	var decodedReqBody struct {
		Uname string
		Pass  string
	}
	err := json.Unmarshal(reqBody, &decodedReqBody)
	if decodedReqBody.Uname == "" || decodedReqBody.Pass == "" || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body, must have a uname and pass field"))
		return
	}
	// insert into db
	db.Create(&decodedReqBody)
	// marshall
	encodedResBody, _ := json.Marshal(decodedReqBody)
	// write header
	w.WriteHeader(http.StatusOK)
	// write res
	_, _ = w.Write(encodedResBody)
}
