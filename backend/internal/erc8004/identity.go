package erc8004

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

type AgentInfo struct {
	AgentId  *big.Int
	TokenURI string
	Owner    common.Address
}

// internal tuple type used for ABI decoding
type agentInfoTuple struct {
	AgentId  *big.Int
	TokenURI string
	Owner    common.Address
}

type Identity struct {
	addr     common.Address
	backend  bind.ContractBackend
	contract *bind.BoundContract
	abi      abi.ABI
}

func NewIdentity(addr common.Address, backend bind.ContractBackend) (*Identity, error) {
	parsed, err := abi.JSON(strings.NewReader(IdentityRegistryABI))
	if err != nil {
		return nil, err
	}
	c := bind.NewBoundContract(addr, parsed, backend, backend, backend)
	return &Identity{addr: addr, backend: backend, contract: c, abi: parsed}, nil
}

// IdentityABI returns the ABI JSON used by this package (helper for indexer)
func IdentityABI() string { return IdentityRegistryABI }

func (i *Identity) Register(auth *bind.TransactOpts) (*big.Int, error) {
	_, err := i.contract.Transact(auth, "register")
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Identity) RegisterWithURI(auth *bind.TransactOpts, tokenURI string) (*big.Int, error) {
	_, err := i.contract.Transact(auth, "register", tokenURI)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Identity) SetMetadata(auth *bind.TransactOpts, agentId *big.Int, key string, value []byte) error {
	_, err := i.contract.Transact(auth, "setMetadata", agentId, key, value)
	return err
}

func (i *Identity) GetMetadata(ctx context.Context, call *bind.CallOpts, agentId *big.Int, key string) ([]byte, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	var result []byte
	out := []interface{}{&result}
	if err := i.contract.Call(call, &out, "getMetadata", agentId, key); err != nil {
		return nil, err
	}
	return result, nil
}

func (i *Identity) TokenURI(ctx context.Context, call *bind.CallOpts, tokenId *big.Int) (string, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	var uri string
	out := []interface{}{&uri}
	if err := i.contract.Call(call, &out, "tokenURI", tokenId); err != nil {
		return "", err
	}
	return uri, nil
}

func (i *Identity) OwnerOf(ctx context.Context, call *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	var owner common.Address
	out := []interface{}{&owner}
	if err := i.contract.Call(call, &out, "ownerOf", tokenId); err != nil {
		return common.Address{}, err
	}
	return owner, nil
}

func (i *Identity) AgentExists(ctx context.Context, call *bind.CallOpts, agentId *big.Int) (bool, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	var exists bool
	out := []interface{}{&exists}
	if err := i.contract.Call(call, &out, "agentExists", agentId); err != nil {
		return false, err
	}
	return exists, nil
}

func (i *Identity) TotalAgents(ctx context.Context, call *bind.CallOpts) (*big.Int, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}
	var cnt *big.Int
	out := []interface{}{&cnt}
	if err := i.contract.Call(call, &out, "totalAgents"); err != nil {
		log.WithError(err).Error("totalAgents call failed")
		return nil, err
	}
	log.WithField("count", cnt).Info("totalAgents ok")
	return cnt, nil
}

// GetAgent gets basic agent info by ID (tokenId, tokenURI, owner)
func (i *Identity) GetAgent(ctx context.Context, call *bind.CallOpts, id *big.Int) (AgentInfo, error) {
	if call == nil {
		call = &bind.CallOpts{}
	}

	// Check if agent exists first
	exists, err := i.AgentExists(ctx, call, id)
	if err != nil {
		return AgentInfo{}, err
	}
	if !exists {
		return AgentInfo{}, fmt.Errorf("agent %s does not exist", id.String())
	}

	// Get tokenURI
	tokenURI, err := i.TokenURI(ctx, call, id)
	if err != nil {
		return AgentInfo{}, err
	}

	// Get owner
	owner, err := i.OwnerOf(ctx, call, id)
	if err != nil {
		return AgentInfo{}, err
	}

	return AgentInfo{
		AgentId:  id,
		TokenURI: tokenURI,
		Owner:    owner,
	}, nil
}
