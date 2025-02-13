package middleware

import (
	"net/http"
	"strings"
)

type spaWriter struct {
	http.ResponseWriter
	notFound bool
}

func (w *spaWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusNotFound {
		w.notFound = true
	} else {
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *spaWriter) Write(p []byte) (int, error) {
	if w.notFound {
		return len(p), nil
	}

	return w.ResponseWriter.Write(p)
}

func NewSPA(fallback http.Handler) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := &spaWriter{ResponseWriter: w}
			next.ServeHTTP(writer, r)

			if writer.notFound {
				fallback.ServeHTTP(w, r)
			}
		})
	}
}

func ServeFileContents(file string, files http.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Restrict only to instances where the browser is looking for an HTML file
		if !strings.Contains(r.Header.Get("Accept"), "text/html") {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Open the file and return its contents using http.ServeContent
		index, err := files.Open(file)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		fi, err := index.Stat()
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), index)
	}
}
