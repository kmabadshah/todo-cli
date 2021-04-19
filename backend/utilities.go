package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	goLog "github.com/withmandala/go-log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"testing"
)

var (
	goLogger       = goLog.New(os.Stderr).WithColor()
	mode           = "test"
	db             = InitDB()
	ErrTodoReqBody = "invalid request body, please include a text field with non-zero length"
	ErrUserReqBody = "invalid request body, must have a uname and pass field"
	ErrInvalidID   = "invalid id"
	ErrInternal    = "please try again later"
	ErrAuth        = "could not authenticate user"
	uid            = 0
)

// initialize the testing environment for subsequent tests
func initTestEnvironment() {
	TruncateTable(&Todo{})
	TruncateTable(&User{})
	// create the user
	user := User{Uname: "adnan", Pass: "badshah"}
	db.Create(&user)
	uid = user.ID
	// create the secret file
	err := ioutil.WriteFile("/tmp/secret.txt", []byte(strconv.Itoa(user.ID)), 0644)
	assertTestError(err)
}

// clean the testing environment
func cleanTestEnvironment() {
	TruncateTable(&User{})
	TruncateTable(&Todo{})
	// remove secret file
	_ = os.Remove("/tmp/secret.txt")
}

func assertRandomErr(t *testing.T, err interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf("Didn't expect an error but got \n%v", err)
	}
}
func assertStatusCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("wanted %#v status but got %#v", want, got)
	}
}

func TruncateTable(t interface{}) {
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(t)
}

func ExtractID(r *http.Request) string {
	var id string

	if mode == "prod" {
		id = mux.Vars(r)["id"]
	} else {
		re := regexp.MustCompile(`/(todos|users)/(.*)`)
		id = string(re.FindSubmatch([]byte(r.URL.Path))[2])
	}

	return id
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

func assertTestError(err error) {
	if err != nil {
		panic(err)
	}
}

func assertServerError(err error, w http.ResponseWriter) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(ErrInternal))
		return false
	}
	return true
}

// CreateTodoReq accepts an optional request body and sends a POST request to /todos
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

func unmarshalAndAssert(t *testing.T, res *httptest.ResponseRecorder) User {
	var decodedResBody User
	assertRandomErr(t, json.Unmarshal(res.Body.Bytes(), &decodedResBody))
	assertStatusCode(t, res.Result().StatusCode, http.StatusOK)

	return decodedResBody
}
