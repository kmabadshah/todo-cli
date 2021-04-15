package frontend

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
)

func init() {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete a todo",
		Run: func(cmd *cobra.Command, args []string) {
			method := http.MethodDelete
			url := "http://localhost:8080/todos/" + id
			err := MakeRequest(method, url, nil)

			if err != nil {
				fmt.Println(err)
			}
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "id of the todo to delete")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		fmt.Println(err)
		return
	}

	rootCmd.AddCommand(cmd)
}
