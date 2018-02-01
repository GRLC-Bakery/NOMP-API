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
	Address       string             `json:"address"`
	Hashrate      float64            `json:"hashrate"`
	UnpaidBalance float64            `json:"unpaid_balance"`
	Workers       map[string]float64 `json:"workers"`
}

func getWorker(r *http.Request) (interface{}, *GoLib.ErrorResponse) {
	address := mux.Vars(r)["address"]

	allWorkers := db.GetWorkers()

	workerStats, ok := allWorkers[address]

	hashrate := 0.0
	unpaid := 0.0

	if ok {
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

	workers := make(map[string]float64)

	// TODO Temporary
	workers["0"] = hashrate

	return WorkerResponse{
		Address:       address,
		Hashrate:      hashrate,
		UnpaidBalance: unpaid,
		Workers:       workers,
	}, nil
}
