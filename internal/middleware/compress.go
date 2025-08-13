/*
Package middleware provides HTTP middleware components for the application.

It includes:
- Response compression using gzip
- Request body decompression
- Content type aware compression
- Error handling for compression operations
*/
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"
	"strings"
)

// compressWriter wraps http.ResponseWriter to provide gzip compression
// for supported content types.
type compressWriter struct {
	w  http.ResponseWriter // Original response writer
	zw *gzip.Writer        // Gzip writer for compression
}

// Compression is middleware that handles request/response compression.
// It supports:
// - Compressing responses with gzip for clients that accept it
// - Decompressing gzip-encoded request bodies
// - Automatic handling of supported content types
//
// Supported content types: application/json, text/html
func Compression(h http.Handler) http.Handler {
	compressFn := func(w http.ResponseWriter, r *http.Request) {
		var err error
		var cr *compressReader
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportContentTypes := []string{"application/json", "text/html"}
		isSupportGzip := strings.Contains(acceptEncoding, "gzip")
		if isSupportGzip && slices.Contains(supportContentTypes, r.Header.Get("Content-Type")) {
			cw := newCompressWriter(w)
			ow = cw
			defer func(cw *compressWriter) {
				err = cw.Close()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}(cw)
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		isReceivedGzip := strings.Contains(contentEncoding, "gzip")
		if isReceivedGzip {
			cr, err = newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func(cr *compressReader) {
				err = cr.Close()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}(cr)
		}

		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(compressFn)
}

// newCompressWriter creates a new compressWriter instance.
// Parameters:
// - w: Original http.ResponseWriter to wrap
// Returns:
// - *compressWriter: Initialized compression writer
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header returns the header map from the original ResponseWriter.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write compresses and writes the data to the underlying connection.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader sends an HTTP response header with the provided status code.
// Sets Content-Encoding header for successful responses (status < 300).
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Set("Accept-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip writer and flushes any pending compressed data.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader wraps io.ReadCloser to provide gzip decompression
// for incoming request bodies.
type compressReader struct {
	r  io.ReadCloser // Original reader
	zr *gzip.Reader  // Gzip reader for decompression
}

// newCompressReader creates a new compressReader instance.
// Parameters:
// - r: Original io.ReadCloser to wrap
// Returns:
// - *compressReader: Initialized decompression reader
// - error: If gzip reader creation fails
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read decompresses and reads data from the underlying connection.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the original reader and gzip reader.
func (c compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
