package controller

import (
	"encoding/json"
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

func SetUser(w http.ResponseWriter, r *http.Request) {
	var newUser model.User

	err := json.NewDecoder(r.Body).Decode(&newUser)

	if err != nil {
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	users = append(users, newUser)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}
