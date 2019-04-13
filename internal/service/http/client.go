package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/hummerd/gophercon/internal/api/http/middleware"

	"github.com/rs/zerolog"

	iou "github.com/hummerd/gostuff/ioutil"
	"github.com/pkg/errors"
)

type httpClient struct {
	serviceName string
	*http.Client
}

func (c *httpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	start := time.Now()

	lg := zerolog.Ctx(ctx)

	if c.serviceName != "" && lg != nil {
		logger := lg.With().Str("service", c.serviceName).Logger()
		lg = &logger
	}

	var reqBody []byte
	if req.Body != nil {
		r, err := iou.NewPrefixReader(req.Body, 1024)
		if err != nil && err != io.EOF {
			return nil, errors.Wrapf(err, "can not read request body from call to %s: ", req.URL.String())
		}
		req.Body = r
		reqBody = r.Prefix()
	}

	lg.Debug().
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Bytes("body", reqBody).
		Msg("call to http client")

	reqID := middleware.GetRequestID(ctx)
	req.Header.Set("X-Request-ID", reqID)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "can not DO call to %s: ", req.URL.String())
	}

	r, err := iou.NewPrefixReader(resp.Body, 1024)
	if err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "can not read response from call to %s: ", req.URL.String())
	}
	resp.Body = r

	lg.Debug().
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Bytes("body", r.Prefix()).
		Str("status", resp.Status).
		Dur("duration", time.Since(start)).
		Msg("call to http client response")

	return resp, nil
}

func (c *httpClient) DoJSON(ctx context.Context, req *http.Request, result interface{}) error {
	start := time.Now()

	lg := zerolog.Ctx(ctx)

	if c.serviceName != "" && lg != nil {
		logger := lg.With().Str("service", c.serviceName).Logger()
		lg = &logger
	}

	lg.Debug().
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Msg("call to http client")

	resp, err := c.Client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to do request to %s: ", req.URL)
	}
	defer drainReader(resp.Body, lg)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wrong status: %s when calling %s", resp.Status, req.URL)
	}

	r, err := iou.NewPrefixReader(resp.Body, 1024)
	if err != nil && err != io.EOF {
		return errors.Wrapf(err, "can not read response from call to %s: ", req.URL.String())
	}

	lg.Debug().
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Bytes("body", r.Prefix()).
		Str("status", resp.Status).
		Dur("duration", time.Since(start)).
		Msg("call to http client response")

	err = json.NewDecoder(r).Decode(result)
	return errors.Wrap(err, "failed to do request: ")
}

const defaultResponseLimit = 5 << (10 * 2) // 5MB

func newCustomClient(opts ...httpOpt) *httpClient {
	c := &httpClient{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			panic(err)
		}
	}

	return c
}

type httpOpt func(*httpClient) error

func withServicename(name string) httpOpt {
	return func(doer *httpClient) error {
		doer.serviceName = name
		return nil
	}
}

func withRootCA(path string) httpOpt {
	return func(doer *httpClient) error {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		certs, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to append %q to root_cas", path)
		}

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			return errors.New("no certs appended, using system certs only")
		}

		config := &tls.Config{
			RootCAs: rootCAs,
		}

		doer.Client.Transport = &http.Transport{TLSClientConfig: config}

		return nil
	}
}

func drainReader(cl io.ReadCloser, logger *zerolog.Logger) {
	_, err := io.Copy(ioutil.Discard, cl)
	if err != nil {
		logger.Error().Err(err).Msg("can not read request's body")
	}

	err = cl.Close()
	if err != nil {
		logger.Error().Err(err).Msg("can not close request's body")
	}
}

func intsToString(ns []int64) string {
	if len(ns) == 0 {
		return ""
	}

	b := []byte{}

	for _, n := range ns {
		b = strconv.AppendInt(b, n, 10)
		b = append(b, ',')
	}

	b = b[:len(b)-1]

	return string(b)
}
