package proxy

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
)

type ExecutionMonitor struct {
	mu        sync.Mutex
	upstreams map[string]ExecutionUpstream
	status    map[string]ExecutionStatus
	log       *logrus.Logger
}

type ExecutionStatus struct {
	HeadBlock uint64 `json:"head_block"`
	ChainID   uint64 `json:"chain_id"`
	PeerCount uint64 `json:"peer_count"`
	IsSyncing bool   `json:"is_syncing"`
	LastCheck int64  `json:"last_check"`
}

func NewExecutionMonitor(log *logrus.Logger, upstreams []ExecutionUpstream) *ExecutionMonitor {
	targets := make(map[string]ExecutionUpstream)
	status := make(map[string]ExecutionStatus)

	for _, u := range upstreams {
		targets[u.Name] = u
		status[u.Name] = ExecutionStatus{}
	}

	em := ExecutionMonitor{
		upstreams: targets,
		status:    status,
		log:       log,
	}

	em.CheckAll()

	ticker := time.NewTicker(15 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				em.CheckAll()
			}
		}
	}()

	return &em
}

func (em *ExecutionMonitor) CheckAll() {
	em.log.WithField("count", len(em.upstreams)).Debug("checking all execution nodes")

	var wg sync.WaitGroup

	wg.Add(len(em.upstreams))

	for _, u := range em.upstreams {
		go func(upstreamName string) {
			err := em.Check(upstreamName)

			if err != nil {
				em.log.WithField("upstream", upstreamName).WithError(err).Error("failed checking execution node")
			}

			wg.Done()
		}(u.Name)
	}

	wg.Wait()
}

func (em *ExecutionMonitor) Check(upstreamName string) error {
	em.log.WithField("node", upstreamName).Debug("checking execution node")

	syncing, err := em.CheckNodeSyncing(upstreamName)
	if err != nil {
		em.log.WithField("upstream", upstreamName).WithError(err).Error("failed getting execution node sync info")
	}

	headBlock, err := em.CheckNodeHeadBlock(upstreamName)
	if err != nil {
		em.log.WithField("upstream", upstreamName).WithError(err).Error("failed getting execution node head block")
	}

	chainID, err := em.CheckNodechainID(upstreamName)
	if err != nil {
		em.log.WithField("upstream", upstreamName).WithError(err).Error("failed getting execution node chain id")
	}

	peerCount, err := em.CheckNodePeerCount(upstreamName)
	if err != nil {
		em.log.WithField("upstream", upstreamName).WithError(err).Error("failed getting execution node peer count")
	}

	now := time.Now().Unix()
	bs := ExecutionStatus{
		IsSyncing: syncing,
		LastCheck: now,
		HeadBlock: headBlock,
		ChainID:   chainID,
		PeerCount: peerCount,
	}

	em.mu.Lock()
	em.status[upstreamName] = bs
	em.mu.Unlock()

	return nil
}

func (em *ExecutionMonitor) CheckNodeSyncing(upstreamName string) (bool, error) {
	upstream := em.upstreams[upstreamName]

	client, err := rpc.Dial(upstream.Address)
	if err != nil {
		return true, err
	}

	var result bool
	err = client.Call(&result, "eth_syncing", nil)

	return result, err
}

func (em *ExecutionMonitor) CheckNodeHeadBlock(upstreamName string) (uint64, error) {
	upstream := em.upstreams[upstreamName]

	client, err := rpc.Dial(upstream.Address)
	if err != nil {
		return 0, err
	}

	var result string

	err = client.Call(&result, "eth_blockNumber", nil)
	if err != nil {
		return 0, err
	}

	res, err := hexutil.DecodeUint64(result)
	if err != nil {
		return 0, err
	}

	return res, err
}

func (em *ExecutionMonitor) CheckNodechainID(upstreamName string) (uint64, error) {
	upstream := em.upstreams[upstreamName]

	client, err := rpc.Dial(upstream.Address)
	if err != nil {
		return 0, err
	}

	var result string

	err = client.Call(&result, "eth_chainID", nil)
	if err != nil {
		return 0, err
	}

	res, err := hexutil.DecodeUint64(result)
	if err != nil {
		return 0, err
	}

	return res, err
}

func (em *ExecutionMonitor) CheckNodePeerCount(upstreamName string) (uint64, error) {
	upstream := em.upstreams[upstreamName]

	client, err := rpc.Dial(upstream.Address)
	if err != nil {
		return 0, err
	}

	var result string

	err = client.Call(&result, "net_peerCount", nil)
	if err != nil {
		return 0, err
	}

	res, err := hexutil.DecodeUint64(result)
	if err != nil {
		return 0, err
	}

	return res, err
}
