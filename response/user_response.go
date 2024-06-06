package response

import "accounts_backend/models"

type Response struct {
	OK     bool        `json:"ok"`
	Status string      `json:"status"`
	Message string     `json:"message"`
	User   models.User `json:"user"`
}
