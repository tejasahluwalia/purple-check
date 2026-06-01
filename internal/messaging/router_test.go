package messaging

import (
	"encoding/json"
	"testing"

	"purple-check/internal/models"
)

func decodeMessageEvent(t *testing.T, body string) models.MessageEvent {
	t.Helper()
	var event models.MessageEvent
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		t.Fatalf("decode message event: %v", err)
	}
	return event
}

func TestGetPayload(t *testing.T) {
	quickReply := decodeMessageEvent(t, `{"message":{"quick_reply":{"payload":"SEARCH"}}}`)
	if payload := getPayload(quickReply); payload != payloadSearch {
		t.Fatalf("quick reply payload = %q, want %q", payload, payloadSearch)
	}

	postback := decodeMessageEvent(t, `{"postback":{"payload":"RATE:purplecheck_org"}}`)
	if payload := getPayload(postback); payload != "RATE:purplecheck_org" {
		t.Fatalf("postback payload = %q", payload)
	}
}

func TestParseRatePayload(t *testing.T) {
	username, err := parseRatePayload("RATE:@PurpleCheck_Org")
	if err != nil {
		t.Fatalf("parseRatePayload returned error: %v", err)
	}
	if username != "purplecheck_org" {
		t.Fatalf("username = %q", username)
	}

	if _, err := parseRatePayload("RATE:..bad"); err == nil {
		t.Fatal("expected invalid username error")
	}
	if _, err := parseRatePayload("SEARCH"); err == nil {
		t.Fatal("expected invalid command error")
	}
}

func TestParseRoleAndDealStagePayload(t *testing.T) {
	if role, ok := parseRolePayload("ROLE:BUYER"); !ok || role != roleBuyer {
		t.Fatalf("role = %q, ok = %v", role, ok)
	}
	if _, ok := parseRolePayload("ROLE:ADMIN"); ok {
		t.Fatal("unexpected valid role")
	}

	if stage, ok := parseDealStagePayload("DEAL_STAGE:COMPLETE"); !ok || stage != dealStageComplete {
		t.Fatalf("deal stage = %q, ok = %v", stage, ok)
	}
	if _, ok := parseDealStagePayload("DEAL_STAGE:PENDING"); ok {
		t.Fatal("unexpected valid deal stage")
	}
}

func TestParseRatingPayload(t *testing.T) {
	rating, username, err := parseRatingPayload("RATING:POSITIVE:@PurpleCheck_Org")
	if err != nil {
		t.Fatalf("parseRatingPayload returned error: %v", err)
	}
	if rating != ratingPositive || username != "purplecheck_org" {
		t.Fatalf("rating = %q, username = %q", rating, username)
	}

	tests := []string{
		"RATING:UNKNOWN:purplecheck_org",
		"RATING:MIXED:purplecheck_org",
		"RATING:POSITIVE:..bad",
		"RATE:POSITIVE:purplecheck_org",
		"RATING:POSITIVE",
	}
	for _, payload := range tests {
		if _, _, err := parseRatingPayload(payload); err == nil {
			t.Fatalf("expected %q to be invalid", payload)
		}
	}
}
