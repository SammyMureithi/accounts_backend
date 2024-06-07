package routes

import (
	"accounts_backend/controllers"
	middleware "accounts_backend/middlewares"

	"net/http"

	"github.com/gorilla/mux"
)

func AdminRoutes(router *mux.Router) {
	//Admin Routes 
	approveEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.ApproveEntryChange), []string{"Admin"})
    router.Handle("/admin/entry/{entryId}", approveEntryChangeRoute).Methods("PUT")
	getAllEntriesRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetAllEntries), []string{"Admin"})
    router.Handle("/admin/entry", getAllEntriesRoute).Methods("GET")
	getEntriesRequestingChanges:=middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetEntriesRequestingChange), []string{"Admin"})
    router.Handle("/admin/entry", getEntriesRequestingChanges).Methods("GET")
	router.HandleFunc("/admin/report/{username}", controllers.GenerateAccountantExcelReport).Methods("GET")
	getUnconfirmedEntryRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetUnconfirmedEntries), []string{"Admin"})
    router.Handle("/admin/unconfirmed_entry", getUnconfirmedEntryRoute).Methods("GET")
	confirmedEntryRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.ApproveEntry), []string{"Admin"})
    router.Handle("/admin/entry/{entryId}", confirmedEntryRoute).Methods("PUT")
	requestEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.RequestEntryChange), []string{"Admin"})
    router.Handle("/admin/entry/{entryId}", requestEntryChangeRoute).Methods("POST")
	makeEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.MakeChangetoEntry), []string{"Admin"})
    router.Handle("/admin/entry_change/{entryId}", makeEntryChangeRoute).Methods("PUT")

	router.HandleFunc("/admin/report/{username}", controllers.GenerateAccountantExcelReport).Methods("GET")
	router.HandleFunc("/admin/report", controllers.GenerateAllAccountantExcelReport).Methods("GET")





}
