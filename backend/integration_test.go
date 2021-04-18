package backend

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestIntegration(t *testing.T) {
	TruncateTable(&Todo{})
	defer TruncateTable(&Todo{})

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
