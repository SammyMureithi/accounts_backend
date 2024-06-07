package main

import (
	"accounts_backend/database"
	"accounts_backend/routes"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
    // Create a new router
    router := mux.NewRouter()

    // Initialize user routes
    routes.UserRoutes(router)
    routes.AccountantRoutes(router)
    routes.ManagerRoutes(router)
    routes.AdminRoutes(router)

    // Setup CORS
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"}, // Allow only this origin
        AllowCredentials: true,
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
    })

    // Use the CORS middleware
    handler := c.Handler(router)

    // Start the server on port 8000 with the router
    fmt.Println("Server running on port 8000...")
    err := http.ListenAndServe(":8000", handler)
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
    client := database.DbInstance()
    println(client)
}
