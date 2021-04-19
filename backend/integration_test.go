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

	var user map[string]interface{}
	t.Run("create and get user", func(t *testing.T) {
		// create the user
		reqBody := map[string]string{
			"uname": "adnan", "pass": "badshah",
		}
		encodedReqBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(
			"POST",
			"http://localhost:8080/users",
			bytes.NewReader(encodedReqBody),
		)
		res := httptest.NewRecorder()
		CreateUser(res, req)
		// decode
		var createdUser map[string]interface{}
		err := json.Unmarshal(res.Body.Bytes(), &createdUser)
		assertTestError(err)

		// get the user
		req = httptest.NewRequest("GET", "http://localhost:8080/users", bytes.NewReader(encodedReqBody))
		res = httptest.NewRecorder()
		GETUser(res, req)
		// decode and check
		var gotUser map[string]interface{}
		assertTestError(json.Unmarshal(res.Body.Bytes(), &gotUser))
		if gotUser["uname"] != reqBody["uname"] {
			t.Fatalf("User not created properly")
		}

		user = gotUser
	})

	t.Run("CRUD todos", func(t *testing.T) {
		LogIn(user)

		// create 2 todos
		res1, _ := CreateTodoReq(nil)
		var todo1 map[string]interface{}
		assertTestError(json.Unmarshal(res1.Body.Bytes(), &todo1))
		id1 := int(todo1["id"].(float64))

		res2, _ := CreateTodoReq(nil)
		var todo2 map[string]interface{}
		assertTestError(json.Unmarshal(res2.Body.Bytes(), &todo2))
		id2 := int(todo2["id"].(float64))

		// update todo1 and check
		updatedTodoReqBody := map[string]string{"text": "updated todo"}
		updateTodo(id1, updatedTodoReqBody)

		todo1 = getTodo(id1)
		if todo1["text"] != updatedTodoReqBody["text"] {
			t.Fatalf("todo not updated properly")
		}

		// delete the todos
		deleteTodo(id1)
		deleteTodo(id2)

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

func getTodo(id int) map[string]interface{} {
	res := httptest.NewRecorder()
	TodoWithID(
		res,
		httptest.NewRequest(
			"GET",
			"http://localhost:8080/todos/"+strconv.Itoa(id),
			nil,
		),
	)
	var todo map[string]interface{}
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
