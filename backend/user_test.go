package backend

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddUser(t *testing.T) {
	t.Run("on valid req body", func(t *testing.T) {
		reqBody := struct {
			Uname string
			Pass  string
		}{
			Uname: "adnan",
			Pass:  "badshah",
		}
		res, _ := RequestCreateUser(reqBody)

		t.Run("check response status and body", func(t *testing.T) {
			var decodedResBody User
			assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &decodedResBody))
			assertStatusCode(t, res.Result().StatusCode, http.StatusOK)

			if decodedResBody.Uname != reqBody.Uname || decodedResBody.Pass != reqBody.Pass {
				t.Errorf("Response body does not implement req body")
			}
		})
	})

	t.Run("on invalid req body", func(t *testing.T) {
		reqBody := struct {
			Pass string
		}{
			Pass: "badshah",
		}
		res, _ := RequestCreateUser(reqBody)

		t.Run("check response status and body", func(t *testing.T) {
			//var decodedResBody User
			//assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &decodedResBody))
			got := res.Body.String()
			want := "invalid request body, must have a uname and pass field"
			assertStatusCode(t, res.Result().StatusCode, http.StatusBadRequest)

			if got != want {
				t.Errorf("Response body does not adhere to req body")
			}
		})
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
