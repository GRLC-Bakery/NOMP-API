package utils

var Coin string
var HashrateMultiplier int
var HashrateWindow int
var HistoricalRetention int

func InitializeFlags(coin *string, hashrateMultiplier *int, hashrateWindow *int, historicalRetention *int) {
	Coin = *coin
	HashrateMultiplier = *hashrateMultiplier
	HashrateWindow = *hashrateWindow
	HistoricalRetention = *historicalRetention
}
