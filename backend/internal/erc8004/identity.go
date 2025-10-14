package erc8004

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

type AgentInfo struct {
	AgentId      *big.Int
	AgentDomain  string
	AgentAddress common.Address
}

// internal tuple type used for ABI decoding of AgentInfo
type agentInfoTuple struct {
	AgentId      *big.Int
	AgentDomain  string
	AgentAddress common.Address
}

// ABI for IdentityRegistry with events (from reference implementation)
const identityABI = `[
  {"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"agentId","type":"uint256"},{"indexed":false,"internalType":"string","name":"tokenURI","type":"string"},{"indexed":true,"internalType":"address","name":"owner","type":"address"}],"name":"Registered","type":"event"},
  {"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"agentId","type":"uint256"},{"indexed":false,"internalType":"string","name":"agentDomain","type":"string"},{"indexed":false,"internalType":"address","name":"agentAddress","type":"address"}],"name":"AgentRegistered","type":"event"},
  {"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"agentId","type":"uint256"},{"indexed":false,"internalType":"string","name":"agentDomain","type":"string"},{"indexed":false,"internalType":"address","name":"agentAddress","type":"address"}],"name":"AgentUpdated","type":"event"},
  {"inputs":[{"internalType":"string","name":"agentDomain","type":"string"},{"internalType":"address","name":"agentAddress","type":"address"}],"name":"newAgent","outputs":[{"internalType":"uint256","name":"agentId","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"agentId","type":"uint256"},{"internalType":"string","name":"newAgentDomain","type":"string"},{"internalType":"address","name":"newAgentAddress","type":"address"}],"name":"updateAgent","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"agentId","type":"uint256"}],"name":"getAgent","outputs":[{"components":[{"internalType":"uint256","name":"agentId","type":"uint256"},{"internalType":"string","name":"agentDomain","type":"string"},{"internalType":"address","name":"agentAddress","type":"address"}],"internalType":"struct IIdentityRegistry.AgentInfo","name":"agentInfo","type":"tuple"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"string","name":"agentDomain","type":"string"}],"name":"resolveByDomain","outputs":[{"components":[{"internalType":"uint256","name":"agentId","type":"uint256"},{"internalType":"string","name":"agentDomain","type":"string"},{"internalType":"address","name":"agentAddress","type":"address"}],"internalType":"struct IIdentityRegistry.AgentInfo","name":"agentInfo","type":"tuple"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"address","name":"agentAddress","type":"address"}],"name":"resolveByAddress","outputs":[{"components":[{"internalType":"uint256","name":"agentId","type":"uint256"},{"internalType":"string","name":"agentDomain","type":"string"},{"internalType":"address","name":"agentAddress","type":"address"}],"internalType":"struct IIdentityRegistry.AgentInfo","name":"agentInfo","type":"tuple"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"getAgentCount","outputs":[{"internalType":"uint256","name":"count","type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"agentId","type":"uint256"}],"name":"agentExists","outputs":[{"internalType":"bool","name":"exists","type":"bool"}],"stateMutability":"view","type":"function"}
]`

type Identity struct {
	addr     common.Address
	backend  bind.ContractBackend
	contract *bind.BoundContract
	abi      abi.ABI
}

func NewIdentity(addr common.Address, backend bind.ContractBackend) (*Identity, error) {
	parsed, err := abi.JSON(strings.NewReader(identityABI))
	if err != nil {
		return nil, err
	}
	c := bind.NewBoundContract(addr, parsed, backend, backend, backend)
	return &Identity{addr: addr, backend: backend, contract: c, abi: parsed}, nil
}

// IdentityABI returns the ABI JSON used by this package (helper for indexer)
func IdentityABI() string { return identityABI }

func (i *Identity) NewAgent(auth *bind.TransactOpts, domain string, addr common.Address) (*big.Int, error) {
	_, err := i.contract.Transact(auth, "newAgent", domain, addr)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Identity) UpdateAgent(auth *bind.TransactOpts, id *big.Int, domain string, addr common.Address) (bool, error) {
	_, err := i.contract.Transact(auth, "updateAgent", id, domain, addr)
	return err == nil, err
}

/*** ---------- Robust tuple readers (raw eth_call + abi.Unpack) ---------- ***/

// normalizeAgentTuple handles map, slice, and struct tuple shapes.
func normalizeAgentTuple(v interface{}) (agentInfoTuple, error) {
	var res agentInfoTuple

	// Map decoding path
	if m, ok := v.(map[string]interface{}); ok {
		// agentId
		if id, ok := m["agentId"]; ok {
			switch x := id.(type) {
			case *big.Int:
				res.AgentId = x
			case big.Int:
				res.AgentId = new(big.Int).Set(&x)
			default:
				return res, fmt.Errorf("unexpected agentId type: %T", id)
			}
		} else if id, ok := m["0"]; ok {
			switch x := id.(type) {
			case *big.Int:
				res.AgentId = x
			case big.Int:
				res.AgentId = new(big.Int).Set(&x)
			default:
				return res, fmt.Errorf("unexpected agentId(0) type: %T", id)
			}
		}
		// agentDomain
		if d, ok := m["agentDomain"]; ok {
			s, ok := d.(string)
			if !ok {
				return res, fmt.Errorf("unexpected agentDomain type: %T", d)
			}
			res.AgentDomain = s
		} else if d, ok := m["1"]; ok {
			s, ok := d.(string)
			if !ok {
				return res, fmt.Errorf("unexpected agentDomain(1) type: %T", d)
			}
			res.AgentDomain = s
		}
		// agentAddress
		if a, ok := m["agentAddress"]; ok {
			ad, ok := a.(common.Address)
			if !ok {
				return res, fmt.Errorf("unexpected agentAddress type: %T", a)
			}
			res.AgentAddress = ad
		} else if a, ok := m["2"]; ok {
			ad, ok := a.(common.Address)
			if !ok {
				return res, fmt.Errorf("unexpected agentAddress(2) type: %T", a)
			}
			res.AgentAddress = ad
		}

		if res.AgentId == nil || res.AgentDomain == "" {
			return res, fmt.Errorf("incomplete tuple (map) decoded: %+v", m)
		}
		return res, nil
	}

	// Slice positional decoding path
	if arr, ok := v.([]interface{}); ok {
		if len(arr) < 3 {
			return res, fmt.Errorf("tuple len < 3: %d", len(arr))
		}
		switch x := arr[0].(type) {
		case *big.Int:
			res.AgentId = x
		case big.Int:
			res.AgentId = new(big.Int).Set(&x)
		default:
			return res, fmt.Errorf("unexpected idx0 type: %T", arr[0])
		}
		s, ok := arr[1].(string)
		if !ok {
			return res, fmt.Errorf("unexpected idx1 type: %T", arr[1])
		}
		res.AgentDomain = s
		ad, ok := arr[2].(common.Address)
		if !ok {
			return res, fmt.Errorf("unexpected idx2 type: %T", arr[2])
		}
		res.AgentAddress = ad
		return res, nil
	}

	// Struct decoding path (what your runtime is returning)
	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() == reflect.Struct {
		// AgentId
		if f := rv.FieldByName("AgentId"); f.IsValid() && f.CanInterface() {
			switch x := f.Interface().(type) {
			case *big.Int:
				res.AgentId = x
			case big.Int:
				res.AgentId = new(big.Int).Set(&x)
			default:
				return res, fmt.Errorf("unexpected AgentId field type: %T", f.Interface())
			}
		}
		// AgentDomain
		if f := rv.FieldByName("AgentDomain"); f.IsValid() && f.CanInterface() {
			if s, ok := f.Interface().(string); ok {
				res.AgentDomain = s
			} else {
				return res, fmt.Errorf("unexpected AgentDomain field type: %T", f.Interface())
			}
		}
		// AgentAddress
		if f := rv.FieldByName("AgentAddress"); f.IsValid() && f.CanInterface() {
			if ad, ok := f.Interface().(common.Address); ok {
				res.AgentAddress = ad
			} else {
				return res, fmt.Errorf("unexpected AgentAddress field type: %T", f.Interface())
			}
		}

		if res.AgentId == nil || res.AgentDomain == "" {
			return res, fmt.Errorf("incomplete tuple (struct) decoded: %+v", v)
		}
		return res, nil
	}

	return res, fmt.Errorf("unsupported tuple type: %T", v)
}

func (i *Identity) callAgentTuple(ctx context.Context, method string, args ...interface{}) (agentInfoTuple, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 1) Pack call data
	data, err := i.abi.Pack(method, args...)
	if err != nil {
		log.WithError(err).WithField("method", method).Error("abi pack failed")
		return agentInfoTuple{}, err
	}

	// 2) eth_call
	to := i.addr
	out, err := i.backend.CallContract(ctx, ethereum.CallMsg{To: &to, Data: data}, nil)
	if err != nil {
		log.WithError(err).WithField("method", method).Error("eth_call failed")
		return agentInfoTuple{}, err
	}

	// 3) Unpack outputs to []interface{} (version-agnostic)
	vals, err := i.abi.Unpack(method, out)
	if err != nil {
		log.WithError(err).WithField("method", method).Error("abi unpack (slice) failed")
		return agentInfoTuple{}, err
	}
	if len(vals) != 1 {
		err := fmt.Errorf("expected one output, got %d", len(vals))
		log.WithError(err).WithField("method", method).Error("output arity mismatch")
		return agentInfoTuple{}, err
	}

	// 4) Normalize any tuple shape to our struct
	res, err := normalizeAgentTuple(vals[0])
	if err != nil {
		log.WithError(err).WithField("method", method).Error("tuple normalize failed")
		return agentInfoTuple{}, err
	}

	log.WithFields(log.Fields{
		"method":      method,
		"agentId":     res.AgentId,
		"agentDomain": res.AgentDomain,
		"agentAddr":   res.AgentAddress.Hex(),
	}).Debug("tuple decoded")
	return res, nil
}

func (i *Identity) ResolveByDomain(ctx context.Context, call *bind.CallOpts, domain string) (AgentInfo, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	res, err := i.callAgentTuple(call.Context, "resolveByDomain", domain)
	if err != nil {
		return AgentInfo{}, err
	}
	return AgentInfo(res), nil
}

func (i *Identity) ResolveByAddress(ctx context.Context, call *bind.CallOpts, addr common.Address) (AgentInfo, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	res, err := i.callAgentTuple(call.Context, "resolveByAddress", addr)
	if err != nil {
		return AgentInfo{}, err
	}
	return AgentInfo(res), nil
}

func (i *Identity) GetAgent(ctx context.Context, call *bind.CallOpts, id *big.Int) (AgentInfo, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	res, err := i.callAgentTuple(call.Context, "getAgent", id)
	if err != nil {
		return AgentInfo{}, err
	}
	return AgentInfo(res), nil
}

func (i *Identity) GetAgentCount(ctx context.Context, call *bind.CallOpts) (*big.Int, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	var cnt *big.Int
	out := []interface{}{&cnt}
	if err := i.contract.Call(call, &out, "getAgentCount"); err != nil {
		log.WithError(err).Error("getAgentCount call failed")
		return nil, err
	}
	log.WithField("count", cnt).Info("getAgentCount ok")
	return cnt, nil
}
