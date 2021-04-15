package frontend

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

func init() {
	var data string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a todo",
		Run: func(cmd *cobra.Command, args []string) {
			// POST the data to /todos
			method := http.MethodPost
			url := "http://localhost:8080/todos"
			err := MakeRequest(method, url, []byte(data))

			if err != nil {
				fmt.Println(err)
			}
		},
	}

	cmd.Flags().StringVar(&data, "data", "", `todo create --data '{"text": "hello world"}'`)
	err := cmd.MarkFlagRequired("data")
	if err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(cmd)
}
