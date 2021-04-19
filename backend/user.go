package backend

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type User struct {
	Uname string `json:"uname"`
	Pass  string `json:"pass"`
	Todos []Todo `json:"todos"`
	ID    int    `gorm:"primaryKey" json:"id"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	// get the req body
	reqBody, _ := ioutil.ReadAll(r.Body)
	// unmarshall and check
	var decodedReqBody struct {
		Uname string
		Pass  string
	}
	err := json.Unmarshal(reqBody, &decodedReqBody)
	if decodedReqBody.Uname == "" || decodedReqBody.Pass == "" || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(ErrUserReqBody))
		return
	}
	// insert into db
	user := User{
		Uname: decodedReqBody.Uname,
		Pass:  decodedReqBody.Pass,
	}
	db.Create(&user)

	// marshall and send
	encodedResBody, _ := json.Marshal(user)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(encodedResBody)
}

func GETUser(w http.ResponseWriter, r *http.Request) {
	// parse the req body
	reqBody, err := ioutil.ReadAll(r.Body)
	if !assertServerError(err, w) {
		return
	}
	decodedResBody := map[string]interface{}{}
	err = json.Unmarshal(reqBody, &decodedResBody)
	if !assertServerError(err, w) {
		return
	}

	uname := decodedResBody["uname"]
	pass := decodedResBody["pass"]

	if uname == nil || pass == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(ErrUserReqBody))
		return
	}

	// get the uname and pass
	// query the db with uname and pass
	var user User
	db.First(&user, "uname=? and pass=?", uname, pass)
	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(ErrInvalidID))
		return
	}

	// marshall and send
	resBody, _ := json.Marshal(user)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resBody)

}
