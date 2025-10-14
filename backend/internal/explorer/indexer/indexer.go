package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	erc "github.com/praxis/praxis-explorer/internal/erc8004"
	"github.com/praxis/praxis-explorer/internal/explorer/store"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)
}

type Chain struct {
	Name       string
	RPC        string
	Identity   string
	Reputation string
	Validation string
}

type Indexer struct {
	store *store.Postgres
	nets  []Chain
	seeds []string // optional list of seed domains to crawl if logs are unavailable
	// runtime
	clients map[string]*ethclient.Client
	idents  map[string]common.Address
	idABI   abi.ABI
}

func New(st *store.Postgres, cfgPath string) (*Indexer, error) {
	nets, err := loadConfig(cfgPath)
	if err != nil {
		log.WithError(err).WithField("cfgPath", cfgPath).Error("failed to load config")
		return nil, err
	}

	for _, n := range nets {
		log.WithFields(log.Fields{
			"chain":      n.Name,
			"rpc":        n.RPC,
			"identity":   n.Identity,
			"reputation": n.Reputation,
			"validation": n.Validation,
		}).Info("loaded network from config")
	}

	seeds := readSeedsFromEnv()
	if len(seeds) > 0 {
		log.WithField("seeds", strings.Join(seeds, ",")).Info("loaded seed domains from env")
	}

	parsed, err := abi.JSON(strings.NewReader(erc.IdentityABI()))
	if err != nil {
		return nil, fmt.Errorf("identity abi parse: %w", err)
	}

	return &Indexer{
		store:   st,
		nets:    nets,
		seeds:   seeds,
		clients: map[string]*ethclient.Client{},
		idents:  map[string]common.Address{},
		idABI:   parsed,
	}, nil
}

func (ix *Indexer) Start(ctx context.Context) {
	log.Info("[indexer] starting background job")

	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	ix.crawlSeeds(ctx)
	ix.upgradeZeroIDs(ctx)
	ix.startOnchainWatchers(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info("[indexer] shutting down")
			return
		case <-t.C:
			ix.crawlSeeds(ctx)
			ix.upgradeZeroIDs(ctx)
		}
	}
}

func (ix *Indexer) crawlSeeds(ctx context.Context) {
	for _, domain := range ix.seeds {
		d := strings.TrimSpace(domain)
		if d == "" {
			continue
		}
		url := fmt.Sprintf("http://%s/.well-known/agent-card.json", d)
		log.WithField("url", url).Info("crawling seed")

		resp, err := http.Get(url) // #nosec G107 (operator-provided domains)
		if err != nil {
			log.WithError(err).Warn("seed fetch error")
			continue
		}
		var card map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&card); err == nil {
			// registry unknown when crawling seeds
			_ = ix.store.UpsertAgentFromCard(ctx, "sepolia", "", 0, d, card)
		}
		resp.Body.Close()
		log.WithField("domain", domain).Info("seed card fetched")
	}
}

func readSeedsFromEnv() []string {
	v := os.Getenv("EXPLORER_SEEDS")
	if v == "" {
		return []string{}
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// --- On-chain watchers ---
func (ix *Indexer) startOnchainWatchers(ctx context.Context) {
	for _, n := range ix.nets {
		if strings.TrimSpace(n.RPC) == "" || strings.TrimSpace(n.Identity) == "" {
			continue
		}
		log.WithFields(log.Fields{
			"chain": n.Name,
			"rpc":   n.RPC,
		}).Info("connecting to chain")

		client, err := ethclient.Dial(os.ExpandEnv(n.RPC))
		if err != nil {
			log.WithError(err).WithField("rpc", n.RPC).Error("failed to dial RPC")
			continue
		}
		ix.clients[n.Name] = client
		ix.idents[n.Name] = common.HexToAddress(n.Identity)
		go ix.watchIdentity(ctx, n.Name, client, ix.idents[n.Name])
		go ix.backfillAgents(ctx, n.Name, client, ix.idents[n.Name])
	}
}

func (ix *Indexer) backfillAgents(ctx context.Context, chain string, client *ethclient.Client, idAddr common.Address) {
	log.WithFields(log.Fields{
		"chain":    chain,
		"registry": idAddr.Hex(),
	}).Info("backfilling agents")

	ident, err := erc.NewIdentity(idAddr, client)
	if err != nil {
		return
	}
	count, err := ident.GetAgentCount(ctx, &bind.CallOpts{Context: ctx})
	if err != nil || count == nil {
		log.WithError(err).Error("failed to get agent count")
		return
	}
	total := count.Int64()
	if total <= 0 {
		log.WithError(err).Error("total count is loe to zero")
		return
	}

	log.WithFields(log.Fields{
		"chain": chain,
		"count": total,
	}).Info("found agents on chain")

	for i := int64(1); i <= total; i++ {
		ai, err := ident.GetAgent(ctx, &bind.CallOpts{Context: ctx}, big.NewInt(i))
		if err != nil || ai.AgentId == nil {
			log.WithError(err).Error("failed to get agent")
			continue
		}
		domain := strings.TrimSpace(ai.AgentDomain)
		if domain == "" {
			log.Error("domain is empty string")
			continue
		}
		log.WithFields(log.Fields{
			"agent_id": ai.AgentId,
			"chain":    chain,
			"count":    total,
		}).Info("storing card")
		ix.fetchAndStoreCard(ctx, chain, idAddr.Hex(), ai.AgentId.Int64(), domain)
	}
}

func (ix *Indexer) watchIdentity(ctx context.Context, chain string, client *ethclient.Client, idAddr common.Address) {
	q := ethereum.FilterQuery{Addresses: []common.Address{idAddr}}
	logsCh := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(ctx, q, logsCh)
	if err != nil {
		// Infura HTTPS and some providers will return this
		if strings.Contains(strings.ToLower(err.Error()), "notifications not supported") {
			log.WithFields(log.Fields{
				"chain": chain,
				"addr":  idAddr.Hex(),
			}).Warn("provider does not support subscriptions; falling back to polling")
			ix.pollIdentity(ctx, chain, client, idAddr) // blocking loop
			return
		}
		log.WithError(err).Error("failed to connect to the Ethereum Chain")
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-sub.Err():
			if err != nil && !errors.Is(err, context.Canceled) {
				log.WithError(err).Warn("subscription error; restarting watcher")
				time.Sleep(3 * time.Second)
				go ix.watchIdentity(ctx, chain, client, idAddr)
			}
			return
		case lg := <-logsCh:
			ix.handleIdentityLog(ctx, chain, lg)
		}
	}
}

func (ix *Indexer) pollIdentity(ctx context.Context, chain string, client *ethclient.Client, idAddr common.Address) {
	// Start from the latest block to avoid reprocessing large history
	start, err := client.BlockNumber(ctx)
	if err != nil {
		log.WithError(err).WithField("chain", chain).Error("cannot get latest block for polling")
		return
	}
	from := start // start at latest; adjust if you want history
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	log.WithFields(log.Fields{
		"chain": chain,
		"from":  from,
		"addr":  idAddr.Hex(),
	}).Info("polling identity logs (fallback)")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			latest, err := client.BlockNumber(ctx)
			if err != nil {
				log.WithError(err).WithField("chain", chain).Warn("poll: failed to fetch latest block")
				continue
			}
			if latest < from {
				log.Info("reorg or provider quirk; reset to latest")
				from = latest
			}
			if latest == from {
				log.Info("nothing new")
				continue
			}

			q := ethereum.FilterQuery{
				Addresses: []common.Address{idAddr},
				FromBlock: new(big.Int).SetUint64(from),
				ToBlock:   new(big.Int).SetUint64(latest),
			}
			logs, err := client.FilterLogs(ctx, q)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"chain": chain, "from": from, "to": latest,
				}).Warn("poll: FilterLogs error")
				continue
			}
			for _, lg := range logs {
				ix.handleIdentityLog(ctx, chain, lg)
			}
			from = latest + 1
		}
	}
}

func (ix *Indexer) handleIdentityLog(ctx context.Context, chain string, lg types.Log) {
	// Try AgentRegistered / AgentUpdated
	if len(lg.Topics) == 0 {
		log.Error("Topics are zero")
		return
	}

	log.WithFields(log.Fields{
		"chain": chain,
		"block": lg.BlockNumber,
	}).Debug("log received")

	// v1 event
	if evV1, ok := ix.idABI.Events["Registered"]; ok && lg.Topics[0] == evV1.ID {
		ix.handleRegistrationV1(ctx, chain, lg, evV1)
		return
	}

	evReg := ix.idABI.Events["AgentRegistered"]
	evUpd := ix.idABI.Events["AgentUpdated"]

	switch lg.Topics[0] {
	case evReg.ID:
		if len(lg.Topics) < 2 {
			log.Error("Number of topics is less thant two")
			return
		}
		id := new(big.Int).SetBytes(lg.Topics[1].Bytes())

		log.WithFields(log.Fields{
			"chain": chain,
			"id":    id.String(),
		}).Info("AgentRegistered event")

		var data struct {
			AgentDomain  string
			AgentAddress common.Address
		}
		if err := ix.idABI.UnpackIntoInterface(&data, "AgentRegistered", lg.Data); err != nil {
			log.WithError(err).Error("failed unpacking agent data from registered event")
			return
		}
		reg := ix.idents[chain].Hex()
		log.WithFields(log.Fields{
			"agent_id":   id.String(),
			"chain":      chain,
			"event_type": "registered",
		}).Info("storing card")
		ix.fetchAndStoreCard(ctx, chain, reg, id.Int64(), data.AgentDomain)

	case evUpd.ID:
		if len(lg.Topics) < 2 {
			log.Error("Number of topics is less thant two")
			return
		}
		id := new(big.Int).SetBytes(lg.Topics[1].Bytes())

		log.WithFields(log.Fields{
			"chain": chain,
			"id":    id.String(),
		}).Info("AgentUpdated event")

		var data struct {
			AgentDomain  string
			AgentAddress common.Address
		}
		if err := ix.idABI.UnpackIntoInterface(&data, "AgentUpdated", lg.Data); err != nil {
			log.WithError(err).Error("failed unpacking agent data from updating event")
			return
		}
		reg := ix.idents[chain].Hex()
		log.WithFields(log.Fields{
			"agent_id":   id.String(),
			"chain":      chain,
			"event_type": "updated",
		}).Info("storing card")
		ix.fetchAndStoreCard(ctx, chain, reg, id.Int64(), data.AgentDomain)
	}
}

func (ix *Indexer) handleRegistrationV1(ctx context.Context, chain string, lg types.Log, ev abi.Event) {
	// Topics: [signature, agentId (indexed), owner (indexed)]
	if len(lg.Topics) < 3 {
		log.Error("Registered v1: not enough topics")
		return
	}
	agentID := new(big.Int).SetBytes(lg.Topics[1].Bytes())
	owner := common.BytesToAddress(lg.Topics[2].Bytes()[12:]) // right-padded 32 bytes

	// Data: NonIndexed = tokenURI (string)
	nonargs := ev.Inputs.NonIndexed()
	vals, err := abi.Arguments(nonargs).Unpack(lg.Data)
	if err != nil {
		log.WithError(err).Error("v1 Registered: unpack tokenURI failed")
		return
	}
	if len(vals) != 1 {
		log.WithField("got", len(vals)).Error("v1 Registered: unexpected outputs arity")
		return
	}
	tokenURI, _ := vals[0].(string)

	log.WithFields(log.Fields{
		"chain":    chain,
		"agent_id": agentID.String(),
		"owner":    owner.Hex(),
		"tokenURI": tokenURI,
	}).Info("Registered v1 event")

	// 1) Fetch registration JSON
	reg, err := ix.fetchJSON(ctx, tokenURI)
	if err != nil {
		log.WithError(err).WithField("tokenURI", tokenURI).Warn("registration fetch error")
		return
	}

	// 2) Parse endpoints (A2A, MCP, DID)
	a2aURL, mcpURL, did := extractEndpoints(reg)
	if a2aURL == "" {
		log.WithFields(log.Fields{
			"agent_id": agentID.String(),
			"chain":    chain,
			"mcp":      mcpURL,
			"did":      did,
		}).Warn("v1 registration has no A2A endpoint; skipping card fetch")
		return
	}

	// 3) Fetch agent card via existing path
	registry := ix.idents[chain].Hex()
	ix.fetchAndStoreCard(ctx, chain, registry, agentID.Int64(), a2aURL)
}

// Helper: fetch arbitrary JSON (supports http(s) and ipfs://)
func (ix *Indexer) fetchJSON(ctx context.Context, uri string) (map[string]any, error) {
	u := strings.TrimSpace(uri)
	if u == "" {
		return nil, fmt.Errorf("empty uri")
	}

	// Support ipfs://CID[/path]
	if strings.HasPrefix(u, "ipfs://") {
		// naive gateway transform; can be made configurable
		u = "https://ipfs.io/ipfs/" + strings.TrimPrefix(u, "ipfs://")
	}

	// Validate URL
	if _, err := url.Parse(u); err != nil {
		return nil, fmt.Errorf("invalid uri: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req) // #nosec G107
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// Helper: extract A2A, MCP, DID from a v1 registration JSON
func extractEndpoints(reg map[string]any) (a2a, mcp, did string) {
	eps, ok := reg["endpoints"]
	if !ok {
		return "", "", ""
	}
	arr, ok := eps.([]interface{})
	if !ok {
		return "", "", ""
	}
	for _, raw := range arr {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		endpoint, _ := m["endpoint"].(string)

		switch strings.ToUpper(strings.TrimSpace(name)) {
		case "A2A":
			a2a = endpoint
		case "MCP":
			mcp = endpoint
		case "DID":
			did = endpoint
		}
	}
	return
}

func (ix *Indexer) fetchAndStoreCard(ctx context.Context, chain string, registryAddr string, agentID int64, domain string) {
	d := strings.TrimSpace(domain)
	if d == "" {
		log.Error("Domain is zero")
		return
	}
	// heuristic: build .well-known URL if needed
	url := d
	if !strings.HasPrefix(d, "http://") && !strings.HasPrefix(d, "https://") {
		url = fmt.Sprintf("http://%s/.well-known/agent-card.json", d)
	} else if !strings.Contains(d, "/.well-known/agent-card.json") {
		url = strings.TrimRight(d, "/") + "/.well-known/agent-card.json"
	}

	log.WithFields(log.Fields{
		"chain":   chain,
		"agentID": agentID,
		"url":     url,
	}).Info("fetching agent card")

	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		log.WithError(err).WithField("url", url).Warn("card fetch error")
		return
	}
	defer resp.Body.Close()

	var card map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		log.WithError(err).WithField("url", url).Warn("card decode error")
		return
	}

	err = ix.store.UpsertAgentFromCard(ctx, chain, registryAddr, agentID, d, card)
	if err != nil {
		log.WithError(err).WithField("agentID", agentID).Error("failed upserting agent from card")
		return
	}
	log.WithFields(log.Fields{
		"chain":   chain,
		"agentID": agentID,
		"domain":  d,
	}).Info("card stored")
}

// upgradeZeroIDs resolves agentId on-chain for domains saved with placeholder agent_id=0
func (ix *Indexer) upgradeZeroIDs(ctx context.Context) {
	for _, n := range ix.nets {
		log.WithField("chain", n.Name).Info("checking zero-id agents")

		client := ix.clients[n.Name]
		if client == nil {
			c, err := ethclient.Dial(os.ExpandEnv(n.RPC))
			if err != nil {
				continue
			}
			ix.clients[n.Name] = c
			client = c
		}
		idAddr, ok := ix.idents[n.Name]
		if !ok || (idAddr == common.Address{}) {
			ix.idents[n.Name] = common.HexToAddress(n.Identity)
			idAddr = ix.idents[n.Name]
		}

		domains, err := ix.store.ListZeroIDAgents(ctx, n.Name, 200)
		if err != nil || len(domains) == 0 {
			continue
		}

		ident, err := erc.NewIdentity(idAddr, client)
		if err != nil {
			continue
		}

		for _, d := range domains {
			ai, err := ident.ResolveByDomain(ctx, &bind.CallOpts{Context: ctx}, d)
			if err != nil || ai.AgentId == nil || ai.AgentId.Int64() == 0 {
				continue
			}
			ix.fetchAndStoreCard(ctx, n.Name, idAddr.Hex(), ai.AgentId.Int64(), d)
			_ = ix.store.DeleteAgent(ctx, n.Name, 0)
		}
	}
}
