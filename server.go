package NOMP_API

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"NOMP-API/db"
	"NOMP-API/nodes"
	"NOMP-API/utils"

	"github.com/Vilsol/GoLib"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func Serve() {
	// Read flags
	redisHost := flag.String("redis-host", "localhost", "Redis Host")
	redisPort := flag.Int("redis-port", 6379, "Redis Port")
	redisPassword := flag.String("redis-password", "", "Redis Password (default \"\")")
	redisDb := flag.Int("redis-db", 0, "Redis DB (default 0)")

	coin := flag.String("coin", "garlicoin", "Coin")
	hashrateMultiplier := flag.Int("hashrate-multiplier", 65536, "Hashrate Multiplier")
	hashrateWindow := flag.Int("hashrate-window", 300, "Hashrate Window")
	historicalRetention := flag.Int("historical-retention", 43200, "Historical Retention Time")

	listenPort := flag.Int("listen-port", 8080, "Listening Port")

	flag.Parse()

	utils.InitializeFlags(coin, hashrateMultiplier, hashrateWindow, historicalRetention)

	db.InitializeRedis(*redisHost, *redisPort, *redisPassword, *redisDb)

	router := mux.NewRouter()

	router.NotFoundHandler = GoLib.LoggerHandler(GoLib.NotFoundHandler())

	v1 := GoLib.RouteHandler(router, "/v1")
	nodes.RegisterWorkerRoutes(v1)
	nodes.RegisterEverythingRoutes(v1)

	CORSHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "HEAD", "OPTIONS"}),
	)

	var finalRouter http.Handler = router
	finalRouter = CORSHandler(finalRouter)
	finalRouter = GoLib.LoggerHandler(finalRouter)
	finalRouter = handlers.CompressHandler(finalRouter)
	finalRouter = handlers.ProxyHeaders(finalRouter)

	fmt.Printf("Listening on port %d\n", *listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *listenPort), finalRouter))

}
