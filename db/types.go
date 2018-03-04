package db

type Block struct {
	BlockHash   string `json:"blockHash"`
	TxHash      string `json:"txHash"`
	BlockHeight int    `json:"blockHeight"`
}

type BlockData struct {
	LastMined  Block `json:"lastMined"`
	LastPayout Block `json:"lastPayout"`
	Height     int   `json:"blockHeight"`
}

type Stats struct {
	Time        int64                   `json:"time,omitempty"`
	Workers     map[string]*WorkerStats `json:"workers,omitempty"`
	PoolStats   PoolStats               `json:"poolStats"`
	BlockStats  BlockStats              `json:"blocks"`
	Hashrate    float64                 `json:"hashrate"`
	WorkerCount int64                   `json:"workerCount"`
}

type PoolStats struct {
	ValidShares   int     `json:"validShares,string"`
	InvalidShares int     `json:"invalidShares,string"`
	ValidBlocks   int     `json:"validBlocks,string"`
	TotalPaid     float64 `json:"totalPaid,string"`
}

type BlockStats struct {
	Pending   int64 `json:"pending"`
	Confirmed int64 `json:"confirmed"`
	Orphaned  int64 `json:"orphaned"`
}

type WorkerStats struct {
	ValidShares   float64                       `json:"shares"`
	InvalidShares float64                       `json:"invalidshares"`
	Hashrate      float64                       `json:"hashrate"`
	Balance       float64                       `json:"balance"`
	Workers       map[string]*CustomWorkerStats `json:"workers"`
}

type CustomWorkerStats struct {
	ValidShares   float64 `json:"shares"`
	InvalidShares float64 `json:"invalidshares"`
	Hashrate      float64 `json:"hashrate"`
}

type PoolStatsHistory struct {
	Time  int64 `json:"time"`
	Stats Stats `json:"stats"`
}
