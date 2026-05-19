package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"purple-check/internal/models"
)

type recordingRouter struct {
	events []models.MessageEvent
}

func decodeMessageEvent(t *testing.T, body string) models.MessageEvent {
	t.Helper()
	var event models.MessageEvent
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		t.Fatalf("decode message event: %v", err)
	}
	return event
}

func (r *recordingRouter) RouteMessage(ctx context.Context, event models.MessageEvent) {
	r.events = append(r.events, event)
}

func TestShouldRouteMessageEvent(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "text",
			body: `{"sender":{"id":"user-1"},"message":{"text":"@alice"}}`,
			want: true,
		},
		{
			name: "quick reply",
			body: `{"sender":{"id":"user-1"},"message":{"quick_reply":{"payload":"SEARCH"}}}`,
			want: true,
		},
		{
			name: "postback",
			body: `{"sender":{"id":"user-1"},"postback":{"payload":"SEARCH"}}`,
			want: true,
		},
		{
			name: "referral",
			body: `{"sender":{"id":"user-1"},"referral":{"ref":"alice"}}`,
			want: true,
		},
		{
			name: "echo",
			body: `{"sender":{"id":"user-1"},"message":{"text":"ignored","is_echo":true}}`,
			want: false,
		},
		{
			name: "deleted",
			body: `{"sender":{"id":"user-1"},"message":{"text":"ignored","is_deleted":true}}`,
			want: false,
		},
		{
			name: "unsupported",
			body: `{"sender":{"id":"user-1"},"message":{"is_unsupported":true}}`,
			want: false,
		},
		{
			name: "missing sender",
			body: `{"message":{"text":"ignored"}}`,
			want: false,
		},
		{
			name: "configured account sender",
			body: `{"sender":{"id":"test"},"message":{"text":"ignored"}}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := decodeMessageEvent(t, tt.body)
			if got := shouldRouteMessageEvent(event); got != tt.want {
				t.Fatalf("shouldRouteMessageEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstagramHandlerRoutesValidEventsOnly(t *testing.T) {
	router := &recordingRouter{}
	handler := NewInstagramHandler(router)
	body := `{
		"object":"instagram",
		"entry":[{
			"id":"entry-1",
			"time":1,
			"messaging":[
				{"sender":{"id":"user-1"},"message":{"text":"@alice"}},
				{"sender":{"id":"user-2"},"message":{"text":"ignored","is_echo":true}},
				{"sender":{"id":"user-3"},"postback":{"payload":"SEARCH"}},
				{"sender":{"id":"user-4"},"referral":{"ref":"bob"}}
			]
		}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/webhook/instagram", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(router.events) != 3 {
		t.Fatalf("routed events = %d, want 3", len(router.events))
	}
}

func TestInstagramHandlerRejectsInvalidJSON(t *testing.T) {
	router := &recordingRouter{}
	handler := NewInstagramHandler(router)
	req := httptest.NewRequest(http.MethodPost, "/webhook/instagram", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if len(router.events) != 0 {
		t.Fatalf("routed events = %d, want 0", len(router.events))
	}
}
