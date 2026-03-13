package mcp

import (
	"encoding/json"
	"testing"
)

func TestRequest_Marshal(t *testing.T) {
	params, _ := json.Marshal(map[string]interface{}{"key": "value"})
	req := Request{
		JSONRPC: "2.0",
		Method:  "test_method",
		Params:  params,
		ID:      1,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}

	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %v", parsed["jsonrpc"])
	}
	if parsed["method"] != "test_method" {
		t.Errorf("Expected method test_method, got %v", parsed["method"])
	}
}

func TestResponse_Marshal(t *testing.T) {
	resp := Response{
		JSONRPC: "2.0",
		Result:  map[string]interface{}{"data": "test"},
		ID:      1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}

	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %v", parsed["jsonrpc"])
	}
	if parsed["id"] != float64(1) {
		t.Errorf("Expected id 1, got %v", parsed["id"])
	}
}

func TestError_Marshal(t *testing.T) {
	e := Error{
		Code:    -32600,
		Message: "Invalid Request",
	}

	data, mErr := json.Marshal(e)
	if mErr != nil {
		t.Errorf("Marshal() error = %v", mErr)
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Unmarshal() error = %v", err)
		return
	}

	if parsed["code"] != float64(-32600) {
		t.Errorf("Expected code -32600, got %v", parsed["code"])
	}
	if parsed["message"] != "Invalid Request" {
		t.Errorf("Expected message Invalid Request, got %v", parsed["message"])
	}
}
