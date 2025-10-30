package client

import (
	"fmt"
	"time"
)

// User represents a Metabase user entity returned by the API.
type User struct {
	ID        int        `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	IsActive  bool       `json:"is_active"`
	LastLogin *time.Time `json:"last_login"`
}

// UsersQueryResponse models the paginated response for user listings in Metabase.
type UsersQueryResponse struct {
	Data   []*User `json:"data"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

type CreateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active"`
	Password  string `json:"password"`
}

// Membership represents the relationship between a user and a group in Metabase.
type Membership struct {
	MembershipID   int  `json:"membership_id"`
	GroupID        int  `json:"group_id"`
	IsGroupManager bool `json:"is_group_manager"`
	UserID         int  `json:"user_id"`
}

// Group represents a group entity in Metabase.
type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"member_count"`
}

type ErrorResponse struct {
	MessageText string `json:"message,omitempty"`
	Status      int    `json:"status,omitempty"`
}

func (e *ErrorResponse) Message() string {
	if e.MessageText != "" {
		return e.MessageText
	}
	return fmt.Sprintf("status code: %d", e.Status)
}

func (e *ErrorResponse) Error() string {
	return e.Message()
}
