package models

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role types
const (
	RoleAdmin     = "admin"
	RoleManager   = "manager"
	RoleUser      = "user"
	RoleAccountant = "accountant"
	RoleShopManager = "shop_manager"
)
