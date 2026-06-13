package workos

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const validateURL = "https://api.workos.com/api_keys/validations"

type apiKeyOwner struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type validateResponse struct {
	APIKey *struct {
		ID    string      `json:"id"`
		Owner apiKeyOwner `json:"owner"`
	} `json:"api_key"`
}

type Owner struct {
	UserID         string
	OrganizationID string
}

type cacheEntry struct {
	owner     Owner
	expiresAt time.Time
}

type APIKeyValidator struct {
	apiKey       string
	defaultOrgID string
	httpClient   *http.Client
	ttl          time.Duration
	mu           sync.RWMutex
	cache        map[string]cacheEntry
}

func NewAPIKeyValidator(apiKey, defaultOrgID string) *APIKeyValidator {
	return &APIKeyValidator{
		apiKey:       apiKey,
		defaultOrgID: defaultOrgID,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		ttl:          60 * time.Second,
		cache:        make(map[string]cacheEntry),
	}
}

func (v *APIKeyValidator) Validate(ctx context.Context, key string) (string, string, error) {
	if owner, ok := v.lookup(key); ok {
		return owner.UserID, owner.OrganizationID, nil
	}
	owner, err := v.validateRemote(ctx, key)
	if err != nil {
		return "", "", err
	}
	v.store(key, owner)
	return owner.UserID, owner.OrganizationID, nil
}

func (v *APIKeyValidator) validateRemote(ctx context.Context, key string) (Owner, error) {
	res, err := v.postValidation(ctx, key)
	if err != nil {
		return Owner{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Owner{}, fmt.Errorf("api key validation failed: status %d", res.StatusCode)
	}
	return v.parseOwner(res)
}

func (v *APIKeyValidator) postValidation(ctx context.Context, key string) (*http.Response, error) {
	body, err := json.Marshal(map[string]string{"value": key})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, validateURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+v.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return v.httpClient.Do(req)
}

func (v *APIKeyValidator) parseOwner(res *http.Response) (Owner, error) {
	var parsed validateResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return Owner{}, err
	}
	if parsed.APIKey == nil || parsed.APIKey.Owner.Type != "user" {
		return Owner{}, fmt.Errorf("api key is not user-scoped")
	}
	return Owner{UserID: parsed.APIKey.Owner.ID, OrganizationID: v.defaultOrgID}, nil
}

func (v *APIKeyValidator) lookup(key string) (Owner, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	entry, ok := v.cache[cacheKey(key)]
	if !ok || time.Now().After(entry.expiresAt) {
		return Owner{}, false
	}
	return entry.owner, true
}

func (v *APIKeyValidator) store(key string, owner Owner) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cache[cacheKey(key)] = cacheEntry{owner: owner, expiresAt: time.Now().Add(v.ttl)}
}

func cacheKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
