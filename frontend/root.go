package frontend

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"todo-cli/backend"
)

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "todo list app for the 90's",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello nigger!!!")
	},
}

var startServerCmd = &cobra.Command{
	Use:   "start server",
	Short: "start the backend server",
	Run: func(cmd *cobra.Command, args []string) {
		backend.StartServer()
	},
}

func Execute() {
	rootCmd.AddCommand(startServerCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
