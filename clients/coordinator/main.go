package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"net/http"
	"opl/clients/publicUsed"
	"opl/database"
	"opl/entity"
	"opl/opt"
	"opl/rlp"
	"opl/smartcontract"
	"opl/stateManager"
	"strconv"
	"time"
)

func main() {
	configPath := flag.String("config", "./coordinator_config.yaml", "Path to the configuration file for initialization")
	autoConMode := flag.Bool("autoconn", false, "Whether to autonomously dial other nodes and bind peers")
	dialDelay := flag.Int("dialDelay", 5, "Dial delay after initialize the coordinator in seconds")
	bindDelay := flag.Int("bindDelay", 10, "Bind delay after dialing all nodes in seconds")

	flag.Parse()

	dialD := time.Duration(*dialDelay) * time.Second
	bindD := time.Duration(*bindDelay) * time.Second

	cfg, err := opt.LoadCoordinatorConfig(*configPath)
	if err != nil {
		log.Fatalf("fail in loading configuration file: %v", err)
	}

	cfg.Print()

	var sldb *database.SimpleLDB
	if cfg.MemDBMode {
		sldb = database.NewSimpleMemLDB()
	} else {
		db, err := leveldb.OpenFile(cfg.DatabasePath, &publicUsed.DB_USED)
		if err != nil {
			panic(err)
		}
		sldb = database.NewSimpleLDB("DB_"+cfg.CoordinatorName, db)
	}
	defer sldb.Close()

	var cdnt *entity.Coordinator
	wsm := stateManager.NewWorldStateManager(sldb)
	sme := smartcontract.NewDemoSME()
	cdnt, err = entity.NewCoordinator(cfg.CoordinatorName, cfg.CoordinatorAddress, cfg.PartyName, cfg.OtherCoordinatorAddresses, cfg.WorkerAddresses, cfg.ShardNumbers, wsm, sme)
	if err != nil {
		panic(err)
	}

	// start the API server
	stop := make(chan struct{})
	go StartAPIServer_CDNT(stop, cfg.APIAddress, cdnt)

	if *autoConMode {
		time.Sleep(dialD)
		fmt.Printf("coordinator %s starts to dial...\n", cdnt.NodeName)
		cdnt.StartDial()
		time.Sleep(bindD)
		fmt.Printf("coordinator %s starts to bind peers...\n", cdnt.NodeName)
		if err := cdnt.BindPeers(); err != nil {
			panic(err)
		}
		fmt.Printf("coordinator %s has finished all of connections\n", cdnt.NodeName)
		if err := cdnt.Run(); err != nil {
			panic(err)
		}
		fmt.Printf("coordinator %s has started\n", cdnt.NodeName)
	}

	<-stop
	fmt.Println("Coordinator shutting down...")
	return
}

func StartAPIServer_CDNT(stop chan struct{}, listenAddr string, cdnt *entity.Coordinator) {
	mux := http.NewServeMux()
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Shutting down coordinator...")
		close(stop)
	})

	mux.HandleFunc("/startDial", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cdnt.StartDial()
		fmt.Fprintf(w, "coordinator %s started dialing\n", cdnt.NodeName)
	})

	mux.HandleFunc("/startBind", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := cdnt.BindPeers(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to bind peers: %v", err), http.StatusInternalServerError)
			return
		}

		if err := cdnt.Run(); err != nil {
			panic(err)
		}
		fmt.Printf("coordinator %s has started\n", cdnt.NodeName)

		fmt.Fprintf(w, "coordinator %s successfully bound peers and run\n", cdnt.NodeName)
	})

	mux.HandleFunc("/autoInit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse `waitDelay` parameter
		waitDelayStr := r.URL.Query().Get("waitDelay")
		if waitDelayStr == "" {
			http.Error(w, "Missing required parameter: waitDelay", http.StatusBadRequest)
			return
		}

		waitDelay, err := strconv.Atoi(waitDelayStr)
		if err != nil || waitDelay < 0 {
			http.Error(w, "Invalid waitDelay: must be a non-negative integer", http.StatusBadRequest)
			return
		}

		waitD := time.Duration(waitDelay) * time.Second

		cdnt.StartDial()
		time.Sleep(waitD)
		fmt.Printf("coordinator %s starts to bind peers...\n", cdnt.NodeName)
		if err := cdnt.BindPeers(); err != nil {
			panic(err)
		}
		fmt.Printf("coordinator %s has finished all of connections\n", cdnt.NodeName)
		if err := cdnt.Run(); err != nil {
			panic(err)
		}
		fmt.Printf("coordinator %s has started\n", cdnt.NodeName)
	})

	mux.HandleFunc("/creatAccounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse `amount` parameter
		amountStr := r.URL.Query().Get("amount")
		if amountStr == "" {
			http.Error(w, "Missing required parameter: amount", http.StatusBadRequest)
			return
		}

		amount, err := strconv.Atoi(amountStr)
		if err != nil || amount < 0 {
			http.Error(w, "Invalid amount: must be a non-negative integer", http.StatusBadRequest)
			return
		}

		cdnt.CreatAccounts(amount)
		fmt.Printf("coordinator %s has create %d accounts\n", cdnt.NodeName, amount)
	})

	mux.HandleFunc("/networkInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Convert WorkerShardNOs to uint64
		shardsNOs := make([]uint64, len(cdnt.WorkerShardNOs))
		for i, shardNo := range cdnt.WorkerShardNOs {
			shardsNOs[i] = uint64(shardNo)
		}

		// Build network info struct
		info := publicUsed.CDNT_NETWORK_INFO{
			NodeName:         cdnt.NodeName,
			SelfAddr:         cdnt.SelfAddr,
			Party:            cdnt.Party,
			ConnectedCDNTs:   uint64(len(cdnt.CdntPG.Peers)),
			ConnectedWorkers: uint64(len(cdnt.WorkerPG.Peers)),
			WorkerShard:      shardsNOs,
		}

		// Encode info into RLP bytes
		infoBytes, err := rlp.EncodeToBytes(&info)
		if err != nil {
			http.Error(w, "Failed to encode network info", http.StatusInternalServerError)
			log.Printf("Error encoding network info: %v", err)
			return
		}

		// Write encoded bytes to response
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(infoBytes); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			log.Printf("Error writing response: %v", err)
			return
		}
	})

	mux.HandleFunc("/stateInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stateCode, nowVersion, nowHeight := cdnt.BackCurrentState()
		info := publicUsed.CDNT_STATE_INFO{
			NodeName:   cdnt.NodeName,
			SelfAddr:   cdnt.SelfAddr,
			Party:      cdnt.Party,
			NowState:   stateCode,
			NowVersion: nowVersion,
			NowHeight:  nowHeight,
		}

		// Encode info into RLP bytes
		infoBytes, err := rlp.EncodeToBytes(&info)
		if err != nil {
			http.Error(w, "Failed to encode state info", http.StatusInternalServerError)
			log.Printf("Error encoding state info: %v", err)
			return
		}

		// Write encoded bytes to response
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(infoBytes); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			log.Printf("Error writing response: %v", err)
			return
		}

	})

	mux.HandleFunc("/details", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		content := cdnt.BackAllStateInfo()
		info := publicUsed.DETAIL_INFO{Content: content}

		// Encode info into RLP bytes
		infoBytes, err := rlp.EncodeToBytes(&info)
		if err != nil {
			http.Error(w, "Failed to encode state info", http.StatusInternalServerError)
			log.Printf("Error encoding state info: %v", err)
			return
		}

		// Write encoded bytes to response
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(infoBytes); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			log.Printf("Error writing response: %v", err)
			return
		}
	})

	mux.HandleFunc("/isRunning", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		running := cdnt.Running

		response := map[string]bool{"isRunning": running}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Printf("Error encoding response: %v", err)
			return
		}
	})

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	<-stop

	if err := srv.Close(); err != nil {
		log.Fatalf("API server shutdown failed: %v", err)
	}
}
