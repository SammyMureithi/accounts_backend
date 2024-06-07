package controllers

import (
	"accounts_backend/database"
	request "accounts_backend/requests"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
func GetUnconfirmedEntries(w http.ResponseWriter, r *http.Request) {
    // Get the MongoDB collection for entries
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users") // Assume users are in 'users' collection

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Query to find documents where ApprovalStatus and ApprovedBy are null
    filter := bson.M{
        "approval_status": bson.M{"$eq": nil},
        "approved_by":     bson.M{"$eq": nil},
    }

    cur, err := entriesCollection.Find(ctx, filter)
    if err != nil {
        http.Error(w, "Failed to fetch records: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer cur.Close(ctx)

    var entries []bson.M
    if err = cur.All(ctx, &entries); err != nil {
        http.Error(w, "Failed to parse records: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Enrich entries with user details
    for _, entry := range entries {
        if processedBy, ok := entry["processed_by"].(string); ok && processedBy != "" {
            var user bson.M
            if err := usersCollection.FindOne(ctx, bson.M{"id": processedBy}).Decode(&user); err == nil {
				delete(user,"password")
                entry["processed_by"] = user
            } else {
                entry["processed_by"] = "User details not found"
            }
        }
    }

    // Sending back the results as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bson.M{"ok": true, "status": "success", "entries": entries})
}

func GetAllEntries(w http.ResponseWriter, r *http.Request) {
    // Get the MongoDB collection for entries and users
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    query := r.URL.Query()
    limitQuery := query.Get("limit")
    pageQuery := query.Get("page")

    limit := 10
    page := 1

    if l, err := strconv.Atoi(limitQuery); err == nil && l > 0 {
        limit = l
    }
    if p, err := strconv.Atoi(pageQuery); err == nil && p > 1 {
        page = p
    }

    skip := (page - 1) * limit

    // Count total documents for pagination
    total, err := entriesCollection.CountDocuments(ctx, bson.D{})
    if err != nil {
        http.Error(w, "Failed to count documents: "+err.Error(), http.StatusInternalServerError)
        return
    }

    totalPages := int(math.Ceil(float64(total) / float64(limit)))

    // Find options for pagination
    findOptions := options.Find()
    findOptions.SetLimit(int64(limit))
    findOptions.SetSkip(int64(skip))

    // Fetch entries with pagination
    cur, err := entriesCollection.Find(ctx, bson.D{}, findOptions)
    if err != nil {
        http.Error(w, "Failed to fetch records: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer cur.Close(ctx)

    var entries []bson.M
    if err = cur.All(ctx, &entries); err != nil {
        http.Error(w, "Failed to parse records: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Enrich entries with user details
    for _, entry := range entries {
        if processedBy, ok := entry["processed_by"].(string); ok && processedBy != "" {
            var user bson.M
            if err := usersCollection.FindOne(ctx, bson.M{"id": processedBy}).Decode(&user); err == nil {
                delete(user, "password")
                entry["processed_by"] = user
            } else {
                entry["processed_by"] = "User details not found"
            }
        }
		if approvedBy, ok := entry["approved_by"].(string); ok && approvedBy != "" {
            var user bson.M
            if err := usersCollection.FindOne(ctx, bson.M{"id": approvedBy}).Decode(&user); err == nil {
                delete(user, "password")
                entry["approved_by"] = user
            } else {
                entry["approved_by"] = "User details not found"
            }
        }
    }

    // Create and populate the pagination map
    pagination := map[string]interface{}{
        "current_page": page,
        "total_pages":  totalPages,
        "limit":        limit,
        "total_items":  total,
    }
	 // Add URLs to the pagination map before adding it to the result
	 if page < totalPages {
        nextURL := fmt.Sprintf("%s?limit=%d&page=%d", r.URL.Path, limit, page+1)
        pagination["next_page_url"] = nextURL
    }
    if page > 1 {
        prevURL := fmt.Sprintf("%s?limit=%d&page=%d", r.URL.Path, limit, page-1)
        pagination["previous_page_url"] = prevURL
    }

    // Sending back the results as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bson.M{
        "ok":     true,
        "status": "success",
        "entries": entries,
        "pagination": pagination,
    })
}

func ApproveEntry(w http.ResponseWriter, r *http.Request) {
	var entryReq request.ApproveEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&entryReq); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := request.ValidateApproveEntryRequest(entryReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Access the MongoDB collection
    collection := database.OpenCollection(database.Client, "entry")
    vars := mux.Vars(r)
    entryId := vars["entryId"]

    // Create a filter to find the entry by ID
    filter := bson.M{"id": entryId}

    // Check if the doctor exists
    var entry bson.M
    if err := collection.FindOne(ctx, filter).Decode(&entry); err != nil {
        if err == mongo.ErrNoDocuments {
            http.Error(w, "No entry found with given ID", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    update := bson.M{
        "$set": bson.M{
            "approval_status": entryReq.ApprovalStatus,
            "approved_by": entryReq.ApprovedBy,
			"rej_reason":entryReq.RejReason,
            "updated_at": time.Now(),
        },
    }

    // Update the doctor document
    if _, err := collection.UpdateOne(ctx, filter, update); err != nil {
        http.Error(w, "Failed to update entry", http.StatusInternalServerError)
        return
    }

    // Return success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bson.M{"ok": true, "message": fmt.Sprintf("Record %s successfully", entryReq.ApprovalStatus),"status":"success"})

}
func RequestEntryChange(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Access the MongoDB collection
    collection := database.OpenCollection(database.Client, "entry")
    vars := mux.Vars(r)
    entryId := vars["entryId"]

    // Create a filter to find the entry by ID
    filter := bson.M{"id": entryId}

    // Check if the doctor exists
    var entry bson.M
    if err := collection.FindOne(ctx, filter).Decode(&entry); err != nil {
        if err == mongo.ErrNoDocuments {
            http.Error(w, "No entry found with given ID", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    update := bson.M{
        "$set": bson.M{
            "change_status": "Requesting",
            "updated_at": time.Now(),
        },
    }

    // Update the doctor document
    if _, err := collection.UpdateOne(ctx, filter, update); err != nil {
        http.Error(w, "Failed to update entry", http.StatusInternalServerError)
        return
    }

    // Return success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bson.M{"ok": true, "message": "Requesting change from Admin","status":"success"})

}


func MakeChangetoEntry(w http.ResponseWriter, r *http.Request){
	//let's first validate our request
	var entryReq request.EntryChangeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&entryReq); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	
	// Initialize the validator and validate the request data
	validate := validator.New()
	err := validate.Struct(entryReq)
	if err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			errMessages := request.CustomeErrorMessage(errs)
			http.Error(w, strings.Join(errMessages, ", "), http.StatusBadRequest)
			return
		}
		http.Error(w, "Validation failed", http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	 // Access the MongoDB collection
	 collection := database.OpenCollection(database.Client, "entry")
	 vars := mux.Vars(r)
	 entryId := vars["entryId"]
	
	 filter := bson.M{"id": entryId,"change_status": bson.M{"$eq":"Approved"}}
	
	 // Check if the entry exists
	 var doctor bson.M
	 if err := collection.FindOne(ctx, filter).Decode(&doctor); err != nil {
		 if err == mongo.ErrNoDocuments {
			 http.Error(w, "No entry found with given ID or Not approved for change", http.StatusNotFound)
			 return
		 }
		 http.Error(w, err.Error(), http.StatusInternalServerError)
		 return
	 }
	
	 update := bson.M{
		 "$set": bson.M{
			 "date": entryReq.Date,
			 "description": entryReq.Description,
			 "main_category": entryReq.MainCategory,
			 "payment": entryReq.Payment,
			 "sub_category": entryReq.SubCategory,
			 "receipts": entryReq.Receipts,
			 "approval_status": nil,
			 "rej_reason": nil,
			 "approved_by":nil,
			 "changed_by": entryReq.ManagerId,
			 "change_status":nil,
			 "updated_at": time.Now(),
		 },
	 }
	
	 // Update the doctor document
	 if _, err := collection.UpdateOne(ctx, filter, update); err != nil {
		 http.Error(w, "Failed to update doctor", http.StatusInternalServerError)
		 return
	 }
	
	 // Return success response
	 w.Header().Set("Content-Type", "application/json")
	 json.NewEncoder(w).Encode(bson.M{"ok": true, "message": "Updated done successfully","status":"success"})
	
	}