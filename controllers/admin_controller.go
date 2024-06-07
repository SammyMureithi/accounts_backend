package controllers

import (
	"accounts_backend/database"
	request "accounts_backend/requests"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetEntriesRequestingChange(w http.ResponseWriter, r *http.Request) {
    // Get the MongoDB collection for entries
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users") 

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    filter := bson.M{
        "change_status": bson.M{"$eq": "Requesting"},
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
func ApproveEntryChange(w http.ResponseWriter, r *http.Request) {
	var entryReq request.ApproveEntryChangeRequest
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
            "change_status": entryReq.ChangeStatus,
            "change_approval_by": entryReq.ChangeApprovalBy,
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
    json.NewEncoder(w).Encode(bson.M{"ok": true, "message": "Record updated successfully","status":"success"})

}



func GenerateAccountantExcelReport(w http.ResponseWriter, r *http.Request) {
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    vars := mux.Vars(r)
    username := vars["username"]

    // Fetch user ID based on username
    var user bson.M
    if err := usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
        http.Error(w, "Failed to fetch user: "+err.Error(), http.StatusBadRequest)
        return
    }
    userId, ok := user["id"].(string)
    if !ok {
        http.Error(w, "User ID not found or invalid", http.StatusBadRequest)
        return
    }

    filter := bson.M{"processed_by": userId}
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
    sheetName := fmt.Sprintf("Report %s", user["name"])
    index, err := f.NewSheet(sheetName)
    if err != nil {
        http.Error(w, "Failed to create a new sheet: "+err.Error(), http.StatusInternalServerError)
        return
    }
    f.SetActiveSheet(index)

    headers := []string{"Reference", "Description", "Main Category", "Sub Category", "Payment", "Approval Status", "Processed by","Approved By","Date",}
    for i, header := range headers {
        col := getColumnLetter(i) + "1"
        f.SetCellValue(sheetName, col, header)
    }
    for i, entry := range entries {
        row := strconv.Itoa(i + 2)  
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
     // Handling processed_by field to fetch user details
 if approvedById, ok := entry["processed_by"].(string); ok && approvedById != "" {
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

 // Handling approved_by field to fetch user details
 if approvedById, ok := entry["approved_by"].(string); ok && approvedById != "" {
	var approvedByUser bson.M
	if err := usersCollection.FindOne(ctx, bson.M{"id": approvedById}).Decode(&approvedByUser); err == nil {
		if name, ok := approvedByUser["name"].(string); ok {
			f.SetCellValue(sheetName, "H"+row, name)
		} else {
			f.SetCellValue(sheetName, "H"+row, "Name not available")
		}
	} else {
		f.SetCellValue(sheetName, "H"+row, "User details not found")
	}
} else {
	f.SetCellValue(sheetName, "H"+row, "_")
}
f.SetCellValue(sheetName, "I"+row, entry["created_at"])

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


func GenerateAllAccountantExcelReport(w http.ResponseWriter, r *http.Request) {
    entriesCollection := database.OpenCollection(database.Client, "entry")
    usersCollection := database.OpenCollection(database.Client, "users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    filter := bson.M{}
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

    headers := []string{"Reference", "Description", "Main Category", "Sub Category", "Payment", "Approval Status","Processed By","Approved By","Date"}
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
     // Handling processed_by field to fetch user details
 if approvedById, ok := entry["processed_by"].(string); ok && approvedById != "" {
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

 // Handling approved_by field to fetch user details
 if approvedById, ok := entry["approved_by"].(string); ok && approvedById != "" {
	var approvedByUser bson.M
	if err := usersCollection.FindOne(ctx, bson.M{"id": approvedById}).Decode(&approvedByUser); err == nil {
		if name, ok := approvedByUser["name"].(string); ok {
			f.SetCellValue(sheetName, "H"+row, name)
		} else {
			f.SetCellValue(sheetName, "H"+row, "Name not available")
		}
	} else {
		f.SetCellValue(sheetName, "H"+row, "User details not found")
	}
} else {
	f.SetCellValue(sheetName, "H"+row, "_")
}
f.SetCellValue(sheetName, "I"+row, entry["created_at"])

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

