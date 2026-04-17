package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseHelpers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"foo": "bar"}
		Success(w, data)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp Envelope
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "bar", resp.Data.(map[string]interface{})["foo"])
	})

	t.Run("Created", func(t *testing.T) {
		w := httptest.NewRecorder()
		Created(w, "item-1")
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		Error(w, http.StatusBadRequest, "bad request")
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp Envelope
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "bad request", resp.Message)
	})

	t.Run("ValidationError", func(t *testing.T) {
		w := httptest.NewRecorder()
		ValidationError(w, map[string]string{"field": "required"})
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		Unauthorized(w, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		Forbidden(w)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		NotFound(w, "User")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		w := httptest.NewRecorder()
		InternalError(w)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
