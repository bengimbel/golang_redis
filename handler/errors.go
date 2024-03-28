package handler

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrorInternalServerError error = errors.New("Internal Server Error")
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Function to create new error instance
func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Render a basic http error response
func RenderInternalServerError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	errResponse := NewError(code, err.Error())
	result, _ := json.Marshal(errResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(result)
	return
}
