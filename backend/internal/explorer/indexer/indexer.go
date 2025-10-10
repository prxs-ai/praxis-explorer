package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	erc "github.com/praxis/praxis-explorer/internal/erc8004"
	"github.com/praxis/praxis-explorer/internal/explorer/store"
	log "github.com/sirupsen/logrus"
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
	ix.startOnchainWatchers(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info("[indexer] shutting down")
			return
		case <-t.C:
			ix.crawlSeeds(ctx)
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
	count, err := ident.TotalAgents(ctx, &bind.CallOpts{Context: ctx})
	if err != nil || count == nil {
		log.WithError(err).Error("failed to get total agents")
		return
	}
	total := count.Int64()
	if total <= 0 {
		log.WithError(err).Error("total count is less than or equal to zero")
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

		// Use tokenURI directly as the domain for fetching
		domain := ai.TokenURI
		if domain == "" {
			log.WithField("agentId", ai.AgentId).Warn("tokenURI is empty for agent")
			continue
		}

		log.WithFields(log.Fields{
			"agent_id": ai.AgentId,
			"chain":    chain,
			"tokenURI": ai.TokenURI,
		}).Info("storing card")
		ix.fetchAndStoreCard(ctx, chain, idAddr.Hex(), ai.AgentId.Int64(), ai.TokenURI)
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

// Add IPFS gateway configuration
func getIPFSGateways() []string {
	gateways := os.Getenv("IPFS_GATEWAYS")
	if gateways != "" {
		return strings.Split(gateways, ",")
	}
	// Default IPFS gateways
	return []string{
		"https://ipfs.io/ipfs/",
		"https://gateway.pinata.cloud/ipfs/",
		"https://cloudflare-ipfs.com/ipfs/",
		"https://dweb.link/ipfs/",
	}
}

func (ix *Indexer) convertIPFSToHTTP(ipfsURL string) []string {
	if !strings.HasPrefix(ipfsURL, "ipfs://") {
		return []string{ipfsURL}
	}

	hash := strings.TrimPrefix(ipfsURL, "ipfs://")
	gateways := getIPFSGateways()
	var urls []string

	for _, gateway := range gateways {
		urls = append(urls, gateway+hash)
	}

	return urls
}

func (ix *Indexer) fetchIPFSMetadata(ctx context.Context, ipfsURL string) (map[string]any, error) {
	urls := ix.convertIPFSToHTTP(ipfsURL)

	for _, url := range urls {
		log.WithField("url", url).Debug("trying IPFS gateway")

		// Create a new context with timeout for each gateway attempt
		reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)

		if err != nil {
			cancel()
			continue
		}

		client := &http.Client{
			Timeout: 15 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			cancel()
			log.WithError(err).WithField("url", url).Warn("IPFS gateway failed")
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			cancel()
			log.WithField("status", resp.StatusCode).WithField("url", url).Warn("IPFS gateway returned error")
			continue
		}

		var metadata map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
			resp.Body.Close()
			cancel()
			log.WithError(err).WithField("url", url).Warn("failed to decode IPFS metadata")
			continue
		}

		resp.Body.Close()
		cancel()
		log.WithField("url", url).Info("successfully fetched IPFS metadata")
		return metadata, nil
	}

	return nil, fmt.Errorf("failed to fetch IPFS metadata from any gateway")
}

func (ix *Indexer) fetchAndStoreCard(ctx context.Context, chain string, registryAddr string, agentID int64, tokenURI string) {
	if strings.TrimSpace(tokenURI) == "" {
		log.Error("TokenURI is empty")
		return
	}

	var card map[string]any
	var err error

	// Check if this looks like an IPFS URL
	if strings.HasPrefix(tokenURI, "ipfs://") {
		card, err = ix.fetchIPFSMetadata(ctx, tokenURI)
		if err != nil {
			log.WithError(err).WithField("ipfsURL", tokenURI).Warn("failed to fetch IPFS metadata")
			return
		}
	} else {
		// Regular HTTP/HTTPS or domain-based fetching
		url := tokenURI
		if !strings.HasPrefix(tokenURI, "http://") && !strings.HasPrefix(tokenURI, "https://") {
			url = fmt.Sprintf("http://%s/.well-known/agent-card.json", tokenURI)
		} else if !strings.Contains(tokenURI, "/.well-known/agent-card.json") {
			url = strings.TrimRight(tokenURI, "/") + "/.well-known/agent-card.json"
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

		if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
			log.WithError(err).WithField("url", url).Warn("card decode error")
			return
		}
	}

	// Create a synthetic domain for database storage
	var domain string
	if strings.HasPrefix(tokenURI, "ipfs://") {
		// For IPFS, create a synthetic domain based on agent name or hash
		if agentName, ok := card["name"].(string); ok && agentName != "" {
			domain = strings.ToLower(strings.ReplaceAll(agentName, " ", "-")) + ".ipfs"
		} else {
			hash := strings.TrimPrefix(tokenURI, "ipfs://")
			if len(hash) > 12 {
				domain = hash[:12] + ".ipfs"
			} else {
				domain = hash + ".ipfs"
			}
		}
	} else {
		// For regular URLs, try to extract domain or use tokenURI
		if strings.HasPrefix(tokenURI, "http://") || strings.HasPrefix(tokenURI, "https://") {
			parts := strings.Split(tokenURI, "/")
			if len(parts) >= 3 {
				domain = parts[2]
			} else {
				domain = tokenURI
			}
		} else {
			domain = tokenURI
		}
	}

	err = ix.store.UpsertAgentFromCard(ctx, chain, registryAddr, agentID, domain, card)
	if err != nil {
		log.WithError(err).WithField("agentID", agentID).Error("failed upserting agent from card")
		return
	}
	log.WithFields(log.Fields{
		"chain":    chain,
		"agentID":  agentID,
		"domain":   domain,
		"tokenURI": tokenURI,
	}).Info("card stored")
}

func (ix *Indexer) handleIdentityLog(ctx context.Context, chain string, lg types.Log) {
	// Handle the new Registered event
	if len(lg.Topics) == 0 {
		log.Error("Topics are zero")
		return
	}

	log.WithFields(log.Fields{
		"chain": chain,
		"block": lg.BlockNumber,
	}).Debug("log received")

	evRegistered := ix.idABI.Events["Registered"]
	evMetadataSet := ix.idABI.Events["MetadataSet"]

	switch lg.Topics[0] {
	case evRegistered.ID:
		if len(lg.Topics) < 3 {
			log.Error("Number of topics is less than three for Registered event")
			return
		}
		agentId := new(big.Int).SetBytes(lg.Topics[1].Bytes())
		owner := common.BytesToAddress(lg.Topics[2].Bytes())

		log.WithFields(log.Fields{
			"chain":   chain,
			"agentId": agentId.String(),
			"owner":   owner.Hex(),
		}).Info("Registered event")

		var data struct {
			TokenURI string
		}
		if err := ix.idABI.UnpackIntoInterface(&data, "Registered", lg.Data); err != nil {
			log.WithError(err).Error("failed unpacking agent data from registered event")
			return
		}

		reg := ix.idents[chain].Hex()
		log.WithFields(log.Fields{
			"agent_id":   agentId.String(),
			"chain":      chain,
			"event_type": "registered",
			"tokenURI":   data.TokenURI,
		}).Info("storing card")
		ix.fetchAndStoreCard(ctx, chain, reg, agentId.Int64(), data.TokenURI)

	case evMetadataSet.ID:
		if len(lg.Topics) < 2 {
			log.Error("Number of topics is less than two for MetadataSet event")
			return
		}
		agentId := new(big.Int).SetBytes(lg.Topics[1].Bytes())

		var data struct {
			Key   string
			Value []byte
		}
		if err := ix.idABI.UnpackIntoInterface(&data, "MetadataSet", lg.Data); err != nil {
			log.WithError(err).Error("failed unpacking metadata from MetadataSet event")
			return
		}

		// For the new contract, we don't need to handle MetadataSet events
		// since all agent data comes from tokenURI
		log.WithFields(log.Fields{
			"agent_id":   agentId.String(),
			"chain":      chain,
			"event_type": "metadata_updated",
			"key":        data.Key,
		}).Debug("metadata set event (ignored)")
	}
}
