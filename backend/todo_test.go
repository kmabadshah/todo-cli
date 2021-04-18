package backend

import (
	"bytes"
	"encoding/json"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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
}

func TestUserMiddleware(t *testing.T) {
	// call the userMiddleware()
	// check for errors
	t.Run("does not throw error when valid user is present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://localhost:8080/todos", nil)
		res := httptest.NewRecorder()
		userMiddleware(res, req)

		// should not write anything to res
		got := res.Body.Bytes()
		var want []byte
		if !reflect.DeepEqual(got, want) {
			t.Errorf("didn't expect anything to be written in the res body, but got %#v", string(got))
		}
	})

	t.Run("throws error when no/invalid user is present", func(t *testing.T) {
		// remove the current user
		err := os.Remove("/tmp/secret.txt")
		assertTestError(err)

		req := httptest.NewRequest("GET", "http://localhost:8080/todos", nil)
		res := httptest.NewRecorder()
		userMiddleware(res, req)

		// should get proper error response code and body
		assertStatusCode(t, res.Result().StatusCode, http.StatusUnauthorized)
		got := res.Body.String()
		want := ErrAuth

		if got != want {
			t.Errorf("expected an error on missing/invalid user")
		}
	})

	TestStart(t)
}

func TestGetTodos(t *testing.T) {
	TruncateTable(&Todo{})
	defer TruncateTable(&Todo{})
	CreateTodoReq(nil)
	CreateTodoReq(nil)
	_ = addRandomUserAndTodo()
	defer TestStart(t)

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
}

func TestGetTodo(t *testing.T) {
	TruncateTable(&Todo{})
	defer TruncateTable(&Todo{})
	resBody := Todo{}

	// create a todo for current user
	reqBody := map[string]interface{}{
		"text": "Hello World",
	}
	res, _ := CreateTodoReq(reqBody)
	assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &resBody))

	t.Run("GET with valid id for current user", func(t *testing.T) {
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

		// check the result
		assertStatusCode(t, res.Result().StatusCode, http.StatusNotFound)
		if res.Body.String() != ErrInvalidID {
			t.Errorf("Expected error but got %#v", res.Body.String())
		}
	})

	t.Run("One User is not able to GET todo of others", func(t *testing.T) {
		todo := addRandomUserAndTodo()
		defer TestStart(t)

		// send the request to /todos/todoId
		url := "http://localhost:8080/todos/" + strconv.Itoa(todo.ID)
		req := httptest.NewRequest("GET", url, nil)
		res := httptest.NewRecorder()
		TodoWithID(res, req)

		// check if the output contains todos
		// since this todo does not belong to the current user, nothing should be returned
		got := res.Body.String()
		want := ErrInvalidID

		assertStatusCode(t, res.Result().StatusCode, http.StatusNotFound)
		if got != want {
			t.Errorf("wanted %#v but got %#v", want, got)
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

// create an arbitrary user, create a todo for him, then switch back to previous user
// returns the todo for the arbitrary user
func addRandomUserAndTodo() Todo {
	// arbitrary user
	user := User{Uname: "test", Pass: "test"}
	db.Create(&user)
	err := ioutil.WriteFile("/tmp/secret.txt", []byte(strconv.Itoa(user.ID)), 0644)
	assertTestError(err)
	// create todo for that user
	res, _ := CreateTodoReq(nil)
	// unmarshall the todo
	var todo Todo
	err = json.Unmarshal(res.Body.Bytes(), &todo)
	assertTestError(err)
	// switch back to our main user
	err = ioutil.WriteFile("/tmp/secret.txt", []byte(strconv.Itoa(uid)), 0644)
	assertTestError(err)

	return todo
}
