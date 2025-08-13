package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressionMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		contentType        string
		acceptEncoding     string
		contentEncoding    string
		requestBody        string
		expectedStatus     int
		expectCompressed   bool
		expectDecompressed bool
	}{
		{
			name:               "compress json response",
			contentType:        "application/json",
			acceptEncoding:     "gzip",
			expectedStatus:     http.StatusOK,
			expectCompressed:   true,
			expectDecompressed: false,
		},
		{
			name:               "compress html response",
			contentType:        "text/html",
			acceptEncoding:     "gzip",
			expectedStatus:     http.StatusOK,
			expectCompressed:   true,
			expectDecompressed: false,
		},
		{
			name:               "do not compress unsupported content type",
			contentType:        "text/plain",
			acceptEncoding:     "gzip",
			expectedStatus:     http.StatusOK,
			expectCompressed:   false,
			expectDecompressed: false,
		},
		{
			name:               "do not compress when client doesn't accept gzip",
			contentType:        "application/json",
			acceptEncoding:     "",
			expectedStatus:     http.StatusOK,
			expectCompressed:   false,
			expectDecompressed: false,
		},
		{
			name:               "decompress gzip request",
			contentType:        "application/json",
			contentEncoding:    "gzip",
			requestBody:        "test request body",
			expectedStatus:     http.StatusOK,
			expectCompressed:   false,
			expectDecompressed: true,
		},
		{
			name:               "error on invalid gzip request",
			contentType:        "application/json",
			contentEncoding:    "gzip",
			requestBody:        "invalid gzip data",
			expectedStatus:     http.StatusInternalServerError,
			expectCompressed:   false,
			expectDecompressed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectDecompressed {
					body, err := io.ReadAll(r.Body)
					require.NoError(t, err, "failed to read request body")
					assert.Equal(t, tt.requestBody, string(body), "unexpected request body")
				}

				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.expectedStatus)
				_, err := w.Write([]byte("test response"))
				require.NoError(t, err)
			})

			var body io.Reader
			if tt.contentEncoding == "gzip" && tt.requestBody != "" {
				if tt.requestBody == "invalid gzip data" {
					body = strings.NewReader(tt.requestBody)
				} else {
					var buf bytes.Buffer
					gz := gzip.NewWriter(&buf)
					_, err := gz.Write([]byte(tt.requestBody))
					require.NoError(t, err, "failed to compress test data")
					require.NoError(t, gz.Close(), "failed to close gzip writer")
					body = &buf
				}
			} else {
				body = strings.NewReader("")
			}

			req := httptest.NewRequest("GET", "https://example.com", body)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			req.Header.Set("Content-Encoding", tt.contentEncoding)
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()

			middleware := Compression(handler)
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "unexpected status code")

			if tt.expectCompressed {
				assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"), "expected Content-Encoding header")

				reader, err := gzip.NewReader(rr.Body)
				require.NoError(t, err, "failed to create gzip reader")
				defer func(reader *gzip.Reader) {
					err = reader.Close()
					if err != nil {
						require.Error(t, err, "failed to close gzip reader")
					}
				}(reader)

				_, err = io.ReadAll(reader)
				assert.NoError(t, err, "failed to decompress response")
			} else {
				assert.Empty(t, rr.Header().Get("Content-Encoding"), "unexpected Content-Encoding header")
			}
		})
	}
}

func TestCompressWriter(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectEncoding bool
	}{
		{
			name:           "success status sets encoding",
			statusCode:     http.StatusOK,
			expectEncoding: true,
		},
		{
			name:           "error status doesn't set encoding",
			statusCode:     http.StatusInternalServerError,
			expectEncoding: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			cw := newCompressWriter(rr)

			assert.NotNil(t, cw.Header(), "Header() returned nil")

			cw.WriteHeader(tt.statusCode)

			if tt.expectEncoding {
				assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"), "expected Content-Encoding header")
			} else {
				assert.Empty(t, rr.Header().Get("Content-Encoding"), "unexpected Content-Encoding header")
			}

			testData := []byte("test data")
			n, err := cw.Write(testData)
			assert.NoError(t, err, "Write() failed")
			assert.Equal(t, len(testData), n, "unexpected number of bytes written")

			assert.NoError(t, cw.Close(), "Close() failed")
		})
	}
}

func TestCompressReader(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	testData := "test data"
	_, err := gz.Write([]byte(testData))
	require.NoError(t, err, "failed to prepare test data")
	require.NoError(t, gz.Close(), "failed to close gzip writer")

	cr, err := newCompressReader(io.NopCloser(&buf))
	require.NoError(t, err, "newCompressReader failed")
	defer func(cr *compressReader) {
		err = cr.Close()
		if err != nil {
			require.Error(t, err, "failed to close compress reader")
		}
	}(cr)

	data, err := io.ReadAll(cr)
	assert.NoError(t, err, "Read() failed")
	assert.Equal(t, testData, string(data), "unexpected decompressed data")

	assert.NoError(t, cr.Close(), "Close() failed")
}

func TestCompressReaderInvalidData(t *testing.T) {
	_, err := newCompressReader(io.NopCloser(strings.NewReader("invalid gzip data")))
	assert.Error(t, err, "expected error for invalid gzip data")
}
