package timeout

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPanicWithoutTimeout_IsRecovered(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTimeout_Handler_AbortsOnSlowHandler(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	mw := New(WithTimeout(50 * time.Millisecond))
	r.Use(mw.Handler())

	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/slow", nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusGatewayTimeout, w.Code)
}
