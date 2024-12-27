package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestClient_GetString(t *testing.T) {
	tests := []struct {
		name           string
		useBearerToken bool
		bearerToken    string
		timeout        int
		retryCount     int
		url            string
		expectedBody   string
		expectedError  bool
	}{
		{
			name:           "Successful GET request",
			useBearerToken: false,
			timeout:        5,
			retryCount:     1,
			url:            "/test",
			expectedBody:   "Hello, World!",
			expectedError:  false,
		},
		{
			name:           "GET request with Bearer Token",
			useBearerToken: true,
			bearerToken:    "testtoken",
			timeout:        5,
			retryCount:     1,
			url:            "/test",
			expectedBody:   "Hello, World!",
			expectedError:  false,
		},
		{
			name:           "GET request with timeout",
			useBearerToken: false,
			timeout:        1,
			retryCount:     1,
			url:            "/timeout",
			expectedBody:   "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.useBearerToken {
					if r.Header.Get("Authorization") != "Bearer "+tt.bearerToken {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
				if r.URL.Path == "/timeout" {
					time.Sleep(2 * time.Second)
				} else {
					w.Write([]byte(tt.expectedBody))
				}
			}))
			defer server.Close()

			client := &Client{
				UseBearerToken: tt.useBearerToken,
				BearerToken:    tt.bearerToken,
				Timeout:        tt.timeout,
				RetryCount:     tt.retryCount,
			}

			body, err := client.GetString(server.URL + tt.url)
			if (err != nil) != tt.expectedError {
				t.Errorf("Client.GetString() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if body != tt.expectedBody {
				t.Errorf("Client.GetString() = %v, expected %v", body, tt.expectedBody)
			}
		})
	}
}

func TestClient_GetFile(t *testing.T) {
	tests := []struct {
		name           string
		useBearerToken bool
		bearerToken    string
		timeout        int
		retryCount     int
		url            string
		expectedBody   string
		expectedError  bool
	}{
		{
			name:           "Successful GET request",
			useBearerToken: false,
			timeout:        5,
			retryCount:     1,
			url:            "/test",
			expectedBody:   "Hello, World!",
			expectedError:  false,
		},
		{
			name:           "GET request with Bearer Token",
			useBearerToken: true,
			bearerToken:    "testtoken",
			timeout:        5,
			retryCount:     1,
			url:            "/test",
			expectedBody:   "Hello, World!",
			expectedError:  false,
		},
		{
			name:           "GET request with timeout",
			useBearerToken: false,
			timeout:        1,
			retryCount:     1,
			url:            "/timeout",
			expectedBody:   "",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.useBearerToken {
					if r.Header.Get("Authorization") != "Bearer "+tt.bearerToken {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
				if r.URL.Path == "/timeout" {
					time.Sleep(2 * time.Second)
				} else {
					w.Write([]byte(tt.expectedBody))
				}
			}))
			defer server.Close()

			client := &Client{
				UseBearerToken: tt.useBearerToken,
				BearerToken:    tt.bearerToken,
				Timeout:        tt.timeout,
				RetryCount:     tt.retryCount,
			}

			filePath, err := client.GetFile(server.URL + tt.url)
			if (err != nil) != tt.expectedError {
				t.Errorf("Client.GetFile() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if tt.expectedError {
				return
			}
			defer os.Remove(filePath)

			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Errorf("Failed to read file: %v", err)
				return
			}
			if string(data) != tt.expectedBody {
				t.Errorf("Client.GetFile() = %v, expected %v", string(data), tt.expectedBody)
			}
		})
	}
}

func TestClient_PostJson(t *testing.T) {
	tests := []struct {
		name           string
		useBearerToken bool
		bearerToken    string
		timeout        int
		retryCount     int
		url            string
		json           string
		expectedBody   string
		expectedError  bool
	}{
		{
			name:           "Successful POST request",
			useBearerToken: false,
			timeout:        5,
			retryCount:     1,
			url:            "/test",
			json:           `{"message": "Hello, World!"}`,
			expectedBody:   "Success",
			expectedError:  false,
		},
		{
			name:           "POST request with Bearer Token",
			useBearerToken: true,
			bearerToken:    "testtoken",
			timeout:        5,
			retryCount:     1,
			url:            "/test",
			json:           `{"message": "Hello, World!"}`,
			expectedBody:   "Success",
			expectedError:  false,
		},
		{
			name:           "POST request with timeout",
			useBearerToken: false,
			timeout:        1,
			retryCount:     1,
			url:            "/timeout",
			json:           `{"message": "Hello, World!"}`,
			expectedBody:   "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.useBearerToken {
					if r.Header.Get("Authorization") != "Bearer "+tt.bearerToken {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
				if r.URL.Path == "/timeout" {
					time.Sleep(2 * time.Second)
				} else {
					w.Write([]byte(tt.expectedBody))
				}
			}))
			defer server.Close()

			client := &Client{
				UseBearerToken: tt.useBearerToken,
				BearerToken:    tt.bearerToken,
				Timeout:        tt.timeout,
				RetryCount:     tt.retryCount,
			}

			body, err := client.PostJson(server.URL+tt.url, tt.json)
			if (err != nil) != tt.expectedError {
				t.Errorf("Client.PostJson() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if body != tt.expectedBody {
				t.Errorf("Client.PostJson() = %v, expected %v", body, tt.expectedBody)
			}
		})
	}
}

func TestClient_Post(t *testing.T) {
	tests := []struct {
		name           string
		useBearerToken bool
		bearerToken    string
		timeout        int
		retryCount     int
		url            string
		data           []byte
		contentType    string
		expectedBody   string
		expectedError  bool
	}{
		{
			name:           "Successful POST request",
			useBearerToken: false,
			timeout:        5,
			retryCount:     1,
			url:            "/test",
			data:           []byte("message=Hello, World!"),
			contentType:    "application/x-www-form-urlencoded",
			expectedBody:   "Success",
			expectedError:  false,
		},
		{
			name:           "POST request with Bearer Token",
			useBearerToken: true,
			bearerToken:    "testtoken",
			timeout:        5,
			retryCount:     1,
			url:            "/test",
			data:           []byte("message=Hello, World!"),
			contentType:    "application/x-www-form-urlencoded",
			expectedBody:   "Success",
			expectedError:  false,
		},
		{
			name:           "POST request with timeout",
			useBearerToken: false,
			timeout:        1,
			retryCount:     1,
			url:            "/timeout",
			data:           []byte("message=Hello, World!"),
			contentType:    "application/x-www-form-urlencoded",
			expectedBody:   "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.useBearerToken {
					if r.Header.Get("Authorization") != "Bearer "+tt.bearerToken {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
				if r.URL.Path == "/timeout" {
					time.Sleep(2 * time.Second)
				} else {
					w.Write([]byte(tt.expectedBody))
				}
			}))
			defer server.Close()

			client := &Client{
				UseBearerToken: tt.useBearerToken,
				BearerToken:    tt.bearerToken,
				Timeout:        tt.timeout,
				RetryCount:     tt.retryCount,
			}

			body, err := client.Post(server.URL+tt.url, tt.data, tt.contentType)
			if (err != nil) != tt.expectedError {
				t.Errorf("Client.Post() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if string(body) != tt.expectedBody {
				t.Errorf("Client.Post() = %v, expected %v", body, tt.expectedBody)
			}
		})
	}
}
