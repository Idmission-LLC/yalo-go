package yalo

import (
	"encoding/json"
	"testing"
)

func TestNotificationResponseParseSuccess(t *testing.T) {
	resp := &Response{
		JSONData: `{"success":true,"id":"123","message_ids":["abc","def"]}`,
	}

	var result NotificationResponse
	if err := resp.ParseResponse(&result); err != nil {
		t.Fatalf("ParseResponse returned error: %v", err)
	}

	if !result.Success {
		t.Fatalf("expected success to be true, got false")
	}

	if result.ID != "123" {
		t.Fatalf("expected id to be '123', got '%s'", result.ID)
	}

	if len(result.MessageIDs) != 2 || result.MessageIDs[0] != "abc" || result.MessageIDs[1] != "def" {
		t.Fatalf("unexpected message IDs: %#v", result.MessageIDs)
	}

	if result.Reason != nil {
		t.Fatalf("expected reason to be nil for success response, got %#v", result.Reason)
	}
}

func TestNotificationResponseParseFailure(t *testing.T) {
	payload := map[string]interface{}{
		"success":     false,
		"id":          "",
		"message_ids": nil,
		"reason": map[string]interface{}{
			"description": "Template's parameter validation encountered errors",
			"error":       "validation_error",
			"details": []map[string]interface{}{
				{
					"phone":       "+15868504413",
					"type":        "invalid-button-param",
					"parameter":   "buttons.0.parameters.text",
					"description": "user `+15868504413` button's text parameter is missing or empty",
				},
			},
		},
	}

	failureJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal failure payload: %v", err)
	}

	resp := &Response{
		JSONData: string(failureJSON),
	}

	var result NotificationResponse
	if err := resp.ParseResponse(&result); err != nil {
		t.Fatalf("ParseResponse returned error: %v", err)
	}

	if result.Success {
		t.Fatalf("expected success to be false, got true")
	}

	if result.Reason == nil {
		t.Fatal("expected reason to be populated")
	}

	if got := result.Reason.Description; got != "Template's parameter validation encountered errors" {
		t.Fatalf("unexpected reason description: %q", got)
	}

	if got := result.Reason.Error; got != "validation_error" {
		t.Fatalf("unexpected reason error: %q", got)
	}

	if len(result.Reason.Details) != 1 {
		t.Fatalf("expected one detail entry, got %d", len(result.Reason.Details))
	}

	detail := result.Reason.Details[0]
	if detail.Phone != "+15868504413" {
		t.Fatalf("unexpected phone: %q", detail.Phone)
	}

	if detail.Parameter != "buttons.0.parameters.text" {
		t.Fatalf("unexpected parameter: %q", detail.Parameter)
	}

	if detail.Type != "invalid-button-param" {
		t.Fatalf("unexpected type: %q", detail.Type)
	}

	expectedDesc := "user `+15868504413` button's text parameter is missing or empty"
	if detail.Description != expectedDesc {
		t.Fatalf("unexpected detail description: %q", detail.Description)
	}
}
