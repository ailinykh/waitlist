package httptest

type response struct {
	code        int
	contentType string
	cookies     map[string]*string
	body        []byte
}

func WithCode(code int) func(*response) {
	return func(resp *response) {
		resp.code = code
	}
}

func WithCookie(key string, values ...string) func(*response) {
	return func(resp *response) {
		if len(values) == 1 {
			for _, v := range values {
				resp.cookies[key] = &v
			}
		} else {
			resp.cookies[key] = nil
		}
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
