package backend

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAddUser(t *testing.T) {
	cleanTestEnvironment()
	defer cleanTestEnvironment()

	t.Run("on valid req body", func(t *testing.T) {
		reqBody := struct {
			Uname string
			Pass  string
		}{
			Uname: "adnan",
			Pass:  "badshah",
		}
		res, _ := RequestCreateUser(reqBody)

		// check response status and body type
		decodedResBody := unmarshalAndAssert(t, res)

		// check if the user was actually created
		var user User
		db.First(&user, "id=?", decodedResBody.ID)
		if user.Uname != reqBody.Uname {
			t.Errorf("User was not created")
		}
	})

	t.Run("on invalid req body", func(t *testing.T) {
		reqBody := struct {
			Pass string
		}{
			Pass: "badshah",
		}
		res, _ := RequestCreateUser(reqBody)

		// check response status and body text
		got := res.Body.String()
		want := ErrUserReqBody
		assertStatusCode(t, res.Result().StatusCode, http.StatusBadRequest)

		if got != want {
			t.Errorf("Response body does not adhere to req body")
		}
	})
}

func TestGETUser(t *testing.T) {
	cleanTestEnvironment()
	defer cleanTestEnvironment()

	reqBody := struct {
		Uname string
		Pass  string
	}{
		Uname: "adnan",
		Pass:  "badshah",
	}
	res, _ := RequestCreateUser(reqBody)

	// check response status and body type
	var decodedResBody User
	assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &decodedResBody))
	assertStatusCode(t, res.Result().StatusCode, http.StatusOK)

	t.Run("valid req id", func(t *testing.T) {
		// get request
		url := "http://localhost:8080/users/" + strconv.Itoa(decodedResBody.ID)
		req := httptest.NewRequest("GET", url, nil)
		res = httptest.NewRecorder()
		GETUser(res, req)

		// unmarshall and check
		decodedResBody = unmarshalAndAssert(t, res)
		if decodedResBody.Uname != reqBody.Uname {
			t.Errorf("Did not GET the user as expected")
		}
	})

	t.Run("invalid req id", func(t *testing.T) {
		// get request
		url := "http://localhost:8080/users/" + strconv.Itoa(-1)
		req := httptest.NewRequest("GET", url, nil)
		res = httptest.NewRecorder()
		GETUser(res, req)
		assertStatusCode(t, res.Result().StatusCode, http.StatusNotFound)

		got := res.Body.String()
		want := ErrInvalidID
		if got != want {
			t.Errorf("didn't get proper response, wanted %s but got %s", want, got)
		}
	})
}

func RequestCreateUser(reqBody interface{}) (*httptest.ResponseRecorder, *http.Request) {
	// marshall
	encodedReqBody, _ := json.Marshal(reqBody)
	// send request
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/users", bytes.NewReader(encodedReqBody))
	res := httptest.NewRecorder()
	CreateUser(res, req)

	return res, req
}

func checkIfSecretFileStored(t *testing.T) {
	// check if there is a secret file stored
	_, err := ioutil.ReadFile("/tmp/secret.txt")
	if err != nil {
		t.Errorf("Secret file has not been stored")
	}
}
