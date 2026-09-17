// Package httpzclient provides an HTTP client with filter middleware support,
// base URL resolution, and lifecycle error handling.
package httpzclient

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"time"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
)

var defaultHttpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

// Options configures a Client instance.
type Options struct {
	// BaseUrl is the default base URL for relative request paths.
	BaseUrl *urlz.Url
	// GetClient optionally supplies a custom *http.Client for a given context.
	GetClient func(ctx context.Context) *http.Client
}

func (o *Options) fix() {
}

// Client represents an HTTP client capable of dispatching requests through a chain of filters.
type Client struct {
	// Options holds the configuration options for the client.
	Options Options
	filters []Filter
	client  *http.Client
}

// New creates a new Client configured with the given Options.
func New(ctx context.Context, opts Options) *Client {
	opts.fix()
	ret := &Client{
		Options: opts,
		client:  defaultHttpClient,
	}
	if opts.GetClient != nil {
		ret.client = opts.GetClient(ctx)
	}
	return ret
}

// AddFilter appends a Filter middleware to the client's request execution chain.
func (c *Client) AddFilter(filter Filter) {
	c.filters = append(c.filters, filter)
}

// MustSuccess executes an HTTP request, ensuring a successful response (status code 200-299),
// and panics if an error occurs or the status code is outside the 2xx range.
func (c *Client) MustSuccess(ctx context.Context, req *httpzrequest.Req) *Resp {
	dfz := deferz.New(ctx)
	defer dfz.Close()
	ret := c.MustDo(ctx, req)
	dfz.AddCloserS(ret.Body)
	if ret.StatusCode < 200 || ret.StatusCode >= 300 {
		panic(fmt.Errorf("http client error: %d, head body: %s", ret.StatusCode, blobz.MustReadAll(req.Body, 1024).Ascii()))
	}
	dfz.Detach()
	return ret
}

// MustDo executes an HTTP request and panics if any error occurs.
func (c *Client) MustDo(ctx context.Context, req *httpzrequest.Req) *Resp {
	ret, err := c.Do(ctx, req)
	errorz.Check(err)
	return ret
}

// Do executes an HTTP request through the registered filter middleware pipeline.
func (c *Client) Do(ctx context.Context, req *httpzrequest.Req) (*Resp, error) {
	var h Handler = c.send
	for i := len(c.filters) - 1; i >= 0; i-- {
		filter := c.filters[i]
		next := h
		h = func(ctx context.Context, req *httpzrequest.Req) (*Resp, error) {
			return filter(ctx, req, next)
		}
	}
	return h(ctx, req)
}

func (c *Client) send(ctx context.Context, req *httpzrequest.Req) (*Resp, error) {
	dfz := deferz.New(ctx)
	defer dfz.Close()

	urlString, err := c.buildURL(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, urlString, req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header[k] = v
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	dfz.AddCloserS(httpResp.Body)
	resp := &Resp{}
	resp.fromHttpResponse(httpResp)

	dfz.Detach()
	return resp, nil
}

func (c *Client) buildURL(req *httpzrequest.Req) (string, error) {
	u, err := req.ResolveUrl(c.Options.BaseUrl)
	// u, err := url.Parse(c.Options.BaseUrl)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid base URL: must be an absolute URL: %s", c.Options.BaseUrl)
	}

	query := url.Values{}
	if u.Query != "" {
		nq, err := url.ParseQuery(u.Query)
		if err != nil {
			return "", fmt.Errorf("invalid query in URL: %s", u)
		}
		maps.Copy(query, nq)
	}
	if req.Query != nil {
		maps.Copy(query, req.Query)
	}

	// u.Path = path.Join(u.Path, req.Path.String())
	u.Query = query.Encode()

	return u.String(), nil
}
