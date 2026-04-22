package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/api/config"
	"github.com/AnchoredLabs/rwa-backend/libs/core/redis_cache"
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	headerSfAccessKey        = "SF-ACCESS-KEY"
	headerSfAccessSign       = "SF-ACCESS-SIGN"
	headerSfAccessTimestamp  = "SF-ACCESS-TIMESTAMP"
	headerSfAccessPassphrase = "SF-ACCESS-PASSPHRASE"
	sfMaxSignWindow          = 30 * time.Second // Market maker allows 30 seconds
)

func MarketMakerApiSignMiddleware(conf *config.Config, apiKeyCacheService *redis_cache.ApiKeyCacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		marketMakerApiSignMiddleware(conf, apiKeyCacheService, c)
	}
}

func marketMakerApiSignMiddleware(conf *config.Config, apiKeyCacheService *redis_cache.ApiKeyCacheService, c *gin.Context) {
	ctx := c.Request.Context()

	if !conf.Server.EnableMarketMakerSignCheck {
		c.Next()
		return
	}

	// Extract headers
	apiKey := strings.TrimSpace(c.GetHeader(headerSfAccessKey))
	signature := strings.TrimSpace(c.GetHeader(headerSfAccessSign))
	timestampStr := strings.TrimSpace(c.GetHeader(headerSfAccessTimestamp))
	passphrase := strings.TrimSpace(c.GetHeader(headerSfAccessPassphrase))

	if apiKey == "" || signature == "" || timestampStr == "" || passphrase == "" {
		log.WarnZ(ctx, "SF sign headers missing",
			zap.String("apiKey", apiKey),
			zap.String("signature", signature),
			zap.String("timestamp", timestampStr),
			zap.String("passphrase", passphrase))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		log.WarnZ(ctx, "SF timestamp parse error",
			zap.String("timestamp", timestampStr),
			zap.Error(err))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	// Check timestamp window (SF uses seconds, not milliseconds)
	now := time.Now().Unix()
	if now-timestamp > int64(sfMaxSignWindow.Seconds()) || timestamp-now > int64(sfMaxSignWindow.Seconds()) {
		log.WarnZ(ctx, "SF timestamp window invalid",
			zap.Int64("timestamp", timestamp),
			zap.Int64("now", now),
			zap.Int64("diff", now-timestamp))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	// Get API key info from cache
	apiKeyInfo, err := apiKeyCacheService.GetCache(ctx, apiKey)
	if err != nil {
		log.WarnZ(ctx, "SF api key not found", zap.String("apiKey", apiKey), zap.Error(err))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	// Verify passphrase
	if apiKeyInfo.Passphrase != passphrase {
		log.WarnZ(ctx, "SF passphrase mismatch", zap.String("apiKey", apiKey))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	// Build message for signature verification
	message := buildSfMessage(timestampStr, c.Request.Method, c.Request.URL.Path, c)

	// Compute expected signature
	expectedSignature := computeSfSignature(message, apiKeyInfo.SecretKey)

	// Verify signature
	if !strings.EqualFold(expectedSignature, signature) {
		log.WarnZ(ctx, "SF signature mismatch",
			zap.String("expected", expectedSignature),
			zap.String("got", signature),
			zap.String("message", message))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	// OK
	c.Next()
}

// buildSfMessage builds the message string for SF signature verification
// Format: timestamp + method + request_path + body
func buildSfMessage(timestamp, method, requestPath string, c *gin.Context) string {
	// Convert method to uppercase
	method = strings.ToUpper(method)

	// Get body for POST requests
	var body string
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				body = string(bodyBytes)
				// Restore body for downstream handlers
				c.Request.Body = io.NopCloser(strings.NewReader(body))

				// For JSON content, sort the keys (like json.dumps(params, sort_keys=True))
				ct := strings.ToLower(c.GetHeader("Content-Type"))
				if strings.HasPrefix(ct, "application/json") {
					body = sortJSONKeys(body)
				}
			}
		}
	}

	// For GET requests, include query parameters in the path (with sorting)
	if method == http.MethodGet && c.Request.URL.RawQuery != "" {
		sortedQuery := buildSortedQueryString(c.Request.URL.RawQuery)
		requestPath = requestPath + "?" + sortedQuery
	}

	return timestamp + method + requestPath + body
}

// sortJSONKeys sorts JSON keys alphabetically (like json.dumps(params, sort_keys=True, separators=(',', ':')))
// Go's json.Marshal automatically produces compact format without spaces
func sortJSONKeys(jsonStr string) string {
	if jsonStr == "" {
		return ""
	}

	// Parse JSON into a generic map
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If parsing fails, return original JSON string
		return jsonStr
	}

	// Marshal back with sorted keys
	sortedBytes, err := json.Marshal(data)
	if err != nil {
		// If marshaling fails, return original JSON string
		return jsonStr
	}

	return string(sortedBytes)
}

// buildSortedQueryString sorts query parameters alphabetically by key
func buildSortedQueryString(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	// Parse the query string
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		// If parsing fails, return original query string
		return rawQuery
	}

	// Get all keys and sort them
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build sorted query string
	var parts []string
	for _, key := range keys {
		for _, value := range values[key] {
			parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(value))
		}
	}

	return strings.Join(parts, "&")
}

// computeSfSignature computes SF signature using HMAC-SHA256 + Base64
func computeSfSignature(message, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
