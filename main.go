package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

func main() {
	router := mux.NewRouter()
	router.Path("/todos").HandlerFunc(TodoWithoutID)
	router.Path("/todos/{id}").HandlerFunc(TodoWithID)

	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

type Todo struct {
	Text string
	ID   int
}

var (
	mode         = "test"
	db           = InitDB()
	ErrReqBody   = "Invalid request body, please include a text field with non-zero length"
	ErrInvalidID = "Invalid id"
	ErrInternal  = "Please try again later"
)

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

func HandleGETAll(w http.ResponseWriter, _ *http.Request) {
	var allTodos []Todo
	db.Find(&allTodos)
	encodedData, _ := json.Marshal(allTodos)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(encodedData)
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

func HandlePOST(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var decodedReqBody map[string]interface{}
	err := json.Unmarshal(reqBody, &decodedReqBody)

	// check if the decodedReqBody includes text field
	if decodedReqBody["text"] == nil || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, ErrReqBody)
		return
	}
	createdTodo := Todo{Text: decodedReqBody["text"].(string)}
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
		_, _ = fmt.Fprint(w, ErrReqBody)
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

func InitDB() *gorm.DB {
	// connect to db
	var dbname string
	if mode == "prod" {
		dbname = "todo_cli"
	} else {
		dbname = "todo_cli_test"
	}
	dsn := fmt.Sprintf("host=localhost user=kmab password=kmab dbname=%s port=5432", dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logger.Silent, // Log level
				IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,          // Disable color
			},
		),
	})

	if err != nil {
		log.Fatalf("Could not connect to db")
	}
	return db
}
