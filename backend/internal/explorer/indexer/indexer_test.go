package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	erc "github.com/praxis/praxis-explorer/internal/erc8004"
)

// helper: compute topic for uint256 id
func topicForUint256(v *big.Int) common.Hash {
	// Topics for indexed values are ABI-encoded as 32-byte right-padded big-endian
	var b [32]byte
	copy(b[32-len(v.Bytes()):], v.Bytes())
	return common.BytesToHash(b[:])
}

// helper: address -> indexed topic (left-pad address to 32 bytes)
func topicForAddress(a common.Address) common.Hash {
	var b [32]byte
	copy(b[12:], a.Bytes())
	return common.BytesToHash(b[:])
}

func TestStart_HandlesAgentRegisteredEventAndFetchesCard(t *testing.T) {
	t.Helper()

	// ---- 1) Spin up a fake agent-card host and record hits
	var hits int32
	card := map[string]any{
		"name": "unit-test-agent",
	}
	// Serve /.well-known/agent-card.json, count requests
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/.well-known/agent-card.json") {
			atomic.AddInt32(&hits, 1)
			_ = json.NewEncoder(w).Encode(card)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	// Domain we encode into the event. Using the full server URL ensures the indexer
	// appends "/.well-known/agent-card.json" (since it starts with http://).
	domain := srv.URL

	// ---- 2) Build an indexer with no networks (we only test event handling path)
	ix := &Indexer{
		// store left as nil; we only assert HTTP fetch was attempted.
		nets:    []Chain{},
		seeds:   []string{},
		clients: make(map[string]*ethclient.Client),
		idents:  map[string]common.Address{"sepolia": common.HexToAddress("0x1111111111111111111111111111111111111111")},
	}
	// Parse ABI same way the real constructor does
	parsed, err := abi.JSON(strings.NewReader(erc.IdentityABI()))
	if err != nil {
		t.Fatalf("parse identity ABI: %v", err)
	}
	ix.idABI = parsed

	// ---- 3) Start the background loop (no nets, so it will idle)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		ix.Start(ctx) // will loop until canceled
		close(done)
	}()

	// ---- 4) Craft a mock AgentRegistered log and feed it to the indexer
	ev := ix.idABI.Events["AgentRegistered"]
	agentID := big.NewInt(42)
	agentAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")

	// Pack data section: non-indexed args -> (agentDomain string, agentAddress address)
	data, err := ev.Inputs.NonIndexed().Pack(domain, agentAddr)
	if err != nil {
		t.Fatalf("pack event data: %v", err)
	}

	lg := types.Log{
		Address: common.HexToAddress("0xeFbcfaB3547EF997A747FeA1fCfBBb2fd3912445"),
		Topics: []common.Hash{
			ev.ID,                    // event signature
			topicForUint256(agentID), // indexed agentId
		},
		Data:        data,
		BlockNumber: 123,
		TxHash:      common.HexToHash(randomHash("txhash")),
		Index:       0,
	}

	// Call the handler directly to simulate the on-chain event arriving on the subscription.
	// fetchAndStoreCard will attempt to GET the agent-card from our test server.
	func() {
		defer func() {
			// If store is nil and UpsertAgentFromCard is called, a panic would occur.
			// We recover to keep the test focused on "did we fetch the card URL?"
			_ = recover()
		}()
		ix.handleIdentityLog(ctx, "sepolia", lg)
	}()

	// ---- 5) Assert the card was fetched (i.e., the indexer tried to track the agent)
	waitUntil(t, 2*time.Second, func() bool {
		return atomic.LoadInt32(&hits) >= 1
	})

	// Clean up the goroutine
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("indexer did not stop in time")
	}
}

func TestStart_HandlesV1RegisteredEvent_ParsesRegistrationAndFetchesCard(t *testing.T) {
	t.Helper()

	// 1) Fake server serving both the registration JSON and the agent-card.json
	var regHits, cardHits int32
	reg := map[string]any{
		"type": "https://eips.ethereum.org/EIPS/eip-8004#registration-v1",
		"endpoints": []any{
			map[string]any{"name": "A2A", "endpoint": "", "version": "0.3.0"},
			map[string]any{"name": "MCP", "endpoint": "https://mcp.example/", "version": "2025-06-18"},
			map[string]any{"name": "DID", "endpoint": "did:method:foobar", "version": "v1"},
		},
	}
	card := map[string]any{"name": "unit-test-agent-v1"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/reg.json"):
			atomic.AddInt32(&regHits, 1)
			_ = json.NewEncoder(w).Encode(reg)
			return
		case strings.HasSuffix(r.URL.Path, "/.well-known/agent-card.json"):
			atomic.AddInt32(&cardHits, 1)
			_ = json.NewEncoder(w).Encode(card)
			return
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// Now that we know the server URL, set the A2A endpoint to it.
	// Indexer will append "/.well-known/agent-card.json".
	reg["endpoints"].([]any)[0].(map[string]any)["endpoint"] = srv.URL

	// 2) Build an indexer with the ABI (which includes v1 "Registered")
	ix := &Indexer{
		nets:    []Chain{},
		seeds:   []string{},
		clients: make(map[string]*ethclient.Client),
		idents:  map[string]common.Address{"sepolia": common.HexToAddress("0x1111111111111111111111111111111111111111")},
	}
	parsed, err := abi.JSON(strings.NewReader(erc.IdentityABI()))
	if err != nil {
		t.Fatalf("parse identity ABI: %v", err)
	}
	ix.idABI = parsed

	// 3) Start the background loop (idle)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		ix.Start(ctx)
		close(done)
	}()

	// 4) Create a v1 Registered log
	ev := ix.idABI.Events["Registered"]
	agentID := big.NewInt(77)
	owner := common.HexToAddress("0x3333333333333333333333333333333333333333")

	// Data: NonIndexed (tokenURI)
	tokenURI := srv.URL + "/reg.json"
	data, err := ev.Inputs.NonIndexed().Pack(tokenURI)
	if err != nil {
		t.Fatalf("pack v1 event data: %v", err)
	}

	lg := types.Log{
		Address: common.HexToAddress("0xeFbcfaB3547EF997A747FeA1fCfBBb2fd3912445"),
		Topics: []common.Hash{
			ev.ID,
			topicForUint256(agentID), // indexed agentId
			topicForAddress(owner),   // indexed owner
		},
		Data:        data,
		BlockNumber: 456,
		TxHash:      common.HexToHash(randomHash("txhash-v1")),
		Index:       0,
	}

	// 5) Trigger the handler; it should fetch the registration, discover A2A, then fetch the card
	func() {
		defer func() { _ = recover() }() // ignore nil store upsert panic
		ix.handleIdentityLog(ctx, "sepolia", lg)
	}()

	// Assert both registration and card were fetched
	waitUntil(t, 2*time.Second, func() bool { return atomic.LoadInt32(&regHits) >= 1 })
	waitUntil(t, 2*time.Second, func() bool { return atomic.LoadInt32(&cardHits) >= 1 })

	// Cleanup
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("indexer did not stop in time")
	}
}

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

func randomHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
