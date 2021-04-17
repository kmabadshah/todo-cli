package backend

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type Todo struct {
	Text   string
	ID     int `gorm:"primaryKey"`
	UserID int `gorm:"column:uid"`
}

func TodoWithoutID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		HandleGETAll(w, r)
	case "POST":
		HandlePOST(w, r)
	}
}

func TodoWithID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		HandleGETOne(w, r)
	case "PUT":
		HandlePUT(w, r)
	case "DELETE":
		HandleDelete(w, r)
	}
}

func HandleGETOne(w http.ResponseWriter, r *http.Request) {
	todo := GetTodoByID(r)
	if todo.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(ErrInvalidID))
		return
	}
	resBody, _ := json.Marshal(todo)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resBody)
}

func HandleGETAll(w http.ResponseWriter, _ *http.Request) {
	// get the secret user id
	_, err := getUserId(w)
	if err != nil {
		return
	}

	var allTodos []Todo
	db.Find(&allTodos)
	encodedData, _ := json.Marshal(allTodos)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(encodedData)
}

func HandlePOST(w http.ResponseWriter, r *http.Request) {
	// get the secret user id
	uid, err := getUserId(w)
	if err != nil {
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var decodedReqBody map[string]interface{}
	err = json.Unmarshal(reqBody, &decodedReqBody)
	// check if the decodedReqBody includes text field
	if decodedReqBody["text"] == nil || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, ErrTodoReqBody)
		return
	}
	createdTodo := Todo{Text: decodedReqBody["text"].(string), UserID: uid}
	db.Create(&createdTodo)
	encodedResBody, _ := json.Marshal(createdTodo)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(encodedResBody)
}

func HandlePUT(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var decodedReqBody map[string]interface{}
	err := json.Unmarshal(reqBody, &decodedReqBody)

	// check if the decodedReqBody includes text field
	if decodedReqBody["text"] == nil || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, ErrTodoReqBody)
		return
	}

	// update todo using id
	todo := GetTodoByID(r)
	if todo.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(ErrInvalidID))
		return
	}

	todo.Text = decodedReqBody["text"].(string)
	db.Save(&todo)

	// send the response
	encodedResBody, _ := json.Marshal(todo)
	_, _ = w.Write(encodedResBody)
}

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	// get the id
	todo := GetTodoByID(r)
	if todo.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(ErrInvalidID))
		return
	}

	tx := db.Delete(&todo)
	if tx.RowsAffected != 1 {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(ErrInternal))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Successfully deleted id " + strconv.Itoa(todo.ID)))
}

func GetTodoByID(r *http.Request) Todo {
	var id string

	if mode == "prod" {
		id = mux.Vars(r)["id"]
	} else {
		re := regexp.MustCompile(`/todos/(.*)/?`)
		id = string(re.FindSubmatch([]byte(r.URL.Path))[1])
	}

	var todo Todo
	db.First(&todo, "id=?", id)

	return todo
}

func StartServer() {
	router := mux.NewRouter()
	router.Path("/todos").HandlerFunc(TodoWithoutID)
	router.Path("/todos/{id}").HandlerFunc(TodoWithID)

	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getUserId(w http.ResponseWriter) (int, error) {
	// get the secret user id
	data, err := ioutil.ReadFile("/tmp/secret.txt")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("could not authenticate user"))
		return 0, err
	}
	uid, err := strconv.Atoi(string(data))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("could not authenticate user"))
		return 0, err
	}

	return uid, nil
}
