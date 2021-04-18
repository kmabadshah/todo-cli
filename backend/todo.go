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

func userMiddleware(w http.ResponseWriter, _ *http.Request) {
	// check if the secret exists
	data, err := ioutil.ReadFile("/tmp/secret.txt")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(ErrAuth))
		return
	}
	// check if the found data is valid
	var user User
	db.First(&user, "id=?", string(data))
	// if not, send error code and body
	if user.ID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(ErrAuth))
	}
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
	// get the uid
	uid, err := getUserId(w)
	if err != nil {
		return
	}

	todo := GetTodoByID(uid, r)
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
	uid, err := getUserId(w)
	if err != nil {
		return
	}

	// use the user id to get data from todo table
	var todos []Todo
	db.Find(&todos, "uid=?", uid)

	encodedData, _ := json.Marshal(todos)
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
	uid, err := getUserId(w)
	if err != nil {
		return
	}
	todo := GetTodoByID(uid, r)
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
	uid, err := getUserId(w)
	if err != nil {
		return
	}
	// get the id
	todo := GetTodoByID(uid, r)
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

func GetTodoByID(uid int, r *http.Request) Todo {
	var id string

	if mode == "prod" {
		id = mux.Vars(r)["id"]
	} else {
		re := regexp.MustCompile(`/todos/(.*)/?`)
		id = string(re.FindSubmatch([]byte(r.URL.Path))[1])
	}

	var todo Todo
	db.First(&todo, "id=? and uid=?", id, uid)

	return todo
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

func StartServer() {
	router := mux.NewRouter()
	router.Path("/todos").HandlerFunc(TodoWithoutID)
	router.Path("/todos/{id}").HandlerFunc(TodoWithID)

	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
