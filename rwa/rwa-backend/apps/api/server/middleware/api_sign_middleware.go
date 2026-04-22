package middleware

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AnchoredLabs/rwa-backend/apps/api/config"
	"github.com/AnchoredLabs/rwa-backend/libs/core/web"
	"github.com/AnchoredLabs/rwa-backend/libs/log"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	headerNonce         = "X-Api-Nonce"
	headerSign          = "X-Api-Sign"
	headerTimestamp     = "X-Api-Ts"
	headerAuthorization = "Authorization"
	maxSignWindow       = 45 * time.Second
)

func ApiSignMiddleware(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiSignMiddleware(conf, c)
	}
}

func apiSignMiddleware(conf *config.Config, c *gin.Context) {
	ctx := c.Request.Context()

	if !conf.Server.EnableSignCheck {
		c.Next()
		return
	}

	nonce := strings.TrimSpace(firstHeaderValue(c, headerNonce))
	signature := strings.TrimSpace(firstHeaderValue(c, headerSign))
	tsHeader := strings.TrimSpace(firstHeaderValue(c, headerTimestamp))

	if nonce == "" || signature == "" || tsHeader == "" {
		log.WarnZ(ctx, "api sign headers missing",
			zap.String("uri", c.Request.RequestURI),
			zap.String("nonce", nonce),
			zap.String("sign", signature),
			zap.String("ts", tsHeader))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		log.WarnZ(ctx, "api sign timestamp parse error",
			zap.String("uri", c.Request.RequestURI),
			zap.String("nonce", nonce),
			zap.String("sign", signature),
			zap.String("ts", tsHeader),
			zap.Error(err))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	if time.Now().UnixMilli()-ts > maxSignWindow.Milliseconds() {
		log.WarnZ(ctx, "api sign timestamp expired",
			zap.String("uri", c.Request.RequestURI),
			zap.String("nonce", nonce),
			zap.Int64("ts", ts),
			zap.Duration("window", maxSignWindow))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	bodyStr, err := extractCanonicalBody(c)
	if err != nil {
		log.WarnZ(ctx, "api sign extract body failed",
			zap.String("uri", c.Request.RequestURI),
			zap.Error(err))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	var bodyPtr *string
	if bodyStr != "" {
		bodyPtr = &bodyStr
	}
	authHeader := strings.TrimSpace(firstHeaderValue(c, headerAuthorization))
	var authPtr *string
	if authHeader != "" {
		authPtr = &authHeader
	}

	requestURI := buildSortedRequestURI(c.Request)

	signer := newAPISigner(nonce)
	expected, err := signer.sign(requestURI, ts, bodyPtr, authPtr)
	if err != nil {
		log.WarnZ(ctx, "api sign compute failed",
			zap.String("uri", c.Request.RequestURI),
			zap.Error(err))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	if !strings.EqualFold(expected, signature) {
		log.WarnZ(ctx, "api sign mismatch",
			zap.String("uri", c.Request.RequestURI),
			zap.String("nonce", nonce),
			zap.String("expected", expected),
			zap.String("provided", signature))
		web.ResponseUnAuthorizedError(c)
		c.Abort()
		return
	}

	c.Next()
}

func firstHeaderValue(c *gin.Context, key string) string {
	header := c.GetHeader(key)
	if header == "" {
		return ""
	}
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func extractCanonicalBody(c *gin.Context) (string, error) {
	if c.Request == nil || c.Request.Body == nil {
		return "", nil
	}

	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		return "", nil
	}

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}

	if err := c.Request.Body.Close(); err != nil {
		return "", err
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))
	trimmed := bytes.TrimSpace(rawBody)
	if len(trimmed) == 0 {
		return "", nil
	}

	canonical, err := canonicalizeJSON(trimmed)
	if err != nil {
		return "", err
	}
	return canonical, nil
}

func canonicalizeJSON(data []byte) (string, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return string(bytes.TrimSpace(data)), nil
	}

	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.Grow(len(data))
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyJSON, _ := json.Marshal(key)
		buf.Write(keyJSON)
		buf.WriteByte(':')

		value := raw[key]
		var compact bytes.Buffer
		if err := json.Compact(&compact, value); err != nil {
			compact.Write(value)
		}
		buf.Write(compact.Bytes())
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

func buildSortedRequestURI(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}

	uri := r.URL.Path
	query := r.URL.Query()
	if len(query) == 0 {
		return uri
	}

	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(query))
	for _, key := range keys {
		values := query[key]
		for _, value := range values {
			pairs = append(pairs, key+"="+value)
		}
	}
	return uri + "?" + strings.Join(pairs, "&")
}

type apiSigner struct {
	nonce string
	key   []byte
	iv    []byte
}

func newAPISigner(nonce string) *apiSigner {
	// use Keccak256 to match web3j's Hash.sha3String implementation
	hash := crypto.Keccak256([]byte(nonce))
	key := make([]byte, 32)
	copy(key, hash[:32])

	iv := make([]byte, 16)
	copy(iv, hash[len(hash)-16:])

	return &apiSigner{
		nonce: nonce,
		key:   key,
		iv:    iv,
	}
}

func (s *apiSigner) sign(uri string, ts int64, body *string, authorization *string) (string, error) {
	// Build JSON manually to match Java's JSON serialization
	// We need to escape strings properly but without HTML escaping (\u0026 for &)
	type Payload struct {
		URI           string  `json:"uri"`
		Body          *string `json:"body,omitempty"`
		Nonce         string  `json:"nonce"`
		Timestamp     int64   `json:"ts"`
		Authorization *string `json:"authorization,omitempty"`
	}

	payload := Payload{
		URI:           uri,
		Body:          body,
		Nonce:         s.nonce,
		Timestamp:     ts,
		Authorization: authorization,
	}

	// Use json.Marshal with custom encoder to avoid HTML escaping
	buffer := &strings.Builder{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false) // Don't escape & as \u0026
	if err := encoder.Encode(payload); err != nil {
		return "", err
	}

	// Remove trailing newline added by Encode
	plaintext := []byte(strings.TrimSuffix(buffer.String(), "\n"))

	base64Cipher, err := s.encrypt(plaintext)
	if err != nil {
		return "", err
	}

	doubleBase64 := base64.StdEncoding.EncodeToString([]byte(base64Cipher))
	hash := crypto.Keccak256([]byte(doubleBase64))
	return hex.EncodeToString(hash), nil
}

func (s *apiSigner) encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	padded := pkcs7Pad(plaintext, block.BlockSize())

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, s.iv)
	mode.CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
