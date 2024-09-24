package policy

import (
	"flare-common/contracts/relay"
	"flare-common/database"
	"flare-common/events"

	"time"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

const (
	listenerBufferSize               = 10
	ListenerInterval   time.Duration = 2 * time.Second
)

type SigningPolicyListenerResponse struct {
	policyData *relay.RelaySigningPolicyInitialized
	timestamp  int64
}

type relayContractClient struct {
	address common.Address

	relay *relay.Relay

	topic0SPI common.Hash // for SigningPolicyInitialized event
	topic0PMR common.Hash // for ProtocolMessageRelayed event
}

func (r *relayContractClient) FetchSigningPolicies(db *gorm.DB, from, to int64) ([]SigningPolicyListenerResponse, error) {
	logs, err := database.FetchLogsByAddressAndTopic0Timestamp(db, r.address, r.topic0SPI, from, to)
	if err != nil {
		return nil, err
	}

	result := make([]SigningPolicyListenerResponse, 0, len(logs))
	for _, log := range logs {
		policyData, err := events.ParseSigningPolicyInitializedEvent(r.relay, log)
		if err != nil {
			return nil, err
		}
		result = append(result, SigningPolicyListenerResponse{policyData, int64(log.Timestamp)})
	}
	return result, nil
}

func (r *relayContractClient) SigningPolicyInitializedListener(db *gorm.DB, startTime time.Time) <-chan SigningPolicyListenerResponse {
	out := make(chan SigningPolicyListenerResponse, listenerBufferSize)
	go func() {
		ticker := time.NewTicker(ListenerInterval)
		eventRangeStart := startTime.Unix()
		for {
			<-ticker.C
			now := time.Now().Unix()
			logs, err := database.FetchLogsByAddressAndTopic0Timestamp(db, r.address, r.topic0SPI, eventRangeStart, now)
			if err != nil {
				continue
			}
			for _, log := range logs {
				policyData, err := events.ParseSigningPolicyInitializedEvent(r.relay, log)
				if err != nil {
					break
				}
				out <- SigningPolicyListenerResponse{policyData, int64(log.Timestamp)}
				// continue with timestamps > log.Timestamp,
				// there should be only one such log per timestamp
				eventRangeStart = int64(log.Timestamp)
			}
		}
	}()
	return out
}
