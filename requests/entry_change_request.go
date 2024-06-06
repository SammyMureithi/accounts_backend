package request

import "time"

type EntryChangeRequest struct {
	Date         time.Time `json:"date" bson:"date" validate:"required"`
	Description  string    `json:"description" bson:"description" validate:"required"`
	ManagerId    string    `json:"manager_id" bson:"manager_id" validate:"required"`
	MainCategory string    `json:"main_category" bson:"main_category" validate:"required"`
	SubCategory  string    `json:"sub_category" bson:"sub_category" validate:"required"`
	Receipts     int       `json:"receipts" bson:"receipts" validate:"required"`
	Payment      int       `json:"payment" bson:"payment" validate:"required"`
	ProcessedBy  string    `json:"processed_by" bson:"processed_by" validate:"required"`
}
