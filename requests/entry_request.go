package request

type EntryRequest struct {
	Description  string `json:"description" bson:"description" validate:"required"`
	MainCategory string `json:"main_category" bson:"main_category" validate:"required"`
	SubCategory  string `json:"sub_category" bson:"sub_category" validate:"required"`
	Receipts     int    `json:"receipts" bson:"receipts" validate:"required"`
	Payment      int    `json:"payment" bson:"payment" validate:"required"`
	ProcessedBy  string `json:"processed_by" bson:"processed_by" validate:"required"`
}
