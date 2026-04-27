package sec

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSearchAndFetchAllCapsParsedFilings(t *testing.T) {
	c := NewClient("test@example.com")
	c.limiter = nil
	c.sleep = func(context.Context, time.Duration) error { return nil }
	c.HTTP = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(req.URL.Host, "efts.sec.gov"):
			body := `{
				"hits": {
					"total": {"value": 2},
					"hits": [
						{"_id":"0000000001-26-000001:primary_doc.xml","_source":{"ciks":["0000000001"],"form":"D","file_date":"2026-01-01","adsh":"0000000001-26-000001"}},
						{"_id":"0000000002-26-000001:primary_doc.xml","_source":{"ciks":["0000000002"],"form":"D","file_date":"2026-01-02","adsh":"0000000002-26-000001"}}
					]
				}
			}`
			return response(200, body), nil
		case strings.Contains(req.URL.Path, "000000000126000001"):
			return response(200, formDXML("First Issuer")), nil
		case strings.Contains(req.URL.Path, "000000000226000001"):
			return response(200, formDXML("Second Issuer")), nil
		default:
			return nil, fmt.Errorf("unexpected request: %s", req.URL.String())
		}
	})}

	got, err := c.SearchAndFetchAll(context.Background(), "issuer", 1)
	if err != nil {
		t.Fatalf("SearchAndFetchAll returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected exactly one parsed filing, got %d: %+v", len(got), got)
	}
	if got[0].EntityName != "First Issuer" {
		t.Fatalf("expected first parsed filing, got %+v", got[0])
	}
}

func TestGetRetries429WithRetryAfter(t *testing.T) {
	var requests int
	var sleeps []time.Duration
	c := NewClient("test@example.com")
	c.limiter = nil
	c.sleep = func(_ context.Context, d time.Duration) error {
		sleeps = append(sleeps, d)
		return nil
	}
	c.HTTP = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if requests == 1 {
			resp := response(429, "slow down")
			resp.Header.Set("Retry-After", "2")
			return resp, nil
		}
		return response(200, `{"ok":true}`), nil
	})}

	body, err := c.get(context.Background(), "https://www.sec.gov/test")
	if err != nil {
		t.Fatalf("get returned error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
	if len(sleeps) != 1 || sleeps[0] != 2*time.Second {
		t.Fatalf("expected one 2s retry-after sleep, got %+v", sleeps)
	}
}

func TestGetExhausted429ReturnsRateLimitError(t *testing.T) {
	c := NewClient("test@example.com")
	c.limiter = nil
	c.sleep = func(context.Context, time.Duration) error { return nil }
	c.HTTP = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return response(429, "too many requests"), nil
	})}

	_, err := c.get(context.Background(), "https://www.sec.gov/test")
	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
	if !strings.Contains(rateErr.Error(), "HTTP 429") {
		t.Fatalf("rate limit error should mention HTTP 429: %v", rateErr)
	}
}

func TestSearchAndFetchAllPropagatesRateLimitFromFilingFetch(t *testing.T) {
	c := NewClient("test@example.com")
	c.limiter = nil
	c.sleep = func(context.Context, time.Duration) error { return nil }
	c.HTTP = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Host, "efts.sec.gov") {
			body := `{
				"hits": {
					"total": {"value": 1},
					"hits": [
						{"_id":"0000000001-26-000001:primary_doc.xml","_source":{"ciks":["0000000001"],"form":"D","file_date":"2026-01-01","adsh":"0000000001-26-000001"}}
					]
				}
			}`
			return response(200, body), nil
		}
		return response(429, "too many requests"), nil
	})}

	_, err := c.SearchAndFetchAll(context.Background(), "issuer", 1)
	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       ioNopCloser{strings.NewReader(body)},
		Header:     make(http.Header),
	}
}

type ioNopCloser struct {
	*strings.Reader
}

func (c ioNopCloser) Close() error { return nil }

func formDXML(entity string) string {
	return fmt.Sprintf(`<edgarSubmission>
		<primaryIssuer>
			<cik>0000000001</cik>
			<entityName>%s</entityName>
			<entityType>Corporation</entityType>
			<jurisdictionOfInc>DELAWARE</jurisdictionOfInc>
			<yearOfInc><value>2026</value></yearOfInc>
		</primaryIssuer>
		<offeringData>
			<industryGroup><industryGroupType>Technology</industryGroupType></industryGroup>
			<offeringSalesAmounts>
				<totalOfferingAmount>1000000</totalOfferingAmount>
				<totalAmountSold>500000</totalAmountSold>
			</offeringSalesAmounts>
			<federalExemptionsExclusions><item>06B</item></federalExemptionsExclusions>
		</offeringData>
	</edgarSubmission>`, entity)
}
