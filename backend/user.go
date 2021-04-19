package backend

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

type User struct {
	Uname string
	Pass  string
	Todo  []Todo
	ID    int `gorm:"primaryKey"`
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
	// get the req id
	id := ExtractID(r)
	// query db
	var user User
	db.First(&user, "id=?", id)
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

func logIn(user User) {
	// store into secret file
	err := ioutil.WriteFile("/tmp/secret.txt", []byte(strconv.Itoa(user.ID)), 0644)
	if err != nil {
		panic(err)
	}
}
