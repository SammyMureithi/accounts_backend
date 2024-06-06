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
	getEntriesRequestingChanges:=middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetEntriesRequestingChange), []string{"Admin"})
    router.Handle("/admin/entry", getEntriesRequestingChanges).Methods("GET")

	router.HandleFunc("/admin/report/{username}", controllers.GenerateAccountantExcelReport).Methods("GET")



}
