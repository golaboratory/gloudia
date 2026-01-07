package pdf

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Convert(t *testing.T) {
	// Mock Gotenberg Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL
		if r.URL.Path != "/forms/libreoffice/convert" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Verify Method
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify Multipart
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad multipart"))
			return
		}

		file, _, err := r.FormFile("files")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing files"))
			return
		}
		defer file.Close()

		// Verify Options
		if r.FormValue("landscape") != "true" {
			// This depends on test case options
		}

		// Success response (fake PDF)
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("%PDF-1.4 mock content"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	t.Run("success conversion", func(t *testing.T) {
		ctx := context.Background()
		src := strings.NewReader("dummy excel content")
		opts := &ConvertOptions{
			Landscape: true,
			Scale:     1.0,
		}

		pdf, err := client.Convert(ctx, "test.xlsx", src, opts)
		assert.NoError(t, err)
		defer pdf.Close()

		content, err := io.ReadAll(pdf)
		assert.NoError(t, err)
		assert.Equal(t, "%PDF-1.4 mock content", string(content))
	})

	t.Run("server error", func(t *testing.T) {
		// Mock error server
		errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		}))
		defer errServer.Close()

		clientErr := NewClient(errServer.URL)
		ctx := context.Background()
		src := strings.NewReader("dummy")

		pdf, err := clientErr.Convert(ctx, "test.xlsx", src, nil)
		assert.Error(t, err)
		assert.Nil(t, pdf)
		assert.Contains(t, err.Error(), "status 500")
	})
}
