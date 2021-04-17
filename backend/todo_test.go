package backend

import (
	"bytes"
	"encoding/json"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var uid int

// initialize the testing environment for subsequent tests
func TestStart(t *testing.T) {
	TruncateTable(&User{})
	// create the user
	user := User{Uname: "adnan", Pass: "badshah"}
	db.Create(&user)
	uid = user.ID
	// create the secret file
	err := ioutil.WriteFile("/tmp/secret.txt", []byte(strconv.Itoa(user.ID)), 0644)
	assertTestError(err)
}

func TestCreateTodo(t *testing.T) {
	TruncateTable(&Todo{})
	defer TruncateTable(&Todo{})

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
		want := ErrTodoReqBody
		assertStatusCode(t, res.Result().StatusCode, http.StatusBadRequest)
		if got != want {
			t.Error("didn't get proper response message on invalid req body")
		}
	})

	testUserIdentity(t, func() *httptest.ResponseRecorder {
		res, _ := CreateTodoReq(nil)
		return res
	})
}

func TestGetTodos(t *testing.T) {
	TruncateTable(&Todo{})
	defer TruncateTable(&Todo{})
	CreateTodoReq(nil)
	CreateTodoReq(nil)

	// create an arbitrary user and save credentials
	// testUserIdentity() will automatically clear everything
	// at the end
	user := User{Uname: "test", Pass: "test"}
	db.Create(&user)
	err := ioutil.WriteFile("/tmp/secret.txt", []byte(strconv.Itoa(user.ID)), 0644)
	assertTestError(err)
	// create todo for that user
	CreateTodoReq(nil)

	// get request
	req := httptest.NewRequest("GET", "http://localhost:8080/todos", nil)
	res := httptest.NewRecorder()
	TodoWithoutID(res, req)

	t.Run("proper status code and todos", func(t *testing.T) {
		var resBody []Todo
		assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))
		assertStatusCode(t, res.Result().StatusCode, http.StatusOK)

		if len(resBody) != 2 {
			t.Fatalf("incorrect amount of todos for this user, expected %v but got %v", 2, len(resBody))
		}

		for _, v := range resBody {
			if v.UserID != uid {
				t.Errorf("todo uid mismatch for this user, expected %v but got %v", uid, v.ID)
			}
		}
	})

	testUserIdentity(t, func() *httptest.ResponseRecorder {
		url := "http://localhost:8080/todos"
		req := httptest.NewRequest("GET", url, nil)
		res := httptest.NewRecorder()
		TodoWithoutID(res, req)

		return res
	})
}

func TestGetTodo(t *testing.T) {
	TruncateTable(&Todo{})
	defer TruncateTable(&Todo{})
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
	TruncateTable(&Todo{})
	defer TruncateTable(&Todo{})

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
	TruncateTable(&Todo{})
	defer TruncateTable(&Todo{})

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

// clean the testing environment
func TestEnd(t *testing.T) {
	TruncateTable(&User{})
	// delete all  users
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{})
	// remove secret file
	err := os.Remove("/tmp/secret.txt")
	assertTestError(err)
}

// accepts an optional request body and sends a POST request to /todos
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

// checks if there is any user file in the os and if the data is correct
// takes a function that describes how the request should be made.
// valid requests include GET, POST, PUT, DELETE etc
func testUserIdentity(t *testing.T, f func() *httptest.ResponseRecorder) {
	t.Run("throws error message and status code on invalid user", func(t *testing.T) {
		// at this point, the secret file should exist due to TestStart()
		// so, first delete the file
		err := os.Remove("/tmp/secret.txt")
		assertTestError(err)

		// send req to server and assert status code
		res := f()
		assertStatusCode(t, res.Result().StatusCode, 401)

		// assert response body
		got := res.Body.String()
		want := ErrAuth
		if got != want {
			t.Errorf("didn't get proper response body on invalid user")
		}

		// since we removed the file, put everything back to normal
		TestStart(t)
	})
}
