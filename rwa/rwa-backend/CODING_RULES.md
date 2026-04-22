# Coding Rules for go-snapshot

This document establishes coding standards and rules for the go-snapshot repository.

## Naming Conventions

### Contract Caller Fields
When working with contract-generated types, use descriptive field names that clearly indicate the purpose:

- ✅ **DO**: Use `gateCaller` instead of `gate` for `*perp_gate.GateCaller`
- ✅ **DO**: Use `instrumentCaller` instead of `instrument` for `*perp_instrument.InstrumentCaller`
- ✅ **DO**: Use `multicall3Caller` instead of `multicall3` for `*multicall3.Multicall3Caller`

**Rationale**: This makes the code more self-documenting and prevents confusion between contract addresses, contract instances, and contract callers.

### Examples

```go
// ✅ Good - Clear and descriptive
type GateClient struct {
    ctx        context.Context
    client     *ethclient.Client
    gateCaller *perp_gate.GateCaller
    address    common.Address
}

// ❌ Bad - Ambiguous naming
type GateClient struct {
    ctx     context.Context
    client  *ethclient.Client
    gate    *perp_gate.GateCaller  // What is this? Address? Instance? Caller?
    address common.Address
}
```

### File Naming
- Use descriptive names that reflect the main functionality
- ✅ **DO**: `gate_client.go` instead of `gate_rpc_query.go`
- ✅ **DO**: `multicall3_client.go` instead of `multicall3_utils.go`

### Function Naming
- Use descriptive names that clearly indicate the purpose
- ✅ **DO**: `ExampleGateClient()` instead of `ExampleGateRPCQuery()`
- ✅ **DO**: `NewGateClient()` instead of `NewGateRPCQuerier()`

### Contract Method Naming
When creating wrapper methods for smart contract functions:

- ✅ **DO**: Prefix with "Query" to make it explicit these are on-chain queries
- ✅ **DO**: `QueryIndexOf()` instead of `IndexOf()`
- ✅ **DO**: `QueryReserveOf()` instead of `ReserveOf()`
- ✅ **DO**: `QueryIsBlacklisted()` instead of `IsBlacklisted()`
- ✅ **DO**: Use contract method names as the base: `Query` + contract method name
- ❌ **DON'T**: Remove the "Query" prefix as it makes the on-chain nature explicit

## Context Handling

### Context as Struct Member
When a client or service needs to maintain context across multiple operations:

- ✅ **DO**: Include `ctx context.Context` as a struct member
- ✅ **DO**: Accept context in constructor functions
- ✅ **DO**: Use the struct's context in all method calls

## Resource Management

### Ownership and Cleanup
When working with external resources (like `ethclient.Client`):

- ✅ **DO**: Only close resources you own
- ✅ **DO**: Don't provide `Close()` methods for resources you don't own
- ✅ **DO**: Document ownership clearly in comments

```go
// ✅ Good - No Close() method since we don't own the ethclient
type GateClient struct {
    ctx        context.Context
    client     *ethclient.Client  // We don't own this
    gateCaller *perp_gate.GateCaller
    address    common.Address
}

// ❌ Bad - Closing a resource we don't own
func (q *GateClient) Close() {
    q.client.Close()  // This closes someone else's client!
}
```

```go
// ✅ Good - Context as member
type GateClient struct {
    ctx        context.Context
    client     *ethclient.Client
    gateCaller *perp_gate.GateCaller
    address    common.Address
}

func NewGateClient(ctx context.Context, client *ethclient.Client, gateAddress common.Address) (*GateClient, error) {
    // ...
}

func (q *GateClient) QueryFullGateState() (*types.GateState, error) {
    // Use q.ctx instead of passing context as parameter
    weth, err := q.gateCaller.Weth(&bind.CallOpts{Context: q.ctx})
    // ...
}
```

## Import Organization

### Standard Library First
```go
import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/SynFutures/go-libs/contracts/perp/gate"
    "github.com/SynFutures/go-snapshot/types"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "go.uber.org/zap"
)
```

## Error Handling

### Consistent Error Messages
- Use descriptive error messages that include context
- Include relevant identifiers (addresses, indices) in error messages

```go
// ✅ Good
log.ErrorZ(q.ctx, "Failed to query reserve balance",
    zap.Error(err),
    zap.String("quote", strings.ToLower(quote.Hex())),
    zap.String("trader", strings.ToLower(trader.Hex())))

// ❌ Bad
log.ErrorZ(q.ctx, "Query failed", zap.Error(err))
```

## Testing

### Test Naming
- Use descriptive test names that indicate what is being tested
- ✅ **DO**: `TestGateClient_NewGateClient()` instead of `TestGateRPCQuerier_NewGateRPCQuerier()`

### Test Structure
- Include context setup in tests
- Use descriptive variable names (`gateClient` instead of `client` when appropriate)

```go
func TestGateClient_NewGateClient(t *testing.T) {
    ctx := context.Background()
    
    // Test with invalid RPC URL using the convenience function
    _, err := NewGateClientWithURL(ctx, "invalid://url", common.Address{})
    if err == nil {
        t.Error("Expected error for invalid RPC URL, got nil")
    }
}
```

## Documentation

### README Files
- Use descriptive names that match the main functionality
- ✅ **DO**: `README_GATE_CLIENT.md` instead of `README_GATE_RPC_QUERY.md`

### Code Comments
- Use clear, descriptive comments that explain the purpose
- Include examples when helpful

```go
// GateClient provides methods to query Gate contract state directly on-chain
type GateClient struct {
    ctx        context.Context
    client     *ethclient.Client
    gateCaller *perp_gate.GateCaller
    address    common.Address
}
```

## Enforcement

These rules should be enforced through:
1. Code reviews
2. Linting tools (where applicable)
3. Team discussions and consensus
4. Regular updates to this document as patterns evolve

## Exceptions

Exceptions to these rules may be made when:
1. Following established patterns in external dependencies
2. Maintaining compatibility with existing APIs
3. Performance requirements necessitate different approaches

All exceptions should be documented with clear rationale.
