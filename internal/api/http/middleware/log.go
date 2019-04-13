package middleware

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hummerd/gostuff/ioutil"
	"github.com/rs/zerolog"
)

// middlewareTrace tracing all http requests
// exclude sensitive data
// trace all incoming requests - only in debug mode?
// trace errors always + add context
// exclude binary Content-Type (images, files, etc.)
func Log(logger *zerolog.Logger, excludePath ...string) func(http.Handler) http.Handler {

	wpool := sync.Pool{New: func() interface{} { return newCachedWriter(nil, 1024) }}
	rpool := sync.Pool{New: func() interface{} { r, _ := ioutil.NewPrefixReader(nil, 1024); return r }}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url := r.URL.String()

			for _, ex := range excludePath {
				if strings.HasPrefix(url, ex) {
					h.ServeHTTP(w, r)
					return
				}
			}

			start := time.Now()

			rid := GetRequestID(r.Context())

			l := logger.
				Level(zerolog.GlobalLevel()).
				With().
				Str("request_id", rid).
				Logger()

			lc := l.WithContext(r.Context())
			r = r.WithContext(lc)

			if l.Debug() != nil {
				// Use ioutil.PrefixReader to log request's body
				cr := rpool.Get().(*ioutil.PrefixReader)
				defer rpool.Put(cr)
				err := cr.Reset(r.Body)
				if err != nil && err != io.EOF {
					l.Error().
						Err(err).
						Str("url", url).
						Str("method", r.Method).
						Msg("request, failed to read body")
					http.Error(w, "can not read body", http.StatusInternalServerError)
					return
				}

				r.Body = cr

				l.Debug().
					Str("url", url).
					Str("method", r.Method).
					Bytes("body", cr.Prefix()).
					Msg("request")
			}

			// Use cachedWriter to log response body
			cw := wpool.Get().(*cachedWriter)
			defer wpool.Put(cw)
			cw.Reset(w)

			h.ServeHTTP(cw, r)

			respLog := l.Debug()
			if cw.statusCode != 0 &&
				(cw.statusCode < http.StatusOK || cw.statusCode >= http.StatusBadRequest) {
				respLog = l.Error()
			}

			respLog.
				Str("url", url).
				Str("method", r.Method).
				Bytes("body", cw.Prefix()).
				Int("status", cw.Status()).
				Dur("duration", time.Since(start)).
				Msg("response")
		})
	}
}

func newCachedWriter(w http.ResponseWriter, s int) *cachedWriter {
	return &cachedWriter{
		PrefixWriter: *ioutil.NewPrefixWriter(w, s),
		w:            w,
	}
}

type cachedWriter struct {
	ioutil.PrefixWriter
	w          http.ResponseWriter
	statusCode int
}

func (cw *cachedWriter) Status() int {
	return cw.statusCode
}

func (cw *cachedWriter) Reset(w http.ResponseWriter) {
	cw.PrefixWriter.Reset(w)
	cw.w = w
	cw.statusCode = 0
}

func (cw *cachedWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *cachedWriter) Write(data []byte) (int, error) {
	if cw.statusCode == 0 {
		cw.statusCode = http.StatusOK
	}

	return cw.PrefixWriter.Write(data)
}

func (cw *cachedWriter) WriteHeader(statusCode int) {
	if cw.statusCode == 0 {
		cw.statusCode = statusCode
	}
	cw.w.WriteHeader(statusCode)
}

func (cw *cachedWriter) Flush() {
	f, ok := cw.w.(http.Flusher)
	if ok && f != nil {
		f.Flush()
	}
}
