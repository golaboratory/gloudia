package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	UseBearerToken bool
	BearerToken    string
	Timeout        int
	RetryCount     int
	HttpClient     *http.Client
}

func (c *Client) Init() {
	var retry = 1
	if c.RetryCount > 1 {
		retry = c.RetryCount
	}

	c.HttpClient = &http.Client{
		Timeout: time.Duration(c.Timeout) * time.Second,
		Transport: &GloudiaRoundTripper{
			t:        http.DefaultTransport,
			maxRetry: retry,
			wait:     2 * time.Second,
		},
	}
}

func (c *Client) GetString(url string) (string, error) {
	arr, err := c.Get(url)
	if err != nil {
		return "", err
	}
	return string(arr), nil
}

func (c *Client) GetFile(url string) (string, error) {
	arr, err := c.Get(url)
	f, err := os.CreateTemp("", "gloudia.core.http.get")
	if err != nil {
		return "", err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			fmt.Printf("failed to remove file: %v\n", err)
		}
	}()

	_, err = f.Write(arr)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

func (c *Client) Get(url string) ([]byte, error) {

	c.Init()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []byte(""), err
	}

	if c.UseBearerToken {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.BearerToken))
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	byteArray, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte(""), err
	}

	return byteArray, nil
}

func (c *Client) PostJson(url string, json string) (string, error) {
	arr, err := c.Post(url, []byte(json), "application/json")
	if err != nil {
		return "", err
	}
	return string(arr), nil
}

func (c *Client) Post(url string, data []byte, contentType string) ([]byte, error) {
	c.Init()

	var ct = "application/x-www-form-urlencoded"
	if contentType != "" {
		ct = contentType
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return []byte(""), err
	}

	if c.UseBearerToken {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.BearerToken))
	}

	req.Header.Set("Content-Type", ct)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	byteArray, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte(""), err
	}

	return byteArray, nil
}
