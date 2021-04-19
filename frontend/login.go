package frontend

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
)

func init() {
	data := ""
	cmd := &cobra.Command{
		Use:   "login",
		Short: "log in a user",
		Run: func(cmd *cobra.Command, args []string) {
			req, err := http.NewRequest(
				"GET",
				"http://localhost:8080/users",
				nil,
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
				fmt.Println("Successfully logged in")
			} else {
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(string(resBody))
			}
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "provide data for existing user")
	err := cmd.MarkFlagRequired("data")
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.AddCommand(cmd)
}
