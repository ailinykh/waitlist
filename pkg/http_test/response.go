package httptest

type response struct {
	code        int
	contentType string
	body        []byte
}

func WithCode(code int) func(*response) {
	return func(resp *response) {
		resp.code = code
	}
}

func WithContentType(contentType string) func(*response) {
	return func(resp *response) {
		resp.contentType = contentType
	}
}

func WithBody(body []byte) func(*response) {
	return func(resp *response) {
		resp.body = body
	}
}
