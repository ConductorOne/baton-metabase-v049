package client

import (
	"fmt"
)

type Database struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Engine      string `json:"engine"`
	CreatedAt   string `json:"created-at"`
}

type DatabaseAPIResponse struct {
	Data []*Database `json:"data"`
}

type DataAccessDetails struct {
	NativePermission string `json:"native"`
}

type GroupPermission struct {
	Data *DataAccessDetails `json:"data"`
}

type DBPermissionGraph struct {
	Groups map[string]map[string]*GroupPermission `json:"groups"`
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
