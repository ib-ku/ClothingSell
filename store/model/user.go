package model

type User struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	Username string `json:"username" bson:"username"`
}

type Users []User
