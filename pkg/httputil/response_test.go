package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", handler)
	return r
}

func TestSendSuccess(t *testing.T) {
	r := setupTestRouter(func(c *gin.Context) {
		SendSuccess(c, http.StatusOK, gin.H{"id": "123", "name": "Test"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SuccessResponse
	parseErr := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, parseErr)
	assert.NotNil(t, resp.Data)
}

func TestSendError(t *testing.T) {
	r := setupTestRouter(func(c *gin.Context) {
		SendError(c, http.StatusBadRequest, "invalid request")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	parseErr := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, parseErr)
	assert.Equal(t, "invalid request", resp.Errors.Message)
}

func TestSendSuccessWithMeta(t *testing.T) {
	r := setupTestRouter(func(c *gin.Context) {
		data := []string{"a", "b"}
		meta := gin.H{"total": 2}
		links := gin.H{"next": "/test?page=2"}
		SendSuccessWithMeta(c, http.StatusOK, data, meta, links)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	parseErr := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, parseErr)
	assert.NotNil(t, resp["data"])
	assert.NotNil(t, resp["meta"])
	assert.NotNil(t, resp["links"])
}
