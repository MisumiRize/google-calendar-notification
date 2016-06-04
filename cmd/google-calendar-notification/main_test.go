package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	"google.golang.org/api/calendar/v3"
)

func TestFormatEvents(t *testing.T) {
	s := formatEvents([]*calendar.Event{
		&calendar.Event{
			Summary: "foo",
			Start: &calendar.EventDateTime{
				DateTime: "2016-06-01 00:00:00",
			},
		},
		&calendar.Event{
			Summary: "bar",
			Start: &calendar.EventDateTime{
				Date: "2016-06-01",
			},
		},
	})
	if s != `これから一週間の予定です！
foo (2016-06-01 00:00:00)
bar (2016-06-01)
` {
		t.Errorf("Assertion failed (actual: %v)", s)
	}
}

func TestUpdateTwitter(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "200 OK")
			},
		),
	)
	defer ts.Close()

	transport := client.Transport
	u, _ := url.Parse(ts.URL)
	client.Transport = rewriteTransport{URL: u}
	defer func() {
		client.Transport = transport
	}()

	updateTwitter(map[string]string{"status": "test"})
}

type rewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}
