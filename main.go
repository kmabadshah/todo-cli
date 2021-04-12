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

func TodoWithoutID(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		HandleGET(w, r)
	} else if r.Method == "POST" {
		HandlePOST(w, r)
	}
}

func HandleGET(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func HandlePOST(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var decodedReqBody map[string]interface{}
	_ = json.Unmarshal(reqBody, &decodedReqBody)

	// check if the decodedReqBody includes text field
	if decodedReqBody["text"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "Invalid request body, please include a text field with non-zero length")
		return
	}

	// init db
	dsn := "host=localhost user=kmab password=kmab dbname=todo_cli_test port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to db")
	}

	// store into db
	createdTodo := Todo{Text: decodedReqBody["text"].(string)}
	db.Create(&createdTodo)

	w.WriteHeader(http.StatusOK)
	encodedResBody, _ := json.Marshal(createdTodo)
	_, _ = w.Write(encodedResBody)
}
