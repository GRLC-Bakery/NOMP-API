package nodes

import (
	"net/http"

	"NOMP-API/db"
	"NOMP-API/utils"

	"github.com/Vilsol/GoLib"
	"github.com/gorilla/mux"
)

var blockReward = 50.0

func RegisterWorkerRoutes(router GoLib.RegisterRoute) {
	router("GET", "/worker/{address}", getWorker)
}

type WorkerResponse struct {
	Address       string                   `json:"address"`
	Hashrate      float64                  `json:"hashrate"`
	UnpaidBalance float64                  `json:"unpaid_balance"`
	Workers       map[string]float64       `json:"workers"`
	WorkerStats   map[string][]WorkerStats `json:"worker_stats"`
}

type WorkerStats struct {
	Time          int64   `json:"time"`
	Hashrate      float64 `json:"hashrate"`
	ValidShares   float64 `json:"valid_shares"`
	InvalidShares float64 `json:"invalid_shares"`
}

func getWorker(r *http.Request) (interface{}, *GoLib.ErrorResponse) {
	address := mux.Vars(r)["address"]

	workerStats := db.GetWorker(address)

	hashrate := 0.0
	unpaid := 0.0

	if workerStats != nil {
		hashrate = workerStats.Hashrate
		unpaid = workerStats.Balance
	}

	shares := db.GetShares()

	for _, block := range shares {
		totalSatoshis := 0.0
		workerSatoshis := 0.0
		for workerAddress, satoshis := range block {
			totalSatoshis += satoshis
			if workerAddress == address {
				workerSatoshis = satoshis
			}
		}
		if workerSatoshis > 0 {
			unpaid += blockReward * (workerSatoshis / totalSatoshis)
		}
	}

	if unpaid == 0 && hashrate == 0 {
		return nil, &utils.ErrorWorkerNotFound
	}

	stats := make(map[string][]WorkerStats)

	for _, snapshot := range db.GetPoolStats() {
		if data, ok := snapshot.Stats.Workers[address]; ok {
			for name, worker := range data.Workers {
				stats[name] = append(stats[name], WorkerStats{
					Time:          snapshot.Time,
					Hashrate:      worker.Hashrate,
					ValidShares:   worker.ValidShares,
					InvalidShares: worker.InvalidShares,
				})
			}
		}
	}

	workers := make(map[string]float64)

	for name, worker := range workerStats.Workers {
		workers[name] = worker.Hashrate
	}

	if len(workerStats.Workers) == 0 {
		workers["Baker"] = hashrate
	}

	return WorkerResponse{
		Address:       address,
		Hashrate:      hashrate,
		UnpaidBalance: unpaid,
		Workers:       workers,
		WorkerStats:   stats,
	}, nil
}
