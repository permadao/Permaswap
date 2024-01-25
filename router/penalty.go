package router

import (
	"sync"
	"time"

	"github.com/permadao/permaswap/router/schema"
)

type Penalty struct {
	blackList      map[string]int64
	failureRecords map[string][]schema.FailureRecord // accid -> []FailRecord
	lock           sync.RWMutex
}

func NewPenalty() *Penalty {
	return &Penalty{
		blackList:      make(map[string]int64),
		failureRecords: make(map[string][]schema.FailureRecord),
	}
}

func (p *Penalty) AddFailRecord(accid string, timestamp int64, everHash string, reason string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.failureRecords[accid]; !ok {
		p.failureRecords[accid] = make([]schema.FailureRecord, 0)
	}
	p.failureRecords[accid] = append(p.failureRecords[accid], schema.FailureRecord{
		Accid:     accid,
		Timestamp: timestamp,
		EverHash:  everHash,
		Reason:    reason,
	})
	if len(p.failureRecords[accid]) >= CumulativeFailures {
		p.blackList[accid] = time.Now().Unix()

		// clear fail records
		p.failureRecords[accid] = make([]schema.FailureRecord, 0)
	}
}

func (p *Penalty) IsBlackListed(accid string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	_, ok := p.blackList[accid]
	return ok
}

func (p *Penalty) GetBlackList() map[string]int64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.blackList
}

func (p *Penalty) GetFailureRecords() map[string][]schema.FailureRecord {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.failureRecords
}

func (p *Penalty) GetPenalty() (map[string]int64, map[string][]schema.FailureRecord) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.blackList, p.failureRecords
}

func (p *Penalty) ClearUpExpired() {
	p.lock.Lock()
	defer p.lock.Unlock()

	now := time.Now().Unix()
	for accid, timestamp := range p.blackList {
		if now-timestamp > ExpirationDuration {
			delete(p.blackList, accid)
		}
	}

	for accid, records := range p.failureRecords {
		for i, record := range records {
			if now-record.Timestamp > ExpirationDuration {
				p.failureRecords[accid] = append(records[:i], records[i+1:]...)
			}
		}
	}
}
