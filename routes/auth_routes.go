package routes

import (
	"accounts_backend/controllers"

	"github.com/gorilla/mux"
)

// UserRoutes function to initialize user routes
func UserRoutes(router *mux.Router) {
  
     router.HandleFunc("/auth/signup", controllers.SignUp).Methods("POST")
     router.HandleFunc("/auth/signin", controllers.Login).Methods("POST")
}
