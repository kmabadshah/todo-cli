package frontend

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
)

var data string
var cmd = &cobra.Command{
	Use:   "create",
	Short: "create a todo",
	Run: func(cmd *cobra.Command, args []string) {
		// get the data
		// POST the data to /todos
		res, err := http.Post(
			"http://localhost:8080/todos",
			"application/json", bytes.NewBuffer([]byte(data)),
		)
		defer res.Body.Close()

		if err != nil {
			fmt.Println(err)
			return
		}

		// output the result
		if resBody, err := ioutil.ReadAll(res.Body); err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(string(resBody))
		}
	},
}

func init() {
	cmd.Flags().StringVar(&data, "data", "something", "data for creating a todo")
	rootCmd.AddCommand(cmd)
}
