package http

import (
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{}, //nolint:exhaustruct
	}
}

type Response struct {
	Body       []byte
	StatusCode int
	Headers    map[string][]string
}

func (h *Client) Get(ctx context.Context, url string) (*Response, error) {
	return h.GetWithHeaders(ctx, url, nil)
}

func (h *Client) GetWithHeaders(ctx context.Context, url string, headers map[string][]string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for k, v := range headers {
		req.Header[k] = v
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Response{
		Body:       body,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}, nil
}

func (h *Client) PostWithHeaders(
	ctx context.Context,
	url string,
	headers map[string][]string,
	reader io.Reader,
) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for k, v := range headers {
		req.Header[k] = v
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Response{
		Body:       body,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}, nil
}
