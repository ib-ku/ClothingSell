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

func HandleUserPostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var requestUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestUser)
	if err != nil {
		response := map[string]string{
			"status":  "fail",
			"message": "Invalid JSON format",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if requestUser.Email == "" || requestUser.Password == "" || requestUser.Username == "" {
		response := map[string]string{
			"status":  "fail",
			"message": "Invalid user data: Email, Password, and Username are required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	newUser := model.User{
		Email:    requestUser.Email,
		Password: requestUser.Password,
		Username: requestUser.Username,
	}
	SetUserAfterHandling(newUser)

	fmt.Println("Received User Information:", "Email:", requestUser.Email, "Password:", requestUser.Password, "Username:", requestUser.Username)

	response := map[string]string{
		"status":  "success",
		"message": "User data successfully received",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func SetUserAfterHandling(user model.User) {
	users = append(users, user)
}
