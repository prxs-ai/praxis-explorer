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
		count  *big.Int
		byID   map[int64]agentInfoTuple
		byDom  map[string]agentInfoTuple
		byAddr map[common.Address]agentInfoTuple
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
			case "getAgentCount":
				return method.Outputs.Pack(m.data.count)

			case "getAgent":
				if len(args) != 1 {
					return nil, nil
				}
				id, _ := args[0].(*big.Int)
				info, ok := m.data.byID[id.Int64()]
				if !ok {
					info = agentInfoTuple{AgentId: big.NewInt(0)}
				}
				return method.Outputs.Pack(info)

			case "resolveByDomain":
				if len(args) != 1 {
					return nil, nil
				}
				domain, _ := args[0].(string)
				info, ok := m.data.byDom[domain]
				if !ok {
					info = agentInfoTuple{AgentId: big.NewInt(0)}
				}
				return method.Outputs.Pack(info)

			case "resolveByAddress":
				if len(args) != 1 {
					return nil, nil
				}
				addr, _ := args[0].(common.Address)
				info, ok := m.data.byAddr[addr]
				if !ok {
					info = agentInfoTuple{AgentId: big.NewInt(0)}
				}
				return method.Outputs.Pack(info)
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
	parsed, err := abi.JSON(strings.NewReader(identityABI))
	if err != nil {
		t.Fatalf("abi parse: %v", err)
	}
	m := &mockBackend{abi: parsed}
	m.data.count = big.NewInt(3)
	m.data.byID = map[int64]agentInfoTuple{}
	m.data.byDom = map[string]agentInfoTuple{}
	m.data.byAddr = map[common.Address]agentInfoTuple{}

	agents := []agentInfoTuple{
		{AgentId: big.NewInt(1), AgentDomain: "alpha.example", AgentAddress: common.HexToAddress("0x00000000000000000000000000000000000000a1")},
		{AgentId: big.NewInt(2), AgentDomain: "beta.example", AgentAddress: common.HexToAddress("0x00000000000000000000000000000000000000b2")},
		{AgentId: big.NewInt(3), AgentDomain: "gamma.example", AgentAddress: common.HexToAddress("0x00000000000000000000000000000000000000c3")},
	}
	for _, a := range agents {
		m.data.byID[a.AgentId.Int64()] = a
		m.data.byDom[a.AgentDomain] = a
		m.data.byAddr[a.AgentAddress] = a
	}
	return m
}

// ---------------------------
//
// Tests
//
// ---------------------------

func TestIdentity_TupleDecoding_GetAgent(t *testing.T) {
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
		wantDom := []string{"alpha.example", "beta.example", "gamma.example"}[i-1]
		if got.AgentDomain != wantDom {
			t.Fatalf("GetAgent(%d) AgentDomain mismatch: got %q want %q", i, got.AgentDomain, wantDom)
		}
	}
}

func TestIdentity_TupleDecoding_ResolveByDomain(t *testing.T) {
	m := newMockBackend(t)
	id, err := NewIdentity(common.HexToAddress("0x000000000000000000000000000000000000dead"), m)
	if err != nil {
		t.Fatalf("new identity: %v", err)
	}

	cases := map[string]int64{
		"alpha.example": 1,
		"beta.example":  2,
		"gamma.example": 3,
	}
	for dom, wantID := range cases {
		got, err := id.ResolveByDomain(context.Background(), nil, dom)
		if err != nil {
			t.Fatalf("ResolveByDomain(%q) error: %v", dom, err)
		}
		if got.AgentId.Int64() != wantID {
			t.Fatalf("ResolveByDomain(%q) AgentId mismatch: got %d want %d", dom, got.AgentId.Int64(), wantID)
		}
		if got.AgentDomain != dom {
			t.Fatalf("ResolveByDomain(%q) AgentDomain mismatch: got %q", dom, got.AgentDomain)
		}
	}
}

func TestIdentity_TupleDecoding_ResolveByAddress(t *testing.T) {
	m := newMockBackend(t)
	id, err := NewIdentity(common.HexToAddress("0x000000000000000000000000000000000000dead"), m)
	if err != nil {
		t.Fatalf("new identity: %v", err)
	}
	for addr, want := range m.data.byAddr {
		got, err := id.ResolveByAddress(context.Background(), nil, addr)
		if err != nil {
			t.Fatalf("ResolveByAddress(%s) error: %v", addr.Hex(), err)
		}
		if got.AgentId.Cmp(want.AgentId) != 0 {
			t.Fatalf("ResolveByAddress(%s) AgentId mismatch: got %s want %s", addr.Hex(), got.AgentId, want.AgentId)
		}
		if got.AgentAddress != addr {
			t.Fatalf("ResolveByAddress(%s) AgentAddress mismatch: got %s", addr.Hex(), got.AgentAddress.Hex())
		}
	}
}

func TestIdentity_GetAgentCount(t *testing.T) {
	m := newMockBackend(t)
	id, err := NewIdentity(common.HexToAddress("0x000000000000000000000000000000000000dead"), m)
	if err != nil {
		t.Fatalf("new identity: %v", err)
	}
	got, err := id.GetAgentCount(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetAgentCount error: %v", err)
	}
	if got.Cmp(m.data.count) != 0 {
		t.Fatalf("GetAgentCount mismatch: got %s want %s", got, m.data.count)
	}
}
