package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/terrasquid/terraform-provider-terrasquid/internal/model"
)

func TestListDestinationConfigs(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %q, want %q", r.Method, "GET")
		}
		if r.URL.Path != "/api/v1/destinations/" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v1/destinations/")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"id": "dc-1",
				"service": "svc",
				"name": "dc-one",
				"key_prefix": "",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
				"dst": "10.0.0.1",
				"type": "allow",
				"ports": [80, 443]
			}
		]`))
	}
	client, _ := newTestClient(t, handler)
	configs, err := client.ListDestinationConfigs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("len(configs) = %d, want 1", len(configs))
	}
	dc := configs[0]
	if dc.ID != "dc-1" {
		t.Errorf("ID = %q, want %q", dc.ID, "dc-1")
	}
	if dc.Dst != "10.0.0.1" {
		t.Errorf("Dst = %q, want %q", dc.Dst, "10.0.0.1")
	}
	if dc.Type != "allow" {
		t.Errorf("Type = %q, want %q", dc.Type, "allow")
	}
	if len(dc.Ports) != 2 || dc.Ports[0] != 80 || dc.Ports[1] != 443 {
		t.Errorf("Ports = %v, want [80 443]", dc.Ports)
	}
}

func TestCreateDestinationConfig(t *testing.T) {
	var gotBody []byte
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want %q", r.Method, "POST")
		}
		if r.URL.Path != "/api/v1/destinations/" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v1/destinations/")
		}
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "dc-new",
			"service": "svc",
			"name": "new-dc",
			"key_prefix": "",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z",
			"dst": "192.168.1.1",
			"type": "deny",
			"ports": [22]
		}`))
	}
	client, _ := newTestClient(t, handler)
	input := model.DestinationConfigInput{
		Name:  "new-dc",
		Dst:   "192.168.1.1",
		Type:  "deny",
		Ports: []int{22},
	}
	dc, err := client.CreateDestinationConfig(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed model.DestinationConfigInput
	if err := json.Unmarshal(gotBody, &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if parsed.Name != "new-dc" {
		t.Errorf("body name = %q, want %q", parsed.Name, "new-dc")
	}
	if dc.ID != "dc-new" {
		t.Errorf("ID = %q, want %q", dc.ID, "dc-new")
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["port_ranges"]; ok {
		t.Fatal("legacy create must omit unset port_ranges")
	}
}

func TestDestinationConfigPortRangesWireFormat(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, ranges := range [][]string{nil, {}, {"30000-32767"}} {
			t.Run(method+fmt.Sprint(ranges), func(t *testing.T) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method != method {
						t.Errorf("method = %s, want %s", r.Method, method)
					}
					var body map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					raw, present := body["port_ranges"]
					if present != (ranges != nil) {
						t.Errorf("port_ranges present = %v, want %v", present, ranges != nil)
					}
					if ranges != nil {
						var got []string
						if err := json.Unmarshal(raw, &got); err != nil {
							t.Fatal(err)
						}
						if got == nil || !slices.Equal(got, ranges) {
							t.Errorf("port_ranges = %v, want explicit list %v", got, ranges)
						}
					}
					_ = json.NewEncoder(w).Encode(body)
				}
				client, _ := newTestClient(t, handler)
				input := model.DestinationConfigInput{Name: "ranges", Dst: "example.com", Type: "ALLOW"}
				if ranges != nil {
					input.PortRanges = &ranges
				}
				var result *model.DestinationConfig
				var err error
				if method == http.MethodPost {
					result, err = client.CreateDestinationConfig(context.Background(), input)
				} else {
					result, err = client.UpdateDestinationConfig(context.Background(), "dc-1", input)
				}
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(result.PortRanges, ranges) {
					t.Errorf("response ranges = %v, want %v", result.PortRanges, ranges)
				}
			})
		}
	}
}

func TestGetDestinationConfig(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %q, want %q", r.Method, "GET")
		}
		if r.URL.Path != "/api/v1/destinations/dc-1/" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v1/destinations/dc-1/")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "dc-1",
			"service": "svc",
			"name": "dc-one",
			"key_prefix": "",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z",
			"dst": "10.0.0.1",
			"type": "allow",
			"ports": [80]
		}`))
	}
	client, _ := newTestClient(t, handler)
	dc, err := client.GetDestinationConfig(context.Background(), "dc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dc.ID != "dc-1" {
		t.Errorf("ID = %q, want %q", dc.ID, "dc-1")
	}
}

func TestUpdateDestinationConfig(t *testing.T) {
	var gotBody []byte
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("method = %q, want %q", r.Method, "PUT")
		}
		if r.URL.Path != "/api/v1/destinations/dc-1/" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v1/destinations/dc-1/")
		}
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "dc-1",
			"service": "svc",
			"name": "updated-dc",
			"key_prefix": "",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-02T00:00:00Z",
			"dst": "10.0.0.2",
			"type": "allow",
			"ports": [8080]
		}`))
	}
	client, _ := newTestClient(t, handler)
	input := model.DestinationConfigInput{
		Name:  "updated-dc",
		Dst:   "10.0.0.2",
		Type:  "allow",
		Ports: []int{},
	}
	dc, err := client.UpdateDestinationConfig(context.Background(), "dc-1", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed model.DestinationConfigInput
	if err := json.Unmarshal(gotBody, &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if parsed.Name != "updated-dc" {
		t.Errorf("body name = %q, want %q", parsed.Name, "updated-dc")
	}
	if parsed.Ports == nil || len(parsed.Ports) != 0 {
		t.Errorf("body ports = %v, want an explicit empty list", parsed.Ports)
	}
	if dc.Name != "updated-dc" {
		t.Errorf("Name = %q, want %q", dc.Name, "updated-dc")
	}
}

func TestDeleteDestinationConfig(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %q, want %q", r.Method, "DELETE")
		}
		if r.URL.Path != "/api/v1/destinations/dc-1/" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/api/v1/destinations/dc-1/")
		}
		w.WriteHeader(http.StatusNoContent)
	}
	client, _ := newTestClient(t, handler)
	err := client.DeleteDestinationConfig(context.Background(), "dc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteDestinationConfig_Error(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found","error":"not_found"}`))
	}
	client, _ := newTestClient(t, handler)
	err := client.DeleteDestinationConfig(context.Background(), "dc-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected IsNotFoundError, got %v", err)
	}
}
