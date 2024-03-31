package errorPkg

import (
	"encoding/json"
	"net/http"
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

// Render http internal server error response
func RenderInternalServerError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	errResponse := NewError(code, err.Error())
	result, _ := json.Marshal(errResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(result)
	return
}

// Render a basic http error response
func RenderBadRequestError(w http.ResponseWriter, err error) {
	code := http.StatusBadRequest
	errResponse := NewError(code, err.Error())
	result, _ := json.Marshal(errResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(result)
	return
}
