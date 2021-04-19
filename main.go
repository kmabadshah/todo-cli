package main

import "todo-cli/frontend"

type User struct {
	Uname string `json:"uname"`
	Pass  string `json:"pass"`
}

func main() {
	//user := User{Uname: "adnan", Pass: "badshah"}
	//encodedUser, _ := json.Marshal(user)
	//var decodedUser User
	//_ = json.Unmarshal(encodedUser, &decodedUser)
	//fmt.Printf("%#v", decodedUser)

	frontend.Execute()
}
