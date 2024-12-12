package controller

import (
	"fmt"
	"net/http"
	"store/model"
	"store/view"
)

func AllUsers(w http.ResponseWriter, r *http.Request) {
	users := model.Users{
		model.User{Email: "ermek@gmail.com", Password: "Qwerty123!", Username: "surf"},
	}

	fmt.Println("Endpoint Hit:All Users Hit")
	view.RenderUsers(w, users)

}
