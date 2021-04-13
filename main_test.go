package main

import (
	"bytes"
	"encoding/json"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTodo(t *testing.T) {
	TruncateTable()
	defer TruncateTable()

	t.Run("on valid req body", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"text":      "Hello World",
			"something": "else",
		}
		res, _ := CreateTodoReq(reqBody)

		var resBody Todo
		assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))
		resBodyText := resBody.Text
		resBodyID := resBody.ID
		reqBodyText, _ := reqBody["text"]

		t.Run("response body is of type Todo and response status code is ok", func(t *testing.T) {
			assertStatusCode(t, res.Result().StatusCode, http.StatusOK)
			if resBodyText != reqBodyText || resBodyID == 0 {
				t.Error("response body does not implement type Todo")
			}
		})

		t.Run("todo has been stored into db", func(t *testing.T) {
			var todo Todo
			db.First(&todo, resBodyID)
			if todo == (Todo{}) {
				t.Errorf("didn't find the todo that was created earlier")
			}
		})
	})

	t.Run("proper response message and status code on invalid req body", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"invalid": "request-body",
		}
		res, _ := CreateTodoReq(reqBody)

		got := res.Body.String()
		want := ErrReqBody
		assertStatusCode(t, res.Result().StatusCode, http.StatusBadRequest)
		if got != want {
			t.Error("didn't get proper response message on invalid req body")
		}
	})
}

func TestGETTodos(t *testing.T) {
	TruncateTable()
	defer TruncateTable()
	CreateTodoReq(nil)
	CreateTodoReq(nil)

	// get request
	req := httptest.NewRequest("GET", "http://localhost:8080/todos", nil)
	res := httptest.NewRecorder()
	TodoWithoutID(res, req)

	t.Run("returns all todos and proper status code", func(t *testing.T) {
		var resBody []Todo
		assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))
		assertStatusCode(t, res.Result().StatusCode, http.StatusOK)
		if len(resBody) != 2 {
			t.Errorf("didn't get same amount of todos that was created")
		}
	})
}

func CreateTodoReq(reqBody map[string]interface{}) (*httptest.ResponseRecorder, *http.Request) {
	if reqBody == nil {
		reqBody = map[string]interface{}{
			"text":      "GET TODOS TEST",
			"something": "else",
		}
	}
	reqJSONBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "http://localhost:8080/todos", bytes.NewReader(reqJSONBody))
	res := httptest.NewRecorder()
	TodoWithoutID(res, req)

	return res, req
}

func TestTruncate(t *testing.T) {
	TruncateTable()
}

func assertRandomErr(t *testing.T, err interface{}) {
	if err != nil {
		t.Fatalf("Didn't expect an error but got \n%v", err)
	}
}
func assertStatusCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("wanted %#v status but got %#v", want, got)
	}
}

func TruncateTable() {
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Todo{})
}
