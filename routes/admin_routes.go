package routes

import (
	"accounts_backend/controllers"
	middleware "accounts_backend/middlewares"

	"net/http"

	"github.com/gorilla/mux"
)

func AdminRoutes(router *mux.Router) {
	//Admin Specific Routes 
	approveEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.ApproveEntryChange), []string{"Admin"})
    router.Handle("/admin/entry/{entryId}", approveEntryChangeRoute).Methods("PUT")

	//Admin Specific
	getEntriesRequestingChanges:=middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetEntriesRequestingChange), []string{"Admin"})
    router.Handle("/admin/entry", getEntriesRequestingChanges).Methods("GET")
	
	getUnconfirmedEntryRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.GetUnconfirmedEntries), []string{"Admin"})
    router.Handle("/admin/unconfirmed_entry", getUnconfirmedEntryRoute).Methods("GET")

	// Admin/Manager
	confirmedEntryRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRoles := r.Context().Value("userRoles").([]string)
		if middleware.IsAllowedAccess(userRoles, []string{"Admin", "Manager"}) {
			controllers.ApproveEntry(w, r)
		} else {
			http.Error(w, "Not Allowed", http.StatusForbidden)
		}
	}), []string{"Admin", "Manager", "Accountant"})
	router.Handle("/entry/{entryId}", confirmedEntryRoute).Methods("PUT")

	
	// Admin/Manager
	requestEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRoles := r.Context().Value("userRoles").([]string)
		if middleware.IsAllowedAccess(userRoles, []string{"Admin", "Manager"}) {
			controllers.RequestEntryChange(w, r)
		} else {
			http.Error(w, "Not Allowed", http.StatusForbidden)
		}
	}), []string{"Admin", "Manager", "Accountant"})
	router.Handle("/entry/{entryId}", requestEntryChangeRoute).Methods("POST")
	
	//Admin alone
	makeEntryChangeRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(controllers.MakeChangetoEntry), []string{"Admin"})
    router.Handle("/admin/entry_change/{entryId}", makeEntryChangeRoute).Methods("PUT")

	combineAccessUsernameReportRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRoles := r.Context().Value("userRoles").([]string)
		if middleware.IsAllowedAccess(userRoles, []string{"Admin", "Manager"}) {
			controllers.GenerateAccountantExcelReport(w, r)
		} else {
			http.Error(w, "Not Allowed", http.StatusForbidden)
		}
	}), []string{"Admin", "Manager", "Accountant"})
	router.Handle("/report/{username}", combineAccessUsernameReportRoute).Methods("GET")


	combineReportRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRoles := r.Context().Value("userRoles").([]string)
		if middleware.IsAllowedAccess(userRoles, []string{"Admin", "Manager"}) {
			controllers.GenerateAllAccountantExcelReport(w, r)
		} else {
			http.Error(w, "Not Allowed", http.StatusForbidden)
		}
	}), []string{"Admin", "Manager", "Accountant"})
	router.Handle("/report", combineReportRoute).Methods("GET")


	combinedAccessAllEntriesRoute := middleware.RoleBasedJWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRoles := r.Context().Value("userRoles").([]string)
		if middleware.IsAllowedAccess(userRoles, []string{"Admin", "Manager"}) {
			controllers.GetAllEntries(w, r)
		} else {
			http.Error(w, "Not Allowed", http.StatusForbidden)
		}
	}), []string{"Admin", "Manager", "Accountant"})
	router.Handle("/entries", combinedAccessAllEntriesRoute).Methods("GET")





}
