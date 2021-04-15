package backend

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
)

type Todo struct {
	Text string
	ID   int
}

type User struct {
	Uname string
	Pass  string
	ID    int `gorm:"primaryKey"`
}

var (
	mode         = "test"
	db           = InitDB()
	ErrReqBody   = "Invalid request body, please include a text field with non-zero length"
	ErrInvalidID = "Invalid id"
	ErrInternal  = "Please try again later"
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

func TruncateTable() {
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Todo{})
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
