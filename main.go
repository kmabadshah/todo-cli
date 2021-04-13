package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	//http.ListenAndServe()
}

type Todo struct {
	Text string
	ID   int
}

var (
	db         = InitDB()
	ErrReqBody = "Invalid request body, please include a text field with non-zero length"
)

func TodoWithoutID(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		HandleGET(w, r)
	} else if r.Method == "POST" {
		HandlePOST(w, r)
	}
}

func HandleGET(w http.ResponseWriter, _ *http.Request) {
	var allTodos []Todo
	db.Find(&allTodos)
	encodedData, _ := json.Marshal(allTodos)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(encodedData)
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
	w.WriteHeader(http.StatusOK)
	encodedResBody, _ := json.Marshal(createdTodo)
	_, _ = w.Write(encodedResBody)
}

func InitDB() *gorm.DB {
	// connect to db
	dsn := "host=localhost user=kmab password=kmab dbname=todo_cli_test port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to db")
	}
	return db
}
