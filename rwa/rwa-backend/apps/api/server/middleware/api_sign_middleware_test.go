package middleware

import "testing"

func TestAPISignerMatchesJavaImplementation(t *testing.T) {
	const (
		nonce     = "AEBHTC8SbLgdhG6dfoe3wqhiuyeYchDob136NuBHebqR73Oo"
		timestamp = int64(1761041626745)
		expected  = "419dd6052a242ae5f44ab662e80fe3ae1b9e888818bad0984366243b85ad9cf5"
		uri       = "/spot/api/spot/pool/kline?address=0x7d148143b7033f150830ff9114797b54671dde2e&chainId=10143&endTime=1761045240&interval=1h&limit=1000"
	)

	signer := newAPISigner(nonce)
	signature, err := signer.sign(uri, timestamp, nil, nil)
	if err != nil {
		t.Fatalf("sign returned error: %v", err)
	}

	if signature != expected {
		t.Fatalf("signature mismatch, want %s got %s", expected, signature)
	}
}

func TestAPISignerWithBodyAndAuthorization(t *testing.T) {
	const (
		nonce     = "dR9jbh9CPHYWkIxktenUEyPb2VxJt5eJ1fU1PZL4a7RRBrHt"
		timestamp = int64(1700000000123)
		expected  = "08dc67e765813db77ca7d3573953a99796bf29a4beb8877a79a8dfd733cac1ff"
		uri       = "/api/v1/frontend-tx/submit"
	)

	bodyJSON, err := canonicalizeJSON([]byte(`{"b":2,"a":1,"c":{"z":9,"b":3}}`))
	if err != nil {
		t.Fatalf("canonicalizeJSON returned error: %v", err)
	}

	auth := "Bearer token123"
	signer := newAPISigner(nonce)
	signature, err := signer.sign(uri, timestamp, &bodyJSON, &auth)
	if err != nil {
		t.Fatalf("sign returned error: %v", err)
	}

	if signature != expected {
		t.Fatalf("signature mismatch, want %s got %s", expected, signature)
	}
}

func TestAPISignerDepthGetVector(t *testing.T) {
	const (
		nonce     = "KvcXqxgUUte8dgFR5RPtMQO7nkOfOvFckHGhzh4bcfRtQCX7"
		timestamp = int64(1761046960085)
		expected  = "ccdf34c8d2fe991a37980d7e19cda9c33bc9f5b6d3a8010563d6409a9877c918"
		uri       = "/spot/api/spot/pool/depth?address=0x7d148143b7033f150830ff9114797b54671dde2e&chainId=10143&ratio=10"
	)

	signer := newAPISigner(nonce)
	signature, err := signer.sign(uri, timestamp, nil, nil)
	if err != nil {
		t.Fatalf("sign returned error: %v", err)
	}

	if signature != expected {
		t.Fatalf("signature mismatch, want %s got %s", expected, signature)
	}
}

func TestAPISignerTxRecordPostVector(t *testing.T) {
	const (
		nonce     = "iQQ0kqpZnNPk4cFCmtlHiREYQ1HsqHqSo7JT3rTkgPFbZ0Q0"
		timestamp = int64(1761047527124)
		expected  = "a70bfac2660388cbeebdc93aa900310b72fd03496df71fd77c01caa5a50932b2"
		uri       = "/spot/api/spot/txRecord"
	)

	// Request body from the example
	bodyInput := `{"chainId":10143,"txHash":"0x1ee1d15aa43e4e033a6e4222bd8d3ca6ae515edf480de080370acd48e1c90c50","walletType":"metaMask"}`
	bodyJSON, err := canonicalizeJSON([]byte(bodyInput))
	if err != nil {
		t.Fatalf("canonicalizeJSON returned error: %v", err)
	}

	signer := newAPISigner(nonce)
	signature, err := signer.sign(uri, timestamp, &bodyJSON, nil)
	if err != nil {
		t.Fatalf("sign returned error: %v", err)
	}

	if signature != expected {
		t.Fatalf("signature mismatch, want %s got %s", expected, signature)
	}
}
