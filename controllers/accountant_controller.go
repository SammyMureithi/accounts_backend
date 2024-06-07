package controllers

import (
	"accounts_backend/database"
	"accounts_backend/models"
	request "accounts_backend/requests"
	"context"
	"encoding/json"
	"fmt"
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

func GenerateMyReport(w http.ResponseWriter, r *http.Request) {
    // Setup database collections
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users")

    // Create a context with a timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get accountantId from URL parameters
    vars := mux.Vars(r)
    accountantId := vars["accountantId"]

    // Define filter to fetch entries
    filter := bson.M{"processed_by": bson.M{"$eq": accountantId}}
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

    // Enrich entries with user details
    for i, entry := range entries {
        enrichUserDetail(ctx, usersCollection, "approved_by", entry, entries, i)
        enrichUserDetail(ctx, usersCollection, "change_approval_by", entry, entries, i)
    }

    // Send the results as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bson.M{"ok": true, "status": "success", "entries": entries})
}

func enrichUserDetail(ctx context.Context, usersCollection *mongo.Collection, fieldName string, entry bson.M, entries []bson.M, index int) {
    userID, ok := entry[fieldName].(string)
    if ok && userID != "" {
        var user bson.M
        // Assuming _id is of type ObjectId in MongoDB, if it's a string adjust accordingly.
        if err := usersCollection.FindOne(ctx, bson.M{"id": userID}).Decode(&user); err == nil {
            delete(user, "password") // Safely remove the password from the user details
            entry[fieldName+"_details"] = user
        } else {
            entry[fieldName+"_details"] = "User details not found"
        }
        entries[index] = entry // Update the entry in the slice with enriched details
    }
}

func getColumnLetter(index int) string {
    return string(rune('A' + index))
}

func GenerateExcelReport(w http.ResponseWriter, r *http.Request) {
    entriesCollection := database.OpenCollection(database.Client, "entry")
   // usersCollection := database.OpenCollection(database.Client, "users")
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

    headers := []string{"ID", "Description", "Main Category", "Sub Category", "Payment", "Approval Status", "Date","Approved By"}
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
        f.SetCellValue(sheetName, "F"+row, entry["approval_status"])
      
		// Handle date formatting
if dateStr, ok := entry["created_at"].(string); ok {
    if dateStr == "" {
        fmt.Println("Date string is empty")
        f.SetCellValue(sheetName, "G"+row, "No date provided")
    } else {
        parsedDate, err := time.Parse(time.RFC3339, dateStr) 
        if err == nil {
            formattedDate := parsedDate.Format("2006-01-02") 
            f.SetCellValue(sheetName, "G"+row, formattedDate)
        } else {
            fmt.Printf("Failed to parse date '%s': %v\n", dateStr, err)
            f.SetCellValue(sheetName, "G"+row, "Invalid date")
        }
    }
} else {
    fmt.Println("Date field is missing or not a string")
    f.SetCellValue(sheetName, "G"+row, "No date field")
}

// Handling approved_by_details
if details, ok := entry["approved_by_details"].(bson.M); ok {
    if name, ok := details["name"].(string); ok {
        f.SetCellValue(sheetName, "H"+row, name)
    } else {
        fmt.Printf("Type assertion for name failed, details['name']: %v\n", details["name"])
        f.SetCellValue(sheetName, "H"+row, "Name not available") 
    }
} else {
    fmt.Printf("Type assertion for approved_by_details failed, entry['approved_by_details']: %v\n", entry["approved_by_details"])
    f.SetCellValue(sheetName, "H"+row, "Details not available")  
}

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
