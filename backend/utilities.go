package backend

import (
	"fmt"
	"github.com/gorilla/mux"
	goLog "github.com/withmandala/go-log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
	"regexp"
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
)

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
