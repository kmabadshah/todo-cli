package main

import (
	"bytes"
	"encoding/json"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTodo(t *testing.T) {
	t.Run("on valid req body", func(t *testing.T) {
		reqBody := map[string]string{
			"text":      "Hello World",
			"something": "else",
		}
		reqJSONBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "http://localhost:8080/todos", bytes.NewReader(reqJSONBody))
		res := httptest.NewRecorder()

		TodoWithoutID(res, req)

		t.Run("response body is of type Todo and response status code is ok", func(t *testing.T) {
			var decodedResBody Todo
			err := json.Unmarshal(res.Body.Bytes(), &decodedResBody)
			if err != nil {
				t.Fatalf("Didn't expect an error but got \n%v", err)
			}

			resBodyText := decodedResBody.Text
			resBodyID := decodedResBody.ID
			reqBodyText, _ := reqBody["text"]

			assertStatusCode(t, res.Result().StatusCode, http.StatusOK)

			if resBodyText != reqBodyText || resBodyID == 0 {
				t.Error("response body does not implement type Todo")
			}
		})

		t.Run("todo has been stored into db", func(t *testing.T) {
			var decodedResBody Todo
			err := json.Unmarshal(res.Body.Bytes(), &decodedResBody)
			if err != nil {
				t.Fatalf("Didn't expect an error but got \n%v", err)
			}

			resBodyID := decodedResBody.ID

			// connect to db
			dsn := "host=localhost user=kmab password=kmab dbname=todo_cli_test port=5432"
			db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err != nil {
				t.Fatalf("Could not connect to db")
			}

			var todo Todo
			db.First(&todo, resBodyID)

			if todo == (Todo{}) {
				t.Errorf("didn't find the todo that was created earlier")
			}
		})
	})

	t.Run("on invalid req body", func(t *testing.T) {
		reqBody := map[string]string{
			"invalid": "request-body",
		}
		reqJSONBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "http://localhost:8080/todos", bytes.NewReader(reqJSONBody))
		res := httptest.NewRecorder()

		TodoWithoutID(res, req)

		t.Run("proper response message and status code on invalid req body", func(t *testing.T) {
			got := res.Body.String()
			want := "Invalid request body, please include a text field with non-zero length"

			assertStatusCode(t, res.Result().StatusCode, http.StatusBadRequest)
			if got != want {
				t.Error("didn't get proper response message on invalid req body")
			}
		})
	})
}

// create the server
// ping the server
// check the response code
// check the response body
// check if the code manipulated the db correctly

func TestGETTodos(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost:8080/todos", nil)
	res := httptest.NewRecorder()

	TodoWithoutID(res, req)

	t.Run("returns 200 response code", func(t *testing.T) {
		got := res.Result().StatusCode
		want := http.StatusOK

		if got != want {
			t.Errorf("wanted %#v status but got %#v", want, got)
		}
	})

	t.Run("returns proper response body", func(t *testing.T) {
		var got []byte
		_, _ = res.Result().Body.Read(got)
		// complete this thing
		// truncate the todos table in todo_cli_test when all tests are done
	})
}

func assertStatusCode(t *testing.T, got, want int) {
	if got != want {
		t.Errorf("wanted %#v status but got %#v", want, got)
	}
}
