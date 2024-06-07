package controllers

import (
	"accounts_backend/database"
	"accounts_backend/models"
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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateAnEntry(w http.ResponseWriter, r *http.Request) {
//let's first validate our request
var entryReq request.EntryRequest
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

newEntry := models.Entry{
	Description: entryReq.Description,
	MainCategory: entryReq.MainCategory,
	SubCategory: entryReq.SubCategory,
	Receipts: entryReq.Receipts,
	Payment: entryReq.Payment,
	ProcessedBy: entryReq.ProcessedBy,
	ApprovalStatus: nil,
	RejReason: nil,
	ApprovedBy: nil,
	ChangeStatus: nil,
	ChangeApprovalBy: nil,
	ChangedBy: nil,
	ID: uuid.NewString(), 
	CreatedAt: time.Now(), 
	UpdatedAt: time.Now(),
}


//lets first open the collection we need to insert the entry in the collection
collection:= database.OpenCollection(database.Client,"entry")
ctx,cancel :=context.WithTimeout(context.Background(),10*time.Second)
defer cancel()
//let's now insert the record
_,err=collection.InsertOne(ctx,newEntry)
if err != nil {

	json.NewEncoder(w).Encode(bson.M{"ok":false,"status":"failed","message":err.Error()})
}

w.Header().Set("Content-Type", "application/json")

json.NewEncoder(w).Encode(bson.M{"ok":true,"status":"success","message":"Record added successfully"})

}

func GetRejectedEntries(w http.ResponseWriter, r *http.Request) {
    // Get the MongoDB collection for entries
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users") // Assume users are in 'users' collection

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
	vars := mux.Vars(r)
    accountantId := vars["accountantId"]

    // Query to find documents where ApprovalStatus and ApprovedBy are null
    filter := bson.M{
        "approval_status": bson.M{"$eq": "Rejected"},
		"processed_by": bson.M{"$eq":accountantId},
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
        if approvedBy, ok := entry["approved_by"].(string); ok && approvedBy != "" {
            var user bson.M
            if err := usersCollection.FindOne(ctx, bson.M{"id": approvedBy}).Decode(&user); err == nil {
				delete(user,"password")
                entry["approved_by"] = user
            } else {
                entry["approved_by"] = "User details not found"
            }
        }
    }

    // Sending back the results as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bson.M{"ok": true, "status": "success", "entries": entries})
}

func UpdateRejectedEntries(w http.ResponseWriter, r *http.Request){
//let's first validate our request
var entryReq request.EntryRequest
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

 filter := bson.M{"id": entryId}

 // Check if the entry exists
 var doctor bson.M
 if err := collection.FindOne(ctx, filter).Decode(&doctor); err != nil {
	 if err == mongo.ErrNoDocuments {
		 http.Error(w, "No entry found with given ID", http.StatusNotFound)
		 return
	 }
	 http.Error(w, err.Error(), http.StatusInternalServerError)
	 return
 }

 update := bson.M{
	 "$set": bson.M{
		 "description": entryReq.Description,
		 "main_category": entryReq.MainCategory,
		 "payment": entryReq.Payment,
		 "sub_category": entryReq.SubCategory,
		 "receipts": entryReq.Receipts,
		 "approval_status": nil,
		 "rej_reason": nil,
		 "approved_by":nil,
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


func getColumnLetter(index int) string {
    return string(rune('A' + index))
}

func GenerateExcelReport(w http.ResponseWriter, r *http.Request) {
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    vars := mux.Vars(r)
    accountantId := vars["accountantId"]
    filter := bson.M{"processed_by": accountantId}
    cur, err := entriesCollection.Find(ctx, filter)
    if err != nil {
        http.Error(w, "Failed to fetch records: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer cur.Close(ctx)

    var entries []bson.M
    if err := cur.All(ctx, &entries); err != nil {
        http.Error(w, "Failed to parse records: "+err.Error(), http.StatusInternalServerError)
        return
    }

    f := excelize.NewFile()
    sheetName := "Report"
    index, err := f.NewSheet(sheetName)
    if err != nil {
        http.Error(w, "Failed to create a new sheet: "+err.Error(), http.StatusInternalServerError)
        return
    }
    f.SetActiveSheet(index)

    headers := []string{"ID", "Description", "Main Category", "Sub Category", "Payment", "Approval Status","Approved By","Date"}
    for i, header := range headers {
        col := getColumnLetter(i) + "1"
        f.SetCellValue(sheetName, col, header)
    }

	for i, entry := range entries {
        row := strconv.Itoa(i + 2)  // Start from the second row
        f.SetCellValue(sheetName, "A"+row, entry["id"])
        f.SetCellValue(sheetName, "B"+row, entry["description"])
        f.SetCellValue(sheetName, "C"+row, entry["main_category"])
        f.SetCellValue(sheetName, "D"+row, entry["sub_category"])
        f.SetCellValue(sheetName, "E"+row, entry["payment"])
		 if approvalStatus, ok := entry["approval_status"].(string); ok && approvalStatus != "" {
            f.SetCellValue(sheetName, "F"+row, approvalStatus)
        } else {
            f.SetCellValue(sheetName, "F"+row, "Pending")
        }
      
      
		// Handle date formatting
		


 // Handling approved_by field to fetch user details
 if approvedById, ok := entry["approved_by"].(string); ok && approvedById != "" {
	var approvedByUser bson.M
	if err := usersCollection.FindOne(ctx, bson.M{"id": approvedById}).Decode(&approvedByUser); err == nil {
		if name, ok := approvedByUser["name"].(string); ok {
			f.SetCellValue(sheetName, "G"+row, name)
		} else {
			f.SetCellValue(sheetName, "G"+row, "Name not available")
		}
	} else {
		f.SetCellValue(sheetName, "G"+row, "User details not found")
	}
} else {
	f.SetCellValue(sheetName, "G"+row, "_")
}
f.SetCellValue(sheetName, "H"+row, entry["created_at"])


    }


    buf, err := f.WriteToBuffer()
    if err != nil {
        http.Error(w, "Failed to create Excel file", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
    w.Write(buf.Bytes())
}


func GetAllMyEntries(w http.ResponseWriter, r *http.Request) {
    // Get the MongoDB collection for entries and users
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users")

    vars := mux.Vars(r)
    accountantId := vars["accountantId"]

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

    // Filter to fetch entries processed by the accountant
    filter := bson.M{"processed_by": accountantId}

    // Count total documents for pagination
    total, err := entriesCollection.CountDocuments(ctx, filter)
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
    cur, err := entriesCollection.Find(ctx, filter, findOptions)
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
        "ok":        true,
        "status":    "success",
        "entries":   entries,
        "pagination": pagination,
    })
}
