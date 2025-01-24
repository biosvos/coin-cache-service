package http

const (
	StatusOK              = 200
	StatusCreated         = 201
	StatusTooManyRequests = 429
)

func IsOK(resp *Response) bool {
	return resp.StatusCode == StatusOK
}
