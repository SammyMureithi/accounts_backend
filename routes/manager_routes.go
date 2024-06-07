package routes

import (
	"accounts_backend/controllers"
	middleware "accounts_backend/middlewares"

	"net/http"

	"github.com/gorilla/mux"
)

// UserRoutes function to initialize user routes
func ManagerRoutes(router *mux.Router) {
	//Manager Routes 
	getUnconfirmedEntryRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetUnconfirmedEntries), []string{"Manager"})
    router.Handle("/manager/unconfirmed_entry", getUnconfirmedEntryRoute).Methods("GET")
	getAllEntriesRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetAllEntries), []string{"Manager"})
    router.Handle("/manager/entry", getAllEntriesRoute).Methods("GET")
	confirmedEntryRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.ApproveEntry), []string{"Manager"})
    router.Handle("/manager/entry/{entryId}", confirmedEntryRoute).Methods("PUT")
	requestEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.RequestEntryChange), []string{"Manager"})
    router.Handle("/manager/entry/{entryId}", requestEntryChangeRoute).Methods("POST")
	makeEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.MakeChangetoEntry), []string{"Manager"})
    router.Handle("/manager/entry_change/{entryId}", makeEntryChangeRoute).Methods("PUT")

	router.HandleFunc("/manager/report/{username}", controllers.GenerateAccountantExcelReport).Methods("GET")
	
	router.HandleFunc("/manager/report", controllers.GenerateAllAccountantExcelReport).Methods("GET")

}
