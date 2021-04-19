package frontend

import (
	"github.com/spf13/cobra"
	"log"
)

func init() {
	data := ""
	cmd := &cobra.Command{
		Use:   "signup",
		Short: "signup for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := MakeRequest(
				"POST",
				"http://localhost:8080/users",
				[]byte(data),
			)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "provide data for creating user")
	err := cmd.MarkFlagRequired("data")
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.AddCommand(cmd)
}
