package erc8004

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ---------------------------
// Mock Contract Backend (satisfies bind.ContractBackend on recent geth)
// ---------------------------

type mockBackend struct {
	abi  abi.ABI
	data struct {
		count     *big.Int
		agents    map[int64]agentInfoTuple
		tokenURIs map[int64]string
		owners    map[int64]common.Address
		metadata  map[int64]map[string][]byte
		exists    map[int64]bool
	}
}

// --- ContractCaller ---

func (m *mockBackend) CodeAt(ctx context.Context, _ common.Address, _ *big.Int) ([]byte, error) {
	return nil, nil
}
func (m *mockBackend) PendingCodeAt(ctx context.Context, _ common.Address) ([]byte, error) {
	return nil, nil
}

func (m *mockBackend) CallContract(ctx context.Context, call ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	if len(call.Data) < 4 {
		return nil, nil
	}
	sig := call.Data[:4]
	for name, method := range m.abi.Methods {
		if strings.EqualFold(common.Bytes2Hex(method.ID), common.Bytes2Hex(sig)) {
			// decode inputs if present
			var args []interface{}
			if len(call.Data) > 4 {
				in, err := method.Inputs.Unpack(call.Data[4:])
				if err != nil {
					return nil, err
				}
				args = in
			}
			switch name {
			case "totalAgents":
				return method.Outputs.Pack(m.data.count)

			case "agentExists":
				if len(args) != 1 {
					return nil, nil
				}
				id, _ := args[0].(*big.Int)
				exists := m.data.exists[id.Int64()]
				return method.Outputs.Pack(exists)

			case "tokenURI":
				if len(args) != 1 {
					return nil, nil
				}
				id, _ := args[0].(*big.Int)
				uri := m.data.tokenURIs[id.Int64()]
				return method.Outputs.Pack(uri)

			case "ownerOf":
				if len(args) != 1 {
					return nil, nil
				}
				id, _ := args[0].(*big.Int)
				owner := m.data.owners[id.Int64()]
				return method.Outputs.Pack(owner)

			case "getMetadata":
				if len(args) != 2 {
					return nil, nil
				}
				id, _ := args[0].(*big.Int)
				key, _ := args[1].(string)
				value := m.data.metadata[id.Int64()][key]
				return method.Outputs.Pack(value)
			}
		}
	}
	return nil, nil
}

// --- ContractTransactor ---

func (m *mockBackend) PendingNonceAt(ctx context.Context, _ common.Address) (uint64, error) {
	return 0, nil
}
func (m *mockBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
}
func (m *mockBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
}
func (m *mockBackend) EstimateGas(ctx context.Context, _ ethereum.CallMsg) (uint64, error) {
	return 21_000, nil
}
func (m *mockBackend) SendTransaction(ctx context.Context, _ *types.Transaction) error { return nil }

// --- ContractFilterer ---

func (m *mockBackend) FilterLogs(ctx context.Context, _ ethereum.FilterQuery) ([]types.Log, error) {
	return nil, nil
}
func (m *mockBackend) SubscribeFilterLogs(ctx context.Context, _ ethereum.FilterQuery, _ chan<- types.Log) (ethereum.Subscription, error) {
	return dummySub{}, nil
}

type dummySub struct{}

func (d dummySub) Unsubscribe()      {}
func (d dummySub) Err() <-chan error { ch := make(chan error); close(ch); return ch }

// --- ChainReader (newer geth requires this via bind.ContractBackend) ---

func (m *mockBackend) HeaderByNumber(ctx context.Context, _ *big.Int) (*types.Header, error) {
	return &types.Header{Number: big.NewInt(0)}, nil
}
func (m *mockBackend) HeaderByHash(ctx context.Context, _ common.Hash) (*types.Header, error) {
	return &types.Header{Number: big.NewInt(0)}, nil
}
func (m *mockBackend) BlockByNumber(ctx context.Context, _ *big.Int) (*types.Block, error) {
	return nil, nil
}
func (m *mockBackend) BlockByHash(ctx context.Context, _ common.Hash) (*types.Block, error) {
	return nil, nil
}

func (m *mockBackend) BalanceAt(ctx context.Context, _ common.Address, _ *big.Int) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (m *mockBackend) StorageAt(ctx context.Context, _ common.Address, _ common.Hash, _ *big.Int) ([]byte, error) {
	return nil, nil
}
func (m *mockBackend) TransactionCount(ctx context.Context, _ common.Hash) (uint, error) {
	return 0, nil
}
func (m *mockBackend) TransactionInBlock(ctx context.Context, _ common.Hash, _ uint) (*types.Transaction, error) {
	return nil, nil
}
func (m *mockBackend) PendingBalanceAt(ctx context.Context, _ common.Address) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (m *mockBackend) PendingStorageAt(ctx context.Context, _ common.Address, _ common.Hash) ([]byte, error) {
	return nil, nil
}
func (m *mockBackend) PendingTransactionCount(ctx context.Context) (uint, error) { return 0, nil }
func (m *mockBackend) SubscribeNewHead(ctx context.Context, _ chan<- *types.Header) (ethereum.Subscription, error) {
	return dummySub{}, nil
}

// ---------------------------
// Test Data Helpers
// ---------------------------

func newMockBackend(t *testing.T) *mockBackend {
	t.Helper()
	parsed, err := abi.JSON(strings.NewReader(IdentityRegistryABI))
	if err != nil {
		t.Fatalf("abi parse: %v", err)
	}
	m := &mockBackend{abi: parsed}
	m.data.count = big.NewInt(3)
	m.data.agents = map[int64]agentInfoTuple{}
	m.data.tokenURIs = map[int64]string{}
	m.data.owners = map[int64]common.Address{}
	m.data.metadata = map[int64]map[string][]byte{}
	m.data.exists = map[int64]bool{}

	// Set up test data
	agents := []struct {
		id       int64
		tokenURI string
		owner    common.Address
		domain   string
	}{
		{1, "alpha.example", common.HexToAddress("0x00000000000000000000000000000000000000a1"), "alpha.example"},
		{2, "beta.example", common.HexToAddress("0x00000000000000000000000000000000000000b2"), "beta.example"},
		{3, "gamma.example", common.HexToAddress("0x00000000000000000000000000000000000000c3"), "gamma.example"},
	}

	for _, a := range agents {
		m.data.exists[a.id] = true
		m.data.tokenURIs[a.id] = a.tokenURI
		m.data.owners[a.id] = a.owner
		m.data.metadata[a.id] = map[string][]byte{
			"domain": []byte(a.domain),
		}
		m.data.agents[a.id] = agentInfoTuple{
			AgentId:  big.NewInt(a.id),
			TokenURI: a.tokenURI,
			Owner:    a.owner,
		}
	}
	return m
}

// ---------------------------
//
// Tests
//
// ---------------------------

func TestIdentity_GetAgent(t *testing.T) {
	m := newMockBackend(t)
	id, err := NewIdentity(common.HexToAddress("0x000000000000000000000000000000000000dead"), m)
	if err != nil {
		t.Fatalf("new identity: %v", err)
	}

	ctx := context.Background()
	for i := int64(1); i <= 3; i++ {
		got, err := id.GetAgent(ctx, nil, big.NewInt(i))
		if err != nil {
			t.Fatalf("GetAgent(%d) error: %v", i, err)
		}
		if got.AgentId.Int64() != i {
			t.Fatalf("GetAgent(%d) AgentId mismatch: got %d", i, got.AgentId.Int64())
		}
		wantURI := []string{"alpha.example", "beta.example", "gamma.example"}[i-1]
		if got.TokenURI != wantURI {
			t.Fatalf("GetAgent(%d) TokenURI mismatch: got %q want %q", i, got.TokenURI, wantURI)
		}
	}
}

func TestIdentity_TotalAgents(t *testing.T) {
	m := newMockBackend(t)
	id, err := NewIdentity(common.HexToAddress("0x000000000000000000000000000000000000dead"), m)
	if err != nil {
		t.Fatalf("new identity: %v", err)
	}
	got, err := id.TotalAgents(context.Background(), nil)
	if err != nil {
		t.Fatalf("TotalAgents error: %v", err)
	}
	if got.Cmp(m.data.count) != 0 {
		t.Fatalf("TotalAgents mismatch: got %s want %s", got, m.data.count)
	}
}

func TestIdentity_AgentExists(t *testing.T) {
	m := newMockBackend(t)
	id, err := NewIdentity(common.HexToAddress("0x000000000000000000000000000000000000dead"), m)
	if err != nil {
		t.Fatalf("new identity: %v", err)
	}

	// Test existing agents
	for i := int64(1); i <= 3; i++ {
		exists, err := id.AgentExists(context.Background(), nil, big.NewInt(i))
		if err != nil {
			t.Fatalf("AgentExists(%d) error: %v", i, err)
		}
		if !exists {
			t.Fatalf("AgentExists(%d) should be true", i)
		}
	}

	// Test non-existing agent
	exists, err := id.AgentExists(context.Background(), nil, big.NewInt(999))
	if err != nil {
		t.Fatalf("AgentExists(999) error: %v", err)
	}
	if exists {
		t.Fatalf("AgentExists(999) should be false")
	}
}

func TestIdentity_GetMetadata(t *testing.T) {
	m := newMockBackend(t)
	id, err := NewIdentity(common.HexToAddress("0x000000000000000000000000000000000000dead"), m)
	if err != nil {
		t.Fatalf("new identity: %v", err)
	}

	domain, err := id.GetMetadata(context.Background(), nil, big.NewInt(1), "domain")
	if err != nil {
		t.Fatalf("GetMetadata error: %v", err)
	}
	if string(domain) != "alpha.example" {
		t.Fatalf("GetMetadata mismatch: got %s want alpha.example", string(domain))
	}
}
