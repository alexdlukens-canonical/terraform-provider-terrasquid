package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/terrasquid/terraform-provider-terrasquid/internal/model"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"terrasquid": providerserver.NewProtocol6WithError(New()),
}

func testAccProviderConfig() string {
	return `provider "terrasquid" {}`
}

type mockStore struct {
	mu          sync.Mutex
	nextID      int
	requireAuth bool
	sourceACLs  map[string]model.SourceACL
	destConfigs map[string]model.DestinationConfig
	destGroups  map[string]model.DestinationGroup
	aclRules    map[string]model.ACLRule
}

func newMockStore() *mockStore {
	return &mockStore{
		sourceACLs:  make(map[string]model.SourceACL),
		destConfigs: make(map[string]model.DestinationConfig),
		destGroups:  make(map[string]model.DestinationGroup),
		aclRules:    make(map[string]model.ACLRule),
	}
}

func (s *mockStore) newID() string {
	s.nextID++
	return fmt.Sprintf("test-id-%d", s.nextID)
}

func (s *mockStore) baseResource(id, name string) model.BaseResource {
	return model.BaseResource{
		ID:        id,
		Service:   "terrasquid",
		Name:      name,
		KeyPrefix: "/test/",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func newMockServer(t *testing.T) (*httptest.Server, *mockStore) {
	store := newMockStore()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/v1/") {
			path = "/" + strings.TrimPrefix(path, "/api/v1/")
			r.URL.Path = path
		}

		if path == "/status/" && r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(model.Status{
				DBConfigVersion:      1,
				AppliedConfigVersion: 1,
				LastReload:           "2024-01-01T00:00:00Z",
				LastReloadOK:         true,
				Unit:                 "terrasquid",
			})
			return
		}

		if store.requireAuth {
			auth := r.Header.Get("Authorization")
			if auth != "Api-Key valid-key" {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Unauthorized",
					"message": "Invalid API key",
				})
				return
			}
		}

		switch {
		case strings.HasPrefix(path, "/sources/"):
			handleSourceACLs(store, w, r)
		case strings.HasPrefix(path, "/destinations/"):
			handleDestConfigs(store, w, r)
		case strings.HasPrefix(path, "/destination-groups/"):
			handleDestinationGroups(store, w, r)
		case strings.HasPrefix(path, "/acl-rules/"):
			handleACLRules(store, w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "NotFound",
				"message": "Resource not found",
			})
		}
	}))
	t.Cleanup(srv.Close)
	return srv, store
}

func handleDestinationGroups(s *mockStore, w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := r.URL.Path

	switch {
	case path == "/destination-groups/" && r.Method == http.MethodPost:
		var input model.DestinationGroupInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if destinationGroupNameExists(s, input.Name, "") {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if !destinationGroupDestinationsExist(s, input.Destinations) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id := s.newID()
		destinations := append([]string(nil), input.Destinations...)
		sort.Strings(destinations)
		item := model.DestinationGroup{BaseResource: s.baseResource(id, input.Name), Destinations: destinations, Comment: input.Comment}
		s.destGroups[id] = item
		_ = json.NewEncoder(w).Encode(item)
	case path == "/destination-groups/" && r.Method == http.MethodGet:
		var items []model.DestinationGroup
		for _, item := range s.destGroups {
			if name := r.URL.Query().Get("name"); name == "" || item.Name == name {
				items = append(items, item)
			}
		}
		_ = json.NewEncoder(w).Encode(items)
	default:
		id := extractID(path, "/destination-groups/")
		item, ok := s.destGroups[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodPut:
			var input model.DestinationGroupInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if destinationGroupNameExists(s, input.Name, id) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			if !destinationGroupDestinationsExist(s, input.Destinations) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			destinations := append([]string(nil), input.Destinations...)
			sort.Strings(destinations)
			item.Name, item.Destinations, item.Comment, item.UpdatedAt = input.Name, destinations, input.Comment, time.Now()
			s.destGroups[id] = item
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodDelete:
			if destinationGroupIsReferenced(s, id) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			delete(s.destGroups, id)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func destinationGroupNameExists(s *mockStore, name, excludedID string) bool {
	for id, group := range s.destGroups {
		if id != excludedID && group.Name == name {
			return true
		}
	}
	return false
}

func destinationGroupDestinationsExist(s *mockStore, destinationIDs []string) bool {
	if len(destinationIDs) == 0 {
		return false
	}
	for _, destinationID := range destinationIDs {
		if _, exists := s.destConfigs[destinationID]; !exists {
			return false
		}
	}
	return true
}

func destinationGroupIsReferenced(s *mockStore, groupID string) bool {
	for _, rule := range s.aclRules {
		for _, referencedGroupID := range rule.DestinationGroups {
			if referencedGroupID == groupID {
				return true
			}
		}
	}
	return false
}

func extractID(path, prefix string) string {
	if path == prefix {
		return ""
	}
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, "/") {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(path, prefix), "/")
}

func handleSourceACLs(s *mockStore, w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := r.URL.Path

	switch {
	case path == "/sources/" && r.Method == http.MethodPost:
		var input model.SourceACLInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id := s.newID()
		item := model.SourceACL{
			BaseResource: s.baseResource(id, input.Name),
			CIDR:         input.CIDR,
			Comment:      input.Comment,
		}
		s.sourceACLs[id] = item
		_ = json.NewEncoder(w).Encode(item)

	case path == "/sources/" && r.Method == http.MethodGet:
		var items []model.SourceACL
		for _, v := range s.sourceACLs {
			items = append(items, v)
		}
		_ = json.NewEncoder(w).Encode(items)

	default:
		id := extractID(path, "/sources/")
		if id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		item, ok := s.sourceACLs[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "NotFound",
				"message": "Source ACL not found",
			})
			return
		}

		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodPut:
			var input model.SourceACLInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			item.Name = input.Name
			item.CIDR = input.CIDR
			item.Comment = input.Comment
			item.UpdatedAt = time.Now()
			s.sourceACLs[id] = item
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodDelete:
			delete(s.sourceACLs, id)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func handleDestConfigs(s *mockStore, w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := r.URL.Path

	switch {
	case path == "/destinations/" && r.Method == http.MethodPost:
		var input model.DestinationConfigInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id := s.newID()
		item := model.DestinationConfig{
			BaseResource: s.baseResource(id, input.Name),
			Dst:          input.Dst,
			Type:         input.Type,
			Ports:        input.Ports,
			Comment:      input.Comment,
		}
		s.destConfigs[id] = item
		_ = json.NewEncoder(w).Encode(item)

	case path == "/destinations/" && r.Method == http.MethodGet:
		var items []model.DestinationConfig
		for _, v := range s.destConfigs {
			items = append(items, v)
		}
		_ = json.NewEncoder(w).Encode(items)

	default:
		id := extractID(path, "/destinations/")
		if id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		item, ok := s.destConfigs[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "NotFound",
				"message": "Destination config not found",
			})
			return
		}

		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodPut:
			var input model.DestinationConfigInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			item.Name = input.Name
			item.Dst = input.Dst
			item.Type = input.Type
			item.Ports = input.Ports
			item.Comment = input.Comment
			item.UpdatedAt = time.Now()
			s.destConfigs[id] = item
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodDelete:
			delete(s.destConfigs, id)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func handleACLRules(s *mockStore, w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := r.URL.Path

	switch {
	case path == "/acl-rules/" && r.Method == http.MethodPost:
		var input model.ACLRuleInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if input.Destinations == nil || input.DestinationGroups == nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string][]string{"destinations": {"This field may not be null."}})
			return
		}
		id := s.newID()
		item := model.ACLRule{
			BaseResource:      s.baseResource(id, input.Name),
			Priority:          input.Priority,
			Comment:           input.Comment,
			Sources:           input.Sources,
			Destinations:      input.Destinations,
			DestinationGroups: input.DestinationGroups,
		}
		s.aclRules[id] = item
		_ = json.NewEncoder(w).Encode(item)

	case path == "/acl-rules/" && r.Method == http.MethodGet:
		var items []model.ACLRule
		for _, v := range s.aclRules {
			items = append(items, v)
		}
		_ = json.NewEncoder(w).Encode(items)

	default:
		id := extractID(path, "/acl-rules/")
		if id == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		item, ok := s.aclRules[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "NotFound",
				"message": "ACL rule not found",
			})
			return
		}

		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodPut:
			var input model.ACLRuleInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if input.Destinations == nil || input.DestinationGroups == nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string][]string{"destinations": {"This field may not be null."}})
				return
			}
			item.Priority = input.Priority
			item.Comment = input.Comment
			item.Sources = input.Sources
			item.Destinations = input.Destinations
			item.DestinationGroups = input.DestinationGroups
			item.UpdatedAt = time.Now()
			s.aclRules[id] = item
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodDelete:
			delete(s.aclRules, id)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func TestAccProvider_InvalidAPIKey(t *testing.T) {
	srv, store := newMockServer(t)
	store.requireAuth = true
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "invalid-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "terrasquid_status" "test" {}
`,
				ExpectError: regexp.MustCompile(`(?s)Invalid Terrasquid Credentials.*API\s+error 401`),
			},
		},
	})
}

func TestAccProvider_InsecureTLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/sources/":
			_ = json.NewEncoder(w).Encode([]model.SourceACL{})
		case "/api/v1/status/":
			_ = json.NewEncoder(w).Encode(model.Status{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("TERRASQUID_ENDPOINT", srv.URL)
	t.Setenv("TERRASQUID_API_KEY", "valid-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `provider "terrasquid" {
  insecure = true
}

data "terrasquid_status" "test" {}
`,
			},
		},
	})
}
