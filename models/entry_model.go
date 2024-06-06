package models

import (
	"time"
)

// Entry struct to define an entry model
type Entry struct {
    ID             string    `json:"id"`
    Date           time.Time `json:"date" bson:"date" validate:"required"`
    Description    string    `json:"description" bson:"description" validate:"required"`
    MainCategory   string    `json:"main_category" bson:"main_category" validate:"required"`
    SubCategory    string    `json:"sub_category" bson:"sub_category" validate:"required"`
    Receipts       int       `json:"receipts" bson:"receipts" validate:"required"`
    Payment        int       `json:"payment" bson:"payment" validate:"required"`
    ProcessedBy    string    `json:"processed_by" bson:"processed_by" validate:"required"`
    ApprovalStatus *int      `json:"approval_status" bson:"approval_status"` 
    RejReason      *string   `json:"rej_reason" bson:"rej_reason"` 
    ApprovedBy     *string   `json:"approved_by" bson:"approved_by"`         
    ChangeStatus   *string      `json:"change_status" bson:"change_status"`       
    ChangeApprovalBy *string      `json:"change_approval_by" bson:"change_approval_by"` 
    ChangedBy      *string      `json:"changed_by" bson:"changed_by"`  
    CreatedAt      time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" bson:"updated_at"`
}
