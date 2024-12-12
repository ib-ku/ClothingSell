package main

import (
	"log"
	"net/http"
	"store/controller"
)

func handleRequests() {
	http.HandleFunc("/", controller.HomePage)
	http.HandleFunc("/allProducts", controller.AllProducts)
	http.HandleFunc("/setProduct", controller.SetProduct)
	http.HandleFunc("/allUsers", controller.AllUsers)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
