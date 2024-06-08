package routes

import (
	"accounts_backend/controllers"
	middleware "accounts_backend/middlewares"

	"net/http"

	"github.com/gorilla/mux"
)

func AccountantRoutes(router *mux.Router) {

	//Accountant Routes 
	createEntryRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.CreateAnEntry), []string{"Accountant"})
    router.Handle("/accountant/entry", createEntryRoute).Methods("POST")

	getRejectedEntriesRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetRejectedEntries), []string{"Accountant"})
    router.Handle("/accountant/reject_entry/{accountantId}", getRejectedEntriesRoute).Methods("GET")

	updateRejectedEntryRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.UpdateRejectedEntries), []string{"Accountant"})
    router.Handle("/accountant/reject_entry/{entryId}", updateRejectedEntryRoute).Methods("PUT")

	getAccountantEntriesRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetAllMyEntries), []string{"Accountant"})
    router.Handle("/entries/{accountantId}", getAccountantEntriesRoute).Methods("GET")
	
	router.HandleFunc("/report/{accountantId}", controllers.GenerateExcelReport).Methods("GET")

	

	
}
