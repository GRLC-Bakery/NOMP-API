package db

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"NOMP-API/utils"

	"github.com/go-redis/redis"
)

var client *redis.Client

func InitializeRedis(redisHost string, redisPort int, redisPassword string, redisDb int) {
	client = redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + strconv.Itoa(redisPort),
		Password: redisPassword,
		DB:       redisDb,
	})

	_, err := client.Ping().Result()

	if err != nil {
		panic(err)
	}

	go reloadLoop()
}

func reloadLoop() {
	for {
		fmt.Println("Reloading Stats")
		ReloadStats()
		ReloadPoolStats()
		ReloadShares()

		time.Sleep(time.Second * 15)
	}
}

func GetPending() []*Block {
	pending := client.SMembers(utils.Coin + ":blocksPending")
	result := pending.Val()
	blocks := make([]*Block, len(result))
	for i, block := range result {
		split := strings.Split(block, ":")
		height, _ := strconv.Atoi(split[2])
		blocks[i] = &Block{
			BlockHash:   split[0],
			TxHash:      split[1],
			BlockHeight: height,
		}
	}
	return blocks
}

func GetBalances() map[string]float64 {
	balances := client.HGetAll(utils.Coin + ":balances")
	wallets := make(map[string]float64)
	for wallet, balance := range balances.Val() {
		wallets[wallet], _ = strconv.ParseFloat(balance, 64)
	}
	return wallets
}

var sharesCache map[int]map[string]float64

func GetShares() map[int]map[string]float64 {
	return sharesCache
}

func ReloadShares() {
	pending := GetPending()

	pipeline := client.TxPipeline()
	rounds := make([]*redis.StringStringMapCmd, len(pending))
	for i, block := range pending {
		rounds[i] = pipeline.HGetAll(utils.Coin + ":shares:round" + strconv.Itoa(block.BlockHeight))
	}
	pipeline.Exec()

	shares := make(map[int]map[string]float64)
	for i := 0; i < len(pending); i++ {
		shares[pending[i].BlockHeight] = make(map[string]float64)
		round := rounds[i].Val()
		for wallet, count := range round {
			sharesFloat, _ := strconv.ParseFloat(count, 64)
			shares[pending[i].BlockHeight][wallet] = sharesFloat
		}
	}

	sharesCache = shares
}

func GetBlockData() *BlockData {
	pipeline := client.TxPipeline()
	lastMined := pipeline.Get(utils.Coin + ":lastMined")
	lastPayout := pipeline.Get(utils.Coin + ":lastPayout")
	blockHeight := pipeline.Get(utils.Coin + ":blockHeight")
	pipeline.Exec()

	lastMinedSplit := strings.Split(lastMined.Val(), ":")
	lastPayoutSplit := strings.Split(lastPayout.Val(), ":")

	lastMinedInt, _ := strconv.Atoi(lastMinedSplit[1])
	lastPayoutInt, _ := strconv.Atoi(lastPayoutSplit[1])
	blockHeightInt, _ := strconv.Atoi(blockHeight.Val())

	return &BlockData{
		LastMined: Block{
			BlockHash:   lastMinedSplit[0],
			BlockHeight: lastMinedInt,
		},
		LastPayout: Block{
			BlockHash:   lastPayoutSplit[0],
			BlockHeight: lastPayoutInt,
		},
		Height: blockHeightInt,
	}
}

var statsCache Stats

func GetStats() Stats {
	return statsCache
}

func ReloadStats() {
	pipeline := client.TxPipeline()

	hashrate := pipeline.ZRangeByScore(utils.Coin+":hashrate", redis.ZRangeBy{
		Min: strconv.Itoa(int(time.Now().Unix()) - utils.HashrateWindow),
		Max: "+inf",
	})

	stats := pipeline.HGetAll(utils.Coin + ":stats")
	blocksPending := pipeline.SCard(utils.Coin + ":blocksPending")
	blocksConfirmed := pipeline.SCard(utils.Coin + ":blocksConfirmed")
	blocksOrphaned := pipeline.SCard(utils.Coin + ":blocksOrphaned")

	pipeline.Exec()

	validShares, _ := strconv.Atoi(stats.Val()["validShares"])
	invalidShares, _ := strconv.Atoi(stats.Val()["invalidShares"])
	validBlocks, _ := strconv.Atoi(stats.Val()["validBlocks"])
	totalPaid, _ := strconv.ParseFloat(stats.Val()["totalPaid"], 64)

	statsCache = statsToModel(validShares, invalidShares, validBlocks, totalPaid, hashrate.Val(), blocksPending.Val(), blocksConfirmed.Val(), blocksOrphaned.Val())
}

var poolStatsHistoryCache []PoolStatsHistory
var poolStatsHistoryWithoutWorkersCache []PoolStatsHistory

func GetPoolStats() []PoolStatsHistory {
	return poolStatsHistoryCache
}

func GetPoolStatsWithoutWorkers() []PoolStatsHistory {
	return poolStatsHistoryWithoutWorkersCache
}

func ReloadPoolStats() {
	statHistoryResult := client.ZRangeByScore("statHistory", redis.ZRangeBy{
		Min: strconv.Itoa(int(time.Now().Unix()) - utils.HistoricalRetention),
		Max: "+inf",
	})

	if statHistoryResult.Err() != nil {
		fmt.Println(statHistoryResult.Err())
	}

	statHistory := statHistoryResult.Val()

	type temp struct {
		Time  int64                      `json:"time"`
		Pools map[string]json.RawMessage `json:"pools"`
	}

	history := make([]PoolStatsHistory, len(statHistory))
	historyWithoutWorkers := make([]PoolStatsHistory, len(statHistory))

	for i, v := range statHistory {
		var data temp
		json.Unmarshal([]byte(v), &data)

		var coinData Stats
		var coinDataWithoutWorkers Stats

		json.Unmarshal(data.Pools[utils.Coin], &coinData)
		json.Unmarshal(data.Pools[utils.Coin], &coinDataWithoutWorkers)

		coinDataWithoutWorkers.Workers = nil

		history[i] = PoolStatsHistory{
			Time:  data.Time,
			Stats: coinData,
		}

		historyWithoutWorkers[i] = PoolStatsHistory{
			Time:  data.Time,
			Stats: coinDataWithoutWorkers,
		}
	}

	poolStatsHistoryCache = history
	poolStatsHistoryWithoutWorkersCache = historyWithoutWorkers
}

func statsToModel(validShares int, invalidShares int, validBlocks int, totalPaid float64, hashrates []string, blocksPending int64, blocksConfirmed int64, blocksOrphaned int64) Stats {
	workers := make(map[string]*WorkerStats)
	totalShares := 0.0
	for _, worker := range hashrates {
		split := strings.Split(worker, ":")

		if _, ok := workers[split[1]]; !ok {
			workers[split[1]] = &WorkerStats{
				ValidShares:   0,
				InvalidShares: 0,
			}
		}

		share, _ := strconv.ParseFloat(split[0], 64)
		workerStats := workers[split[1]]
		if share > 0 {
			totalShares += share
			workerStats.ValidShares += share
		} else {
			workerStats.InvalidShares += math.Abs(share)
		}
	}

	for _, worker := range workers {
		worker.Hashrate = float64(utils.HashrateMultiplier) * worker.ValidShares / float64(utils.HashrateWindow)
	}

	return Stats{
		Workers: workers,
		PoolStats: PoolStats{
			ValidShares:   validShares,
			InvalidShares: invalidShares,
			ValidBlocks:   validBlocks,
			TotalPaid:     totalPaid,
		},
		BlockStats: BlockStats{
			Pending:   blocksPending,
			Confirmed: blocksConfirmed,
			Orphaned:  blocksOrphaned,
		},
		WorkerCount: int64(len(workers)),
		Hashrate:    float64(utils.HashrateMultiplier) * totalShares / float64(utils.HashrateWindow),
	}
}

func GetWorkers() map[string]*WorkerStats {
	balances := GetBalances()
	workers := GetStats().Workers

	for address, worker := range workers {
		if _, ok := balances[address]; ok {
			worker.Balance = balances[address]
		}
	}

	return workers
}
