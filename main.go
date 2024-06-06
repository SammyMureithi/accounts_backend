package main

import (
	"accounts_backend/database"
	"accounts_backend/routes"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Create a new router
	router := mux.NewRouter()

	// Initialize user routes
	routes.UserRoutes(router)
	routes.AccountantRoutes(router)
	routes.ManagerRoutes(router)
	routes.AdminRoutes(router)
	

	// Start the server on port 8000 with the router
	fmt.Println("Server running on port 8000...")
	err := http.ListenAndServe(":8000", router)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
	client := database.DbInstance()
	println(client)

}


