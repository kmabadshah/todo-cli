package frontend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"todo-cli/backend"
)

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "todo list app for the 90's",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var startServerCmd = &cobra.Command{
	Use:   "start server",
	Short: "start the backend server",
	Run: func(cmd *cobra.Command, args []string) {
		backend.StartServer()
	},
}

var cmd = &cobra.Command{
	Use:   "hello",
	Short: "hello",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a color argument")
		}
		//if myapp.IsValidColor(args[0]) {
		//	return nil
		//}
		fmt.Println(args)
		return nil
		//return fmt.Errorf("invalid color specified: %s", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, World!")
	},
}

func Execute() {
	rootCmd.AddCommand(startServerCmd)
	rootCmd.AddCommand(cmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func MakeRequest(method, url string, data []byte) error {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// decode
	var decodedResBody interface{}
	if err := json.Unmarshal(resBody, &decodedResBody); err != nil {
		fmt.Print(string(resBody))
		return nil
	}

	// if the response is an array, loop over
	// otherwise print directly
	sliceResBody, ok := decodedResBody.([]interface{})
	if ok {
		for _, v := range sliceResBody {
			fmt.Println(v)
		}
	} else {
		fmt.Println(decodedResBody)
	}

	return nil
}

func HandleError(err error) (b bool) {
	if err != nil {
		// notice that we're using 1, so it will actually log where
		// the error happened, 0 = this function, we don't want that.
		_, fn, line, _ := runtime.Caller(1)
		log.Printf("[error] %s:%d %v", fn, line, err)
		b = true
	}
	return
}
