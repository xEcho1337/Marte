package submitter

import (
	"backend/database"
	"sync"
	"time"
)

type FlagManager struct {
	mu       sync.Mutex
	seen     map[string]SeenEntry
	order    []string
	pending  []Flag
	started  bool
	database *database.Manager
	stopChan chan struct{}
}

func NewFlagManager(initialCapacity int, database *database.Manager) *FlagManager {
	if initialCapacity < 0 {
		initialCapacity = 0
	}

	return &FlagManager{
		seen:     make(map[string]SeenEntry, initialCapacity),
		order:    make([]string, 0, initialCapacity),
		pending:  make([]Flag, 0, initialCapacity),
		stopChan: make(chan struct{}),
		database: database,
	}
}

func (m *FlagManager) LoadCache(records []database.FlagRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, r := range records {
		key := r.Value
		if _, ok := m.seen[key]; ok {
			continue
		}
		m.seen[key] = SeenEntry{
			Status:    FlagResultStatus(r.Status),
			TeamId:    r.TeamId,
			Submitter: r.Submitter,
			Timestamp: r.Timestamp,
		}
		m.order = append(m.order, key)
	}
}

func (m *FlagManager) SubmitFlag(f Flag) bool {
	key := f.Value

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.seen[key]; ok {
		return false
	}

	m.seen[key] = SeenEntry{
		Status:    StatusPending,
		TeamId:    f.TeamId,
		Submitter: f.Submitter,
		Timestamp: time.Now().UnixMilli(),
	}
	m.order = append(m.order, key)
	m.pending = append(m.pending, f)
	return true
}

func (m *FlagManager) Start(submitterType, urlFlagSubmit, teamToken string, flagBatch, flagSubmitRate int) {
	m.mu.Lock()

	if m.started {
		m.mu.Unlock()
		return
	}

	m.started = true
	m.mu.Unlock()

	submitter := createSubmitter(submitterType, urlFlagSubmit, teamToken)
	if submitter == nil {
		return
	}

	batchSize := max(flagBatch, 1)

	interval := time.Duration(flagSubmitRate) * time.Second
	if interval <= 0 {
		interval = time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				batch := m.drain(batchSize)

				if len(batch) == 0 {
					continue
				}

				subLog.Info("Submitting batch of %d flags", len(batch))

				submitted, err := submitter.Submit(batch)

				if err != nil {
					subLog.Error("Error: " + err.Error())

					if err.RateLimited {
						subLog.Error("Ratelimited")
						for i := range batch {
							m.requeue(batch[i])
						}
					}
					continue
				}

				records := make([]database.FlagRecord, 0, len(submitted))

				for _, result := range submitted {
					if entry, ok := m.seen[result.Value]; ok {
						entry.Status = result.Status
						m.seen[result.Value] = entry
						records = append(records, database.FlagRecord{
							Value:     result.Value,
							Status:    int(entry.Status),
							TeamId:    entry.TeamId,
							Submitter: entry.Submitter,
							Timestamp: entry.Timestamp,
						})
					}
				}

				m.database.InsertAllFlags(records)
			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *FlagManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.stopChan:
		return
	default:
		close(m.stopChan)
	}
}

func (m *FlagManager) requeue(f Flag) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pending = append(m.pending, f)
}

func (m *FlagManager) drain(n int) []Flag {
	m.mu.Lock()
	defer m.mu.Unlock()

	if n <= 0 || len(m.pending) == 0 {
		return nil
	}

	if n > len(m.pending) {
		n = len(m.pending)
	}

	batch := make([]Flag, n)
	copy(batch, m.pending[:n])

	for i := 0; i < n; i++ {
		m.pending[i] = Flag{}
	}

	m.pending = m.pending[n:]

	if len(m.pending) == 0 {
		m.pending = m.pending[:0]
	} else if cap(m.pending) > 0 && len(m.pending)*4 < cap(m.pending) {
		tmp := make([]Flag, len(m.pending))
		copy(tmp, m.pending)
		m.pending = tmp
	}

	return batch
}

func (m *FlagManager) GetSeenSlice(limit, offset int) []FlagEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	total := len(m.order)
	if offset >= total {
		return nil
	}

	end := offset + limit
	if limit <= 0 || end > total {
		end = total
	}

	out := make([]FlagEntry, 0, end-offset)
	for i := total - 1 - offset; i >= total-end && i >= 0; i-- {
		key := m.order[i]
		if entry, ok := m.seen[key]; ok {
			out = append(out, FlagEntry{
				Value:     key,
				Status:    entry.Status,
				TeamId:    entry.TeamId,
				Submitter: entry.Submitter,
				Timestamp: entry.Timestamp,
			})
		}
	}

	return out
}

func (m *FlagManager) GetTotalCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.order)
}

func (m *FlagManager) GetSummary() FlagSummary {
	m.mu.Lock()
	defer m.mu.Unlock()

	var s FlagSummary
	s.Total = len(m.order)
	for _, entry := range m.seen {
		switch entry.Status {
		case StatusAccepted:
			s.Accepted++
		case StatusPending:
			s.Pending++
		case StatusDuplicate, StatusTooOld:
			s.Rejected++
		}
	}
	return s
}

func (m *FlagManager) GetTimeline() []TimelinePoint {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.order) == 0 {
		return nil
	}

	var minTs, maxTs int64
	first := true
	for _, entry := range m.seen {
		if first {
			minTs = entry.Timestamp
			maxTs = entry.Timestamp
			first = false
		} else {
			if entry.Timestamp < minTs {
				minTs = entry.Timestamp
			}
			if entry.Timestamp > maxTs {
				maxTs = entry.Timestamp
			}
		}
	}

	durationMs := maxTs - minTs
	if durationMs <= 0 {
		durationMs = 60000
	}

	bucketMs := int64(60000)
	switch {
	case durationMs < 30*60000:
		bucketMs = 60000
	case durationMs < 2*3600000:
		bucketMs = 5 * 60000
	case durationMs < 12*3600000:
		bucketMs = 15 * 60000
	default:
		bucketMs = 60 * 60000
	}

	numBuckets := int(durationMs/bucketMs) + 1
	if numBuckets > 100 {
		numBuckets = 100
		bucketMs = durationMs / int64(numBuckets-1)
		if bucketMs <= 0 {
			bucketMs = 60000
		}
	}

	buckets := make(map[int64]*TimelinePoint, numBuckets)
	startBucket := (minTs / bucketMs) * bucketMs
	for i := 0; i < numBuckets; i++ {
		ts := startBucket + int64(i)*bucketMs
		buckets[ts] = &TimelinePoint{Timestamp: ts}
	}

	for _, entry := range m.seen {
		bucketKey := (entry.Timestamp / bucketMs) * bucketMs
		bp, ok := buckets[bucketKey]
		if !ok {
			bp = &TimelinePoint{Timestamp: bucketKey}
			buckets[bucketKey] = bp
		}
		switch entry.Status {
		case StatusAccepted:
			bp.Accepted++
		case StatusDuplicate, StatusTooOld:
			bp.Rejected++
		}
	}

	out := make([]TimelinePoint, 0, len(buckets))
	for _, bp := range buckets {
		out = append(out, *bp)
	}

	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Timestamp < out[i].Timestamp {
				out[i], out[j] = out[j], out[i]
			}
		}
	}

	return out
}

func (m *FlagManager) GetTopAttackers(N int) []TopAttackerEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	type stats struct {
		accepted int
		total    int
	}
	userStats := make(map[string]*stats)

	for _, entry := range m.seen {
		s, ok := userStats[entry.Submitter]
		if !ok {
			s = &stats{}
			userStats[entry.Submitter] = s
		}
		s.total++
		if entry.Status == StatusAccepted {
			s.accepted++
		}
	}

	entries := make([]TopAttackerEntry, 0, len(userStats))
	for username, s := range userStats {
		entries = append(entries, TopAttackerEntry{Username: username, Accepted: s.accepted, Total: s.total})
	}

	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Accepted > entries[i].Accepted {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	if N > 0 && N < len(entries) {
		entries = entries[:N]
	}

	return entries
}
