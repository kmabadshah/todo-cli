package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"todo-cli/backend"
)

func init() {
	data := ""
	cmd := &cobra.Command{
		Use:   "login",
		Short: "log in a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := http.NewRequest(
				"GET",
				"http://localhost:8080/users",
				bytes.NewReader([]byte(data)),
			)
			if err != nil {
				log.Fatal(err)
			}
			client := http.Client{}
			res, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if res.StatusCode == 200 {
				// decode and login
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					return err
				}
				var user map[string]interface{}
				if json.Unmarshal(resBody, &user) != nil {
					return err
				}
				backend.LogIn(user)

				fmt.Println("Successfully logged in")
			} else {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					return err
				}
				fmt.Println(string(resBody))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "provide data for existing user")
	err := cmd.MarkFlagRequired("data")
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.AddCommand(cmd)
}
