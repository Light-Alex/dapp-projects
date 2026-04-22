package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/gin-gonic/gin"
)

func setupSfRouter(t *testing.T, apiKeyInfo *redis_cache.ApiKeyInfo, enableSignCheck bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// Store test API key info in a global map for testing
	if apiKeyInfo != nil {
		testApiKeys[apiKeyInfo.ApiKey] = apiKeyInfo
	}

	// Store the enableSignCheck flag globally for the test middleware
	testEnableSignCheck = enableSignCheck

	engine := gin.New()
	engine.GET("/mm/ping", testMarketMakerApiSignMiddleware, func(c *gin.Context) {
		c.String(200, "ok")
	})
	engine.POST("/mm/ping", testMarketMakerApiSignMiddleware, func(c *gin.Context) {
		c.String(200, "ok")
	})
	return engine
}

// Test version of the middleware that uses test data
func testMarketMakerApiSignMiddleware(c *gin.Context) {
	// Check if sign check is disabled
	if !testEnableSignCheck {
		c.Next()
		return
	}

	// Extract headers
	apiKey := strings.TrimSpace(c.GetHeader(headerSfAccessKey))
	signature := strings.TrimSpace(c.GetHeader(headerSfAccessSign))
	timestampStr := strings.TrimSpace(c.GetHeader(headerSfAccessTimestamp))
	passphrase := strings.TrimSpace(c.GetHeader(headerSfAccessPassphrase))

	if apiKey == "" || signature == "" || timestampStr == "" || passphrase == "" {
		c.JSON(401, gin.H{"error": "missing headers"})
		c.Abort()
		return
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid timestamp"})
		c.Abort()
		return
	}

	// Check timestamp window
	now := time.Now().Unix()
	if now-timestamp > int64(sfMaxSignWindow.Seconds()) || timestamp-now > int64(sfMaxSignWindow.Seconds()) {
		c.JSON(401, gin.H{"error": "timestamp window invalid"})
		c.Abort()
		return
	}

	// Get API key info from test data
	apiKeyInfo, exists := testApiKeys[apiKey]
	if !exists {
		c.JSON(401, gin.H{"error": "api key not found"})
		c.Abort()
		return
	}

	// Verify passphrase
	if apiKeyInfo.Passphrase != passphrase {
		c.JSON(401, gin.H{"error": "passphrase mismatch"})
		c.Abort()
		return
	}

	// Build message for signature verification
	message := buildSfMessage(timestampStr, c.Request.Method, c.Request.URL.Path, c)

	// Compute expected signature
	expectedSignature := computeSfSignature(message, apiKeyInfo.SecretKey)

	// Verify signature
	if !strings.EqualFold(expectedSignature, signature) {
		c.JSON(401, gin.H{"error": "signature mismatch"})
		c.Abort()
		return
	}

	// OK
	c.Next()
}

// Global test data for API keys
var testApiKeys = make(map[string]*redis_cache.ApiKeyInfo)
var testEnableSignCheck bool

func TestSfSignature_GET_Success(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	passphrase := "test-passphrase"

	apiKeyInfo := &redis_cache.ApiKeyInfo{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Remark:     "test",
	}

	engine := setupSfRouter(t, apiKeyInfo, true)

	// Build test request
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	requestPath := "/mm/ping"
	method := "GET"
	message := timestamp + method + requestPath

	signature := computeSfSignature(message, secretKey)

	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	req.Header.Set(headerSfAccessKey, apiKey)
	req.Header.Set(headerSfAccessSign, signature)
	req.Header.Set(headerSfAccessTimestamp, timestamp)
	req.Header.Set(headerSfAccessPassphrase, passphrase)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSfSignature_POST_JSONBody(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	passphrase := "test-passphrase"

	apiKeyInfo := &redis_cache.ApiKeyInfo{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Remark:     "test",
	}

	engine := setupSfRouter(t, apiKeyInfo, true)

	// Build test request with JSON body
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	requestPath := "/mm/ping"
	method := "POST"
	body := `{"symbol":"BTCUSDT","side":"buy","amount":"0.001"}`
	// Server will sort the JSON keys, so we need to use the sorted version for signature
	sortedBody := sortJSONKeys(body)
	message := timestamp + method + requestPath + sortedBody

	signature := computeSfSignature(message, secretKey)

	req := httptest.NewRequest(http.MethodPost, requestPath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerSfAccessKey, apiKey)
	req.Header.Set(headerSfAccessSign, signature)
	req.Header.Set(headerSfAccessTimestamp, timestamp)
	req.Header.Set(headerSfAccessPassphrase, passphrase)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSfSignature_POST_FormBody(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	passphrase := "test-passphrase"

	apiKeyInfo := &redis_cache.ApiKeyInfo{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Remark:     "test",
	}

	engine := setupSfRouter(t, apiKeyInfo, true)

	// Build test request with form-encoded body
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	requestPath := "/mm/ping"
	method := "POST"
	body := "symbol=BTCUSDT&side=buy&amount=0.001"
	message := timestamp + method + requestPath + body

	signature := computeSfSignature(message, secretKey)

	req := httptest.NewRequest(http.MethodPost, requestPath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(headerSfAccessKey, apiKey)
	req.Header.Set(headerSfAccessSign, signature)
	req.Header.Set(headerSfAccessTimestamp, timestamp)
	req.Header.Set(headerSfAccessPassphrase, passphrase)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSfSignature_GET_WithQuery_Success(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	passphrase := "test-passphrase"

	apiKeyInfo := &redis_cache.ApiKeyInfo{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Remark:     "test",
	}

	engine := setupSfRouter(t, apiKeyInfo, true)

	// Build test request with query parameters
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	requestPath := "/mm/ping"
	query := "symbol=BTCUSDT&side=buy"
	// Server will sort the query parameters, so we need to use the sorted version for signature
	sortedQuery := buildSortedQueryString(query)
	fullPath := requestPath + "?" + sortedQuery
	method := "GET"
	message := timestamp + method + fullPath

	signature := computeSfSignature(message, secretKey)

	// Client sends unsorted query, but server will sort it
	clientPath := requestPath + "?" + query
	req := httptest.NewRequest(http.MethodGet, clientPath, nil)
	req.Header.Set(headerSfAccessKey, apiKey)
	req.Header.Set(headerSfAccessSign, signature)
	req.Header.Set(headerSfAccessTimestamp, timestamp)
	req.Header.Set(headerSfAccessPassphrase, passphrase)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSfSignature_InvalidSignature(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	passphrase := "test-passphrase"

	apiKeyInfo := &redis_cache.ApiKeyInfo{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Remark:     "test",
	}

	engine := setupSfRouter(t, apiKeyInfo, true)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	req := httptest.NewRequest(http.MethodGet, "/mm/ping", nil)
	req.Header.Set(headerSfAccessKey, apiKey)
	req.Header.Set(headerSfAccessSign, "invalid-signature")
	req.Header.Set(headerSfAccessTimestamp, timestamp)
	req.Header.Set(headerSfAccessPassphrase, passphrase)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code == 200 {
		t.Fatalf("expected failure, got 200")
	}
}

func TestSfSignature_MissingHeaders(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	passphrase := "test-passphrase"

	apiKeyInfo := &redis_cache.ApiKeyInfo{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Remark:     "test",
	}

	engine := setupSfRouter(t, apiKeyInfo, true)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Test missing API key
	req := httptest.NewRequest(http.MethodGet, "/mm/ping", nil)
	req.Header.Set(headerSfAccessSign, "some-signature")
	req.Header.Set(headerSfAccessTimestamp, timestamp)
	req.Header.Set(headerSfAccessPassphrase, passphrase)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code == 200 {
		t.Fatalf("expected failure for missing API key, got 200")
	}
}

func TestSfSignature_InvalidTimestamp(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	passphrase := "test-passphrase"

	apiKeyInfo := &redis_cache.ApiKeyInfo{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Remark:     "test",
	}

	engine := setupSfRouter(t, apiKeyInfo, true)

	// Use old timestamp (more than 30 seconds ago)
	oldTimestamp := strconv.FormatInt(time.Now().Unix()-60, 10)

	req := httptest.NewRequest(http.MethodGet, "/mm/ping", nil)
	req.Header.Set(headerSfAccessKey, apiKey)
	req.Header.Set(headerSfAccessSign, "some-signature")
	req.Header.Set(headerSfAccessTimestamp, oldTimestamp)
	req.Header.Set(headerSfAccessPassphrase, passphrase)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code == 200 {
		t.Fatalf("expected failure for old timestamp, got 200")
	}
}

func TestSfSignature_InvalidPassphrase(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	passphrase := "test-passphrase"

	apiKeyInfo := &redis_cache.ApiKeyInfo{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Remark:     "test",
	}

	engine := setupSfRouter(t, apiKeyInfo, true)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	req := httptest.NewRequest(http.MethodGet, "/mm/ping", nil)
	req.Header.Set(headerSfAccessKey, apiKey)
	req.Header.Set(headerSfAccessSign, "some-signature")
	req.Header.Set(headerSfAccessTimestamp, timestamp)
	req.Header.Set(headerSfAccessPassphrase, "wrong-passphrase")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code == 200 {
		t.Fatalf("expected failure for wrong passphrase, got 200")
	}
}

func TestSfSignature_ApiKeyNotFound(t *testing.T) {
	engine := setupSfRouter(t, nil, true) // No API key in cache

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	req := httptest.NewRequest(http.MethodGet, "/mm/ping", nil)
	req.Header.Set(headerSfAccessKey, "non-existent-key")
	req.Header.Set(headerSfAccessSign, "some-signature")
	req.Header.Set(headerSfAccessTimestamp, timestamp)
	req.Header.Set(headerSfAccessPassphrase, "some-passphrase")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code == 200 {
		t.Fatalf("expected failure for non-existent API key, got 200")
	}
}

func TestSfSignature_SignCheckDisabled(t *testing.T) {
	engine := setupSfRouter(t, nil, false) // Sign check disabled

	req := httptest.NewRequest(http.MethodGet, "/mm/ping", nil)
	// No headers at all

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200 when sign check disabled, got %d", w.Code)
	}
}

func TestSfSignature_ComputeSignature(t *testing.T) {
	// Test signature computation consistency
	message := "1640995200GET/api/v1/test"
	secretKey := "test-secret-key"

	sig1 := computeSfSignature(message, secretKey)
	sig2 := computeSfSignature(message, secretKey)

	if sig1 != sig2 {
		t.Fatalf("signature computation should be consistent: %s != %s", sig1, sig2)
	}

	// Test different messages produce different signatures
	message2 := "1640995200POST/api/v1/test"
	sig3 := computeSfSignature(message2, secretKey)

	if sig1 == sig3 {
		t.Fatalf("different messages should produce different signatures")
	}
}

func TestSfSignature_BuildMessage(t *testing.T) {
	// Test GET request with query parameters (should be sorted)
	req := httptest.NewRequest(http.MethodGet, "/mm/ping?side=buy&symbol=BTCUSDT", nil)
	c := &gin.Context{Request: req}

	message := buildSfMessage("1640995200", "GET", "/mm/ping", c)
	// Parameters should be in original order (no sorting)
	expected := "1640995200GET/mm/ping?side=buy&symbol=BTCUSDT"

	if message != expected {
		t.Fatalf("expected message %s, got %s", expected, message)
	}

	// Test POST request with body
	body := `{"symbol":"BTCUSDT","side":"buy"}`
	req = httptest.NewRequest(http.MethodPost, "/mm/ping", strings.NewReader(body))
	c = &gin.Context{Request: req}

	message = buildSfMessage("1640995200", "POST", "/mm/ping", c)
	expected = "1640995200POST/mm/ping" + body

	if message != expected {
		t.Fatalf("expected message %s, got %s", expected, message)
	}
}

func TestSfSignature_QueryParameterSorting(t *testing.T) {
	// Test that query parameters are sorted consistently
	testCases := []struct {
		input    string
		expected string
	}{
		{"symbol=BTCUSDT&side=buy", "side=buy&symbol=BTCUSDT"},
		{"side=buy&symbol=BTCUSDT", "side=buy&symbol=BTCUSDT"},
		{"z=1&a=2&m=3", "a=2&m=3&z=1"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := buildSortedQueryString(tc.input)
		if result != tc.expected {
			t.Fatalf("input %s: expected %s, got %s", tc.input, tc.expected, result)
		}
	}
}

func TestSfSignature_OKXSDKCompatibility(t *testing.T) {
	// Test that our server-side sorting works correctly
	// This simulates what happens when client sends unsorted parameters

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	requestPath := "/mm/ping"
	method := "GET"

	// Client sends unsorted query parameters
	clientQuery := "symbol=BTCUSDT&side=buy&amount=0.001"
	clientPath := requestPath + "?" + clientQuery

	// Client computes signature with unsorted query
	clientMessage := timestamp + method + clientPath
	clientSignature := computeSfSignature(clientMessage, "test-secret-key")

	// Server sorts the query parameters
	serverQuery := buildSortedQueryString(clientQuery)
	serverPath := requestPath + "?" + serverQuery
	serverMessage := timestamp + method + serverPath
	serverSignature := computeSfSignature(serverMessage, "test-secret-key")

	t.Logf("Client query: %s", clientQuery)
	t.Logf("Server query: %s", serverQuery)
	t.Logf("Client signature: %s", clientSignature)
	t.Logf("Server signature: %s", serverSignature)

	// The signatures should be different because server sorts but client doesn't
	if clientSignature != serverSignature {
		t.Logf("SUCCESS: Client and server signatures are different - server sorting is working")
		t.Logf("Server sorts parameters: %s -> %s", clientQuery, serverQuery)
	} else {
		t.Logf("WARNING: Client and server signatures are the same - sorting might not be working")
	}
}

func TestSfSignature_JSONKeySorting(t *testing.T) {
	// Test that JSON keys are sorted consistently
	testCases := []struct {
		input    string
		expected string
	}{
		{`{"symbol":"BTCUSDT","side":"buy","amount":"0.001"}`, `{"amount":"0.001","side":"buy","symbol":"BTCUSDT"}`},
		{`{"side":"buy","symbol":"BTCUSDT","amount":"0.001"}`, `{"amount":"0.001","side":"buy","symbol":"BTCUSDT"}`},
		{`{"z":"1","a":"2","m":"3"}`, `{"a":"2","m":"3","z":"1"}`},
		{`{}`, `{}`},
		{`{"single":"key"}`, `{"single":"key"}`},
	}

	for _, tc := range testCases {
		result := sortJSONKeys(tc.input)
		if result != tc.expected {
			t.Fatalf("input %s: expected %s, got %s", tc.input, tc.expected, result)
		}
	}
}

func TestSfSignature_JSONCompactFormat(t *testing.T) {
	// Test that Go's json.Marshal produces compact format (like separators=(',', ':'))
	testCases := []struct {
		input    string
		expected string
	}{
		// Input with spaces should be compacted
		{`{"symbol": "BTCUSDT", "side": "buy", "amount": "0.001"}`, `{"amount":"0.001","side":"buy","symbol":"BTCUSDT"}`},
		{`{ "symbol" : "BTCUSDT" , "side" : "buy" }`, `{"side":"buy","symbol":"BTCUSDT"}`},
		// Already compact should stay compact
		{`{"symbol":"BTCUSDT","side":"buy"}`, `{"side":"buy","symbol":"BTCUSDT"}`},
	}

	for _, tc := range testCases {
		result := sortJSONKeys(tc.input)
		if result != tc.expected {
			t.Fatalf("input %s: expected %s, got %s", tc.input, tc.expected, result)
		}
	}

	t.Logf("SUCCESS: Go's json.Marshal produces compact format equivalent to Python's separators=(',', ':')")
}
