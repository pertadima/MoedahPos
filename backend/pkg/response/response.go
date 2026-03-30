package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard API response shape.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// JSON writes a JSON response with the given status code and payload.
func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Success writes a successful 200 response.
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Envelope{Success: true, Data: data})
}

// Created writes a 201 response with data.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Envelope{Success: true, Data: data})
}

// Error writes an error response with the given status code and message.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, Envelope{Success: false, Message: message})
}

// ValidationError writes a 422 response with field-level validation errors.
func ValidationError(w http.ResponseWriter, errors interface{}) {
	JSON(w, http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

// Unauthorized writes a 401 response.
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	Error(w, http.StatusUnauthorized, message)
}

// Forbidden writes a 403 response.
func Forbidden(w http.ResponseWriter) {
	Error(w, http.StatusForbidden, "You do not have permission to perform this action")
}

// NotFound writes a 404 response.
func NotFound(w http.ResponseWriter, resource string) {
	Error(w, http.StatusNotFound, resource+" not found")
}

// InternalError writes a 500 response.
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "An internal error occurred")
}
