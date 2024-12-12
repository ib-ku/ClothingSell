package controller

import (
	"fmt"
	"net/http"
	"store/model"
	"store/view"
)

var users = model.Users{
	model.User{Email: "ermek@gmail.com", Password: "Qwerty123!", Username: "surf"},
}

func AllUsers(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Endpoint Hit:All Users Hit")
	view.RenderUsers(w, users)

}
