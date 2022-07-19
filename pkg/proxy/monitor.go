package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type BeaconMonitor struct {
	mu        sync.Mutex
	upstreams map[string]BeaconUpstream
	status    map[string]BeaconStatus
	log       *logrus.Logger
}

type BeaconStatus struct {
	Version   string        `json:"version"`
	Health    string        `json:"health"`
	Syncing   SyncInfo      `json:"syncing"`
	PeerCount PeerCountInfo `json:"peerCount"`
	LastCheck int64         `json:"lastCheck"`
}

type SyncInfo struct {
	HeadSlot     string `json:"head_slot"`
	SyncDistance string `json:"sync_distance"`
	IsSyncing    bool   `json:"is_syncing"`
	IsOptimistic bool   `json:"is_optimistic"`
}

type PeerCountInfo struct {
	Disconnected int `json:"disconnected"`
	Connected    int `json:"connected"`
}

// TODO:
// - Make timers configurable (ticker, http timeout)
// - Make node check requests async and wait for all results

func NewBeaconMonitor(log *logrus.Logger, upstreams []BeaconUpstream) *BeaconMonitor {
	targets := make(map[string]BeaconUpstream)
	status := make(map[string]BeaconStatus)

	for _, u := range upstreams {
		targets[u.Name] = u
		status[u.Name] = BeaconStatus{}
	}

	bm := BeaconMonitor{
		upstreams: targets,
		status:    status,
		log:       log,
	}

	bm.CheckAll()

	ticker := time.NewTicker(15 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				bm.CheckAll()
			}
		}
	}()

	return &bm
}

func (bm *BeaconMonitor) CheckAll() {
	bm.log.WithField("count", len(bm.upstreams)).Debug("checking all nodes")

	var wg sync.WaitGroup

	wg.Add(len(bm.upstreams))

	for _, u := range bm.upstreams {
		go func(upstreamName string) {
			err := bm.Check(upstreamName)

			if err != nil {
				bm.log.WithField("upstream", upstreamName).WithError(err).Error("failed checking node")
			}

			wg.Done()
		}(u.Name)
	}

	wg.Wait()
}

func (bm *BeaconMonitor) Check(upstreamName string) error {
	bm.log.WithField("node", upstreamName).Debug("checking node")
	version, err := bm.CheckNodeVersion(upstreamName)

	if err != nil {
		bm.log.WithField("upstream", upstreamName).WithError(err).Error("failed getting node version")
	}

	syncInfo, err := bm.CheckNodeSyncing(upstreamName)
	if err != nil {
		bm.log.WithField("upstream", upstreamName).WithError(err).Error("failed getting node sync info")
	}

	peerCountInfo, err := bm.CheckNodePeerCount(upstreamName)
	if err != nil {
		bm.log.WithField("upstream", upstreamName).WithError(err).Error("failed getting node peer count info")
	}

	now := time.Now().Unix()
	bs := BeaconStatus{
		Version:   version,
		LastCheck: now,
	}

	if syncInfo != nil {
		bs.Syncing = *syncInfo
	}

	if peerCountInfo != nil {
		bs.PeerCount = *peerCountInfo
	}

	bm.mu.Lock()
	bm.status[upstreamName] = bs
	defer bm.mu.Unlock()

	return nil
}

func (bm *BeaconMonitor) CheckNodeVersion(upstreamName string) (string, error) {
	upstream := bm.upstreams[upstreamName]
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/version", upstream.Address))

	if err != nil {
		return "", err
	}

	r := struct {
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return "", errors.Wrap(err, "failed decoding response body")
	}

	return r.Data.Version, nil
}

func (bm *BeaconMonitor) CheckNodeSyncing(upstreamName string) (*SyncInfo, error) {
	upstream := bm.upstreams[upstreamName]
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/syncing", upstream.Address))
	if err != nil {
		return nil, err
	}

	r := struct {
		Data struct {
			SyncInfo
		} `json:"data"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "failed decoding response body")
	}

	return &r.Data.SyncInfo, nil
}

func (bm *BeaconMonitor) CheckNodePeerCount(upstreamName string) (*PeerCountInfo, error) {
	upstream := bm.upstreams[upstreamName]
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("%s/eth/v1/node/peer_count", upstream.Address))

	if err != nil {
		return nil, err
	}

	// Some clients return the fields as int, others as strings, so we have to deal with both cases
	r := struct {
		Data struct {
			Connected    interface{} `json:"connected"`
			Disconnected interface{} `json:"disconnected"`
		} `json:"data"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "failed decoding response body")
	}

	ret := PeerCountInfo{}

	switch v := r.Data.Connected.(type) {
	case float64:
		ret.Connected = int(v)
	case string:
		m, _ := strconv.Atoi(v)
		ret.Connected = m
	}

	switch v := r.Data.Disconnected.(type) {
	case float64:
		ret.Disconnected = int(v)
	case string:
		m, _ := strconv.Atoi(v)
		ret.Disconnected = m
	}

	return &ret, nil
}
