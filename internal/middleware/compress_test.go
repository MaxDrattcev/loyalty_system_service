package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompress_RequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("gzipped request body is decompressed", func(t *testing.T) {
		router := gin.New()
		router.Use(Compress())
		router.POST("/echo", func(c *gin.Context) {
			body, err := io.ReadAll(c.Request.Body)
			require.NoError(t, err)
			c.String(http.StatusOK, string(body))
		})
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, err := gz.Write([]byte(`{"hello":"world"}`))
		require.NoError(t, err)
		require.NoError(t, gz.Close())
		req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(buf.Bytes()))
		req.Header.Set("Content-Encoding", "gzip")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, `{"hello":"world"}`, w.Body.String())
	})
	t.Run("invalid gzip body -> 400", func(t *testing.T) {
		router := gin.New()
		router.Use(Compress())
		router.POST("/echo", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader([]byte("not-gzip")))
		req.Header.Set("Content-Encoding", "gzip")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCompress_ResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Compress())
	router.GET("/data", func(c *gin.Context) {
		c.String(http.StatusOK, "compressed-response")
	})
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	gr, err := gzip.NewReader(bytes.NewReader(w.Body.Bytes()))
	require.NoError(t, err)
	defer gr.Close()
	decoded, err := io.ReadAll(gr)
	require.NoError(t, err)
	require.Equal(t, "compressed-response", string(decoded))
}
