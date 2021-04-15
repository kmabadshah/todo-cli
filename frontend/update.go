package frontend

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
)

func init() {
	var id, data string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a todo with id and data",
		Run: func(cmd *cobra.Command, args []string) {
			method := http.MethodPut
			url := "http://localhost:8080/todos/" + id
			err := MakeRequest(method, url, []byte(data))

			if err != nil {
				fmt.Println(err)
			}
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "specify the id of the todo")
	cmd.Flags().StringVar(&data, "data", "", "specify the todo data to update")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		fmt.Println(err)
		return
	} else if err := cmd.MarkFlagRequired("data"); err != nil {
		fmt.Println(err)
		return
	}

	rootCmd.AddCommand(cmd)
}
