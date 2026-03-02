package provider

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type MockHTTPClient struct {
	Response     *http.Response
	Err          error
	RequestsMade []*http.Request
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.RequestsMade = append(m.RequestsMade, req)
	return m.Response, m.Err
}

func TestKrakenProvider_Success(t *testing.T) {
	mockBody := `{"error":[],"result":{"XXBTZUSD":{"c":["65000.50","65000.50"]}}}`
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockBody)),
		},
	}

	p := NewKraken(mockClient, "http://dummy.com")
	price, err := p.FetchPrice(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if price != 65000.50 {
		t.Fatalf("expected price 65000.50, got %v", price)
	}
}

func TestCoinbaseProvider_Success(t *testing.T) {
	mockBody := `{"data":{"amount":"66000.75"}}`
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockBody)),
		},
	}

	p := NewCoinbase(mockClient, "http://dummy.com")
	price, err := p.FetchPrice(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if price != 66000.75 {
		t.Fatalf("expected price 66000.75, got %v", price)
	}
}

func TestCoinDeskProvider_Success(t *testing.T) {
	mockBody := `{"Data":{"BTC-USD":{"PRICE":67000.25}}}`
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockBody)),
		},
	}

	p := NewCoinDesk(mockClient, "http://dummy.com")
	price, err := p.FetchPrice(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if price != 67000.25 {
		t.Fatalf("expected price 67000.25, got %v", price)
	}
}

func TestProvider_APIError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
		},
	}

	p := NewKraken(mockClient, "http://dummy.com")
	_, err := p.FetchPrice(context.Background())

	if err == nil {
		t.Fatal("expected error on 500 status, got nil")
	}
}
