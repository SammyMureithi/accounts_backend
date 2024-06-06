package models

import (
	"time"
)

// User struct to define a user model
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username" bson:"username" validate:"required,min=3,max=20"`
    Name      string    `json:"name" bson:"name" validate:"required"`
    Email     string    `json:"email" bson:"email" validate:"required,email"`
    Password  string    `json:"password" bson:"password" validate:"required,min=8"`
	Role      string    `json:"role" bson:"role" validate:"required,min=8"`
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
