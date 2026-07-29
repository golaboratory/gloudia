package pdf

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Convert_APIErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 4xx is NOT retried by net/httpclient, so Convert's status check runs immediately.
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid file"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	pdf, err := client.Convert(context.Background(), "test.xlsx", strings.NewReader("dummy"), nil)
	assert.Nil(t, pdf)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAPIError)
	assert.Contains(t, err.Error(), "400")
}

func TestClient_Convert_MultipartFields(t *testing.T) {
	type captured struct {
		landscape  string
		pageRanges string
		scale      []string
		filename   string
		fileBody   string
	}
	var cap captured

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(10<<20))
		cap.landscape = r.FormValue("landscape")
		cap.pageRanges = r.FormValue("nativePageRanges")
		cap.scale = r.MultipartForm.Value["scale"]
		f, hdr, err := r.FormFile("files")
		require.NoError(t, err)
		defer f.Close()
		cap.filename = hdr.Filename
		b, _ := io.ReadAll(f)
		cap.fileBody = string(b)
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF-1.4"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	opts := &ConvertOptions{Landscape: true, PageRanges: "1-5", Scale: 1.5}
	pdf, err := client.Convert(context.Background(), "report.xlsx", strings.NewReader("excel-bytes"), opts)
	require.NoError(t, err)
	defer pdf.Close()
	_, _ = io.ReadAll(pdf)

	assert.Equal(t, "true", cap.landscape)
	assert.Equal(t, "1-5", cap.pageRanges)
	require.Len(t, cap.scale, 1)
	assert.Equal(t, "1.50", cap.scale[0])
	assert.Equal(t, "report.xlsx", cap.filename)
	assert.Equal(t, "excel-bytes", cap.fileBody)
}

func TestClient_Convert_DefaultOptionsOmitFields(t *testing.T) {
	var form *multipart.Form
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(10<<20))
		form = r.MultipartForm
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("%PDF"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	pdf, err := client.Convert(context.Background(), "x.xlsx", strings.NewReader("d"), nil)
	require.NoError(t, err)
	defer pdf.Close()
	_, _ = io.ReadAll(pdf)

	_, hasLandscape := form.Value["landscape"]
	_, hasScale := form.Value["scale"]
	_, hasRanges := form.Value["nativePageRanges"]
	assert.False(t, hasLandscape, "landscape must be omitted under default options")
	assert.False(t, hasScale, "scale must be omitted under default options")
	assert.False(t, hasRanges, "nativePageRanges must be omitted under default options")
}
