package backend

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestIntegration(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("set INTEGRATION to run this test")
	}

	cleanTestEnvironment()
	defer cleanTestEnvironment()

	user := User{}
	t.Run("create and get user", func(t *testing.T) {
		// create the user
		reqBody := map[string]string{
			"uname": "adnan", "pass": "badshah",
		}
		encodedReqBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "http://localhost:8080/users", bytes.NewReader(encodedReqBody))
		res := httptest.NewRecorder()
		CreateUser(res, req)
		createdUser := User{}
		err := json.Unmarshal(res.Body.Bytes(), &createdUser)
		assertTestError(err)

		// get the user
		req = httptest.NewRequest("GET", "http://localhost:8080/users/"+strconv.Itoa(createdUser.ID), nil)
		res = httptest.NewRecorder()
		GETUser(res, req)
		gotUser := User{}
		err = json.Unmarshal(res.Body.Bytes(), &gotUser)
		assertTestError(err)
		if gotUser.Uname != reqBody["uname"] {
			t.Fatalf("User not created properly")
		}

		user = gotUser
	})

	t.Run("CRUD todos", func(t *testing.T) {
		logIn(user)

		// create 2 todos
		res1, _ := CreateTodoReq(nil)
		todo1 := Todo{}
		err := json.Unmarshal(res1.Body.Bytes(), &todo1)
		assertTestError(err)

		res2, _ := CreateTodoReq(nil)
		todo2 := Todo{}
		err = json.Unmarshal(res2.Body.Bytes(), &todo2)
		assertTestError(err)

		// update todo1 and check
		updatedTodoReqBody := map[string]string{"text": "updated todo"}
		updateTodo(todo1.ID, updatedTodoReqBody)

		todo1 = getTodo(todo1.ID)
		if todo1.Text != updatedTodoReqBody["text"] {
			t.Fatalf("todo not updated properly")
		}

		// delete the todos
		deleteTodo(todo1.ID)
		deleteTodo(todo2.ID)

	})
}

func updateTodo(id int, reqBody interface{}) {
	encodedReqBody, _ := json.Marshal(reqBody)
	TodoWithID(
		httptest.NewRecorder(),
		httptest.NewRequest(
			"PUT",
			"http://localhost:8080/todos/"+strconv.Itoa(id),
			bytes.NewReader(encodedReqBody),
		),
	)
}

func getTodo(id int) Todo {
	res := httptest.NewRecorder()
	TodoWithID(
		res,
		httptest.NewRequest(
			"GET",
			"http://localhost:8080/todos/"+strconv.Itoa(id),
			nil,
		),
	)
	todo := Todo{}
	err := json.Unmarshal(res.Body.Bytes(), &todo)
	assertTestError(err)

	return todo
}

func deleteTodo(id int) {
	TodoWithoutID(
		httptest.NewRecorder(),
		httptest.NewRequest(
			"DELETE",
			"http://localhost:8080/todos/"+strconv.Itoa(id),
			nil,
		),
	)
}
