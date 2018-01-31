package nodes

import (
	"net/http"

	"NOMP-API/db"

	"github.com/Vilsol/GoLib"
)

func RegisterEverythingRoutes(router GoLib.RegisterRoute) {
	router("GET", "/stats", getStats)
	router("GET", "/pool_stats", getPoolStats)
	router("GET", "/shares", getShares)
	router("GET", "/pending", getPending)
	router("GET", "/blocks", getBlocks)
	router("GET", "/workers", getWorkers)
}

func getStats(_ *http.Request) (interface{}, *GoLib.ErrorResponse) {
	return db.GetStats(), nil
}

func getPoolStats(_ *http.Request) (interface{}, *GoLib.ErrorResponse) {
	return db.GetPoolStatsWithoutWorkers(), nil
}

func getShares(_ *http.Request) (interface{}, *GoLib.ErrorResponse) {
	return db.GetShares(), nil
}

func getPending(_ *http.Request) (interface{}, *GoLib.ErrorResponse) {
	return db.GetPending(), nil
}

func getBlocks(_ *http.Request) (interface{}, *GoLib.ErrorResponse) {
	return db.GetBlockData(), nil
}

func getWorkers(_ *http.Request) (interface{}, *GoLib.ErrorResponse) {
	return db.GetWorkers(), nil
}
