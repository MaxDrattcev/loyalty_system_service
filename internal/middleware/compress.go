package middleware

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type gzipReadCloser struct {
	*gzip.Reader
	orig io.Closer
}

func (g *gzipReadCloser) Close() error {
	_ = g.Reader.Close()
	if g.orig != nil {
		return g.orig.Close()
	}
	return nil
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	gz *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.gz.Write(b)
}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.gz.Write([]byte(s))
}

func Compress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			originBody := c.Request.Body
			gz, err := gzip.NewReader(originBody)
			if err != nil {
				log.Error("failed to encoding request", zap.Error(err))
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Request.Body = &gzipReadCloser{Reader: gz, orig: originBody}
			defer func() { _ = c.Request.Body.Close() }()
		}
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Header("Content-Encoding", "gzip")
			gz := gzip.NewWriter(c.Writer)
			defer gz.Close()

			c.Writer = &gzipResponseWriter{
				ResponseWriter: c.Writer,
				gz:             gz,
			}
		}
		c.Next()
	}
}
