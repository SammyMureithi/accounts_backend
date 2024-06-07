package routes

import (
	"accounts_backend/controllers"
	middleware "accounts_backend/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

// UserRoutes function to initialize user routes
func UserRoutes(router *mux.Router) {
//Admin alone is allowed to sign up users
	approveEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.SignUp), []string{"Admin"})
    router.Handle("/auth/signup", approveEntryChangeRoute).Methods("POST")
  
     router.HandleFunc("/auth/signin", controllers.Login).Methods("POST")
}
