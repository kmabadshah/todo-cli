package main

import (
	"bytes"
	"encoding/json"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestGetTodos(t *testing.T) {
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

func TestGetTodo(t *testing.T) {
	TruncateTable()
	defer TruncateTable()
	resBody := Todo{}

	// POST a todo
	reqBody := map[string]interface{}{
		"text": "Hello World",
	}
	res, _ := CreateTodoReq(reqBody)
	assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))

	t.Run("GET with valid id", func(t *testing.T) {
		// GET the todo
		req := httptest.NewRequest("GET", "http://localhost:8080/todos/"+strconv.Itoa(resBody.ID), nil)
		res = httptest.NewRecorder()
		TodoWithID(res, req)
		assertStatusCode(t, res.Result().StatusCode, http.StatusOK)

		// Check the result
		resBody = Todo{}
		assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))

		if resBody.Text != reqBody["text"] {
			t.Errorf("Did not get the requested todo text")
		}
	})

	t.Run("GET with invalid id", func(t *testing.T) {
		// GET the todo
		req := httptest.NewRequest("GET", "http://localhost:8080/todos/"+"-1", nil)
		res = httptest.NewRecorder()
		TodoWithID(res, req)
		assertStatusCode(t, res.Result().StatusCode, http.StatusNotFound)

		if res.Body.String() != ErrInvalidID {
			t.Errorf("Expected error but got %#v", res.Body.String())
		}
	})
}

func TestUpdateTodo(t *testing.T) {
	TruncateTable()
	defer TruncateTable()

	res, _ := CreateTodoReq(nil)
	var resBody Todo
	assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))
	updatedTodo := map[string]interface{}{"text": "yo adnan"}
	reqJSONBody, _ := json.Marshal(updatedTodo)

	t.Run("valid id", func(t *testing.T) {
		res = httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "http://localhost:8080/todos/"+strconv.Itoa(resBody.ID), bytes.NewReader(reqJSONBody))
		TodoWithID(res, req)

		t.Run("Proper res code and body with valid ID", func(t *testing.T) {
			assertStatusCode(t, res.Result().StatusCode, http.StatusOK)
			assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))

			t.Log("http://localhost:8080/todos/" + strconv.Itoa(resBody.ID))
			if resBody.Text != updatedTodo["text"] {
				t.Errorf("Did not get the updated text with response")
			}
		})

		t.Run("Check if the todo got updated", func(t *testing.T) {
			var todo Todo
			db.First(&todo, "id=?", resBody.ID)

			if todo.Text != updatedTodo["text"] {
				t.Errorf("Todo did not get updated")
			}
		})
	})

	t.Run("invalid id", func(t *testing.T) {
		res = httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "http://localhost:8080/todos/"+strconv.Itoa(10), bytes.NewReader(reqJSONBody))
		TodoWithID(res, req)

		assertStatusCode(t, res.Result().StatusCode, http.StatusNotFound)
		got := res.Body.String()
		want := ErrInvalidID

		if got != want {
			t.Errorf("Error message mismatch, wanted %#v, got %#v", want, got)
		}
	})

}

func TestDeleteTodo(t *testing.T) {
	TruncateTable()
	defer TruncateTable()

	// create the todo
	res, _ := CreateTodoReq(nil)
	var resBody Todo
	assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))
	todoID := strconv.Itoa(resBody.ID)

	t.Run("valid id", func(t *testing.T) {
		// delete the todo
		req := httptest.NewRequest("DELETE", "http://localhost:8080/todos/"+todoID, nil)
		res = httptest.NewRecorder()
		TodoWithID(res, req)

		t.Run("proper status code and body", func(t *testing.T) {
			assertStatusCode(t, res.Result().StatusCode, http.StatusOK)
			got := res.Body.String()
			want := "Successfully deleted id " + todoID

			if got != want {
				t.Errorf("No proper deletion message; wanted %#v, got %#v", want, got)
			}
		})

		t.Run("check db for todo deletion", func(t *testing.T) {
			var todo Todo
			db.First(todo, "id=?", resBody.ID)
			if todo.ID != 0 {
				t.Errorf("Didn't expect todo to exist in the db")
			}
		})
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "http://localhost:8080/todos/"+strconv.Itoa(-1), nil)
		res = httptest.NewRecorder()
		TodoWithID(res, req)

		assertStatusCode(t, res.Result().StatusCode, http.StatusNotFound)
		got := res.Body.String()
		want := ErrInvalidID

		if got != want {
			t.Errorf("Didn't get proper error message")
		}
	})
}

func TestIntegration(t *testing.T) {
	TruncateTable()
	defer TruncateTable()

	// create 2 todos
	reqBody1 := map[string]interface{}{
		"text":  "integration create todo 1",
		"hello": "world",
	}

	reqBody2 := map[string]interface{}{"text": "integration create todo 1"}
	res1, _ := CreateTodoReq(reqBody1)
	_, _ = CreateTodoReq(reqBody2)

	t.Run("get all todos and check", func(t *testing.T) {
		var todos []Todo
		db.Find(&todos)

		if len(todos) != 2 {
			t.Fatalf("Created 2 todos but got %#v", len(todos))
		}
	})

	// get todo one
	var resBody Todo
	assertRandomErr(t, json.Unmarshal(res1.Body.Bytes(), &resBody))
	req := httptest.NewRequest("GET", "http://localhost:8080/todos/"+strconv.Itoa(resBody.ID), nil)
	res := httptest.NewRecorder()

	if resBody.ID == 0 {
		t.Fatalf("Todo didn't get created")
	}

	// update todo one
	updatedTodo := map[string]interface{}{
		"text": "integration update todo 1",
	}
	encodedReqBody, _ := json.Marshal(updatedTodo)
	req = httptest.NewRequest("PUT", "http://localhost:8080/todos/"+strconv.Itoa(resBody.ID), bytes.NewReader(encodedReqBody))
	res = httptest.NewRecorder()
	TodoWithID(res, req)

	// decode and check
	assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))
	if resBody.Text != updatedTodo["text"] {
		t.Errorf("didn't update todo properly")
	}

	// delete todo one
	url := "http://localhost:8080/todos/" + strconv.Itoa(resBody.ID)
	req = httptest.NewRequest("DELETE", url, nil)
	res = httptest.NewRecorder()
	TodoWithID(res, req)

	// check if deleted
	var td Todo
	db.Find(&td, "id=?", resBody.ID)
	if td.ID != 0 {
		t.Errorf("Expected todo to be deleted")
	}
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

func assertRandomErr(t *testing.T, err interface{}) {
	t.Helper()
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

// Truncates the Todo table in test database
func TruncateTable() {
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Todo{})
}
