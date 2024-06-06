package request

import (
	"github.com/go-playground/validator/v10"
)

type ApproveEntryRequest struct {
    ApprovalStatus string `json:"approval_status" bson:"approval_status" validate:"required,oneof=Approved Rejected"`
    RejReason      string `json:"rej_reason" bson:"rej_reason" validate:"rejreason_required_if_rejected"`
    ApprovedBy     string `json:"approved_by" bson:"approved_by" validate:"required"`
}

// Setup the validator instance as a package level variable
var validate = validator.New()

func init() {
    // Register custom validation function
    validate.RegisterValidation("rejreason_required_if_rejected", rejreasonRequiredIfRejected)
}

// rejreasonRequiredIfRejected checks that RejReason is not empty when ApprovalStatus is "Reject"
func rejreasonRequiredIfRejected(fl validator.FieldLevel) bool {
    approvalStatusField, _, _ := fl.GetStructFieldOK()
    if approvalStatusField.String() == "Reject" {
        return fl.Field().String() != ""
    }
    return true
}

func ValidateApproveEntryRequest(req ApproveEntryRequest) error {
    // Validate the request using the registered rules
    return validate.Struct(req)
}
