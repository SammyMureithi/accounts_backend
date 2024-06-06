package request

type ApproveEntryChangeRequest struct {
	ChangeStatus     string `json:"change_status" bson:"change_status" validate:"required,oneof=Approved Rejected"`
	ChangeApprovalBy string `json:"change_approval_by" bson:"change_approval_by" validate:"required"`
}