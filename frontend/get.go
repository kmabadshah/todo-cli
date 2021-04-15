package frontend

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
)

func init() {
	var id string
	cmd := &cobra.Command{
		Use:       "get",
		Short:     "get a todo",
		ValidArgs: []string{"all"},
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.OnlyValidArgs(cmd, args)
			if id == "" && err != nil {
				return errors.New("missing argument or flag")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			method := http.MethodGet
			url := "http://localhost:8080/todos"
			if id != "" {
				url += "/" + id
			}

			err := MakeRequest(method, url, nil)
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "get todo by id")
	rootCmd.AddCommand(cmd)
}
