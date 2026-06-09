package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
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
	configPath := flag.String("config", "./worker_config.yaml", "Path to the configuration file for initialization")
	autoConMode := flag.Bool("autoconn", false, "Whether to autonomously dial other nodes and bind peers")
	dialDelay := flag.Int("dialDelay", 5, "Dial delay after initialize the worker in seconds")
	bindDelay := flag.Int("bindDelay", 10, "Bind delay after dialing all nodes in seconds")

	flag.Parse()

	dialD := time.Duration(*dialDelay) * time.Second
	bindD := time.Duration(*bindDelay) * time.Second

	cfg, err := opt.LoadWorkerConfig(*configPath)
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
		sldb = database.NewSimpleLDB("DB_"+cfg.WorkerName, db)
	}
	defer sldb.Close()

	var worker *entity.Worker
	wsm := stateManager.NewWorldStateManager(sldb)
	sme := smartcontract.NewDemoSME()
	worker, err = entity.NewWorker(cfg.WorkerName, cfg.WorkerAddress, cfg.AssociatedCoordinatorAddress, cfg.PartyName, cfg.ShardNumber, cfg.OtherWorkerAddresses, wsm, sme)
	if err != nil {
		panic(err)
	}

	// start the API server
	stop := make(chan struct{})
	go StartAPIServer_WORKER(stop, cfg.APIAddress, worker)

	if *autoConMode {
		time.Sleep(dialD)
		fmt.Printf("worker %s starts to dial...\n", worker.NodeName)
		worker.StartDial()
		time.Sleep(bindD)
		fmt.Printf("worker %s starts to bind peers...\n", worker.NodeName)
		if err := worker.BindPeers(); err != nil {
			panic(err)
		}
		fmt.Printf("worker %s has finished all of connections\n", worker.NodeName)
		if err := worker.Run(); err != nil {
			panic(err)
		}
		fmt.Printf("worker %s has started\n", worker.NodeName)
	}

	<-stop
	fmt.Println("Worker shutting down...")
	return
}

func StartAPIServer_WORKER(stop chan struct{}, listenAddr string, worker *entity.Worker) {

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

		worker.StartDial()
		fmt.Fprintf(w, "worker %s started dialing\n", worker.NodeName)
	})

	mux.HandleFunc("/startBind", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := worker.BindPeers(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to bind peers: %v", err), http.StatusInternalServerError)
			return
		}

		if err := worker.Run(); err != nil {
			panic(err)
		}
		fmt.Printf("worker %s has started\n", worker.NodeName)

		fmt.Fprintf(w, "worker %s successfully bound peers and run\n", worker.NodeName)
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

		worker.StartDial()
		time.Sleep(waitD)
		fmt.Printf("worker %s starts to bind peers...\n", worker.NodeName)

		if err := worker.BindPeers(); err != nil {
			panic(err)
		}
		fmt.Printf("worker %s has finished all of connections\n", worker.NodeName)
		if err := worker.Run(); err != nil {
			panic(err)
		}
		fmt.Printf("worker %s has started\n", worker.NodeName)
	})

	mux.HandleFunc("/networkInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Build network info struct
		info := publicUsed.WORKER_NETWORK_INFO{
			NodeName:          worker.NodeName,
			SelfAddr:          worker.SelfAddr,
			Party:             worker.Party,
			ConnectedCDNTAddr: worker.CdntAddr,
			ConnectedWorkers:  uint64(len(worker.ShardPG.Peers)),
			ShardNumber:       uint64(worker.Shard),
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

		stateCode, nowVersion, nowHeight := worker.BackCurrentState()
		info := publicUsed.WORKER_STATE_INFO{
			NodeName:    worker.NodeName,
			SelfAddr:    worker.SelfAddr,
			Party:       worker.Party,
			ShardNumber: uint64(worker.Shard),
			NowState:    stateCode,
			NowVersion:  nowVersion,
			NowHeight:   nowHeight,
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

	mux.HandleFunc("/publishTX", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			log.Printf("Error reading request body: %v", err)
			return
		}
		defer r.Body.Close()

		var txMsg publicUsed.TX_MSG
		if err := rlp.DecodeBytes(body, &txMsg); err != nil {
			http.Error(w, "Failed to decode RLP data", http.StatusBadRequest)
			log.Printf("Error decoding RLP data: %v", err)
			return
		}

		err = worker.GetTransaction(txMsg.TX)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to retrieve transaction: %v", err), http.StatusInternalServerError)
			log.Printf("Error retrieving transaction: %v", err)
			return
		}

		succMsg := fmt.Sprintf("Transaction %x published successfully", txMsg.TX.TxID)

		response := map[string]string{
			"status":  "success",
			"message": succMsg,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Printf("Error encoding response: %v", err)
			return
		}
	})

	mux.HandleFunc("/publishTXs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			log.Printf("Error reading request body: %v", err)
			return
		}
		defer r.Body.Close()

		var txsMsg publicUsed.TX_GROUP_MSG
		if err := rlp.DecodeBytes(body, &txsMsg); err != nil {
			http.Error(w, "Failed to decode RLP data", http.StatusBadRequest)
			log.Printf("Error decoding RLP data: %v", err)
			return
		}

		err = worker.GetTransactionGroup(txsMsg.TXS)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to retrieve transaction group: %v", err), http.StatusInternalServerError)
			log.Printf("Error retrieving transactions: %v", err)
			return
		}

		succMsg := fmt.Sprintf("%d have been published successfully", len(txsMsg.TXS))

		response := map[string]string{
			"status":  "success",
			"message": succMsg,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Printf("Error encoding response: %v", err)
			return
		}
	})

	mux.HandleFunc("/getTxIDsByHeight", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		heightStr := r.URL.Query().Get("height")
		if heightStr == "" {
			http.Error(w, "Height parameter is missing", http.StatusBadRequest)
			return
		}

		height, err := strconv.Atoi(heightStr)
		if err != nil {
			http.Error(w, "Invalid height parameter", http.StatusBadRequest)
			return
		}

		txIDs, txTypes, have := worker.GetTxIDsAndTypesInEpochByHeight(height)
		got := uint64(0)
		if have {
			got = 1
		}

		tbh := publicUsed.TXIDS_BY_HEIGHT{
			GOT:     got,
			TXIDS:   txIDs,
			TXTYPES: txTypes,
		}

		tbh_bytes, err := rlp.EncodeToBytes(&tbh)
		if err != nil {
			http.Error(w, "Failed to encode bytes", http.StatusInternalServerError)
			log.Printf("Error encoding tbh_bytes: %v", err)
			return
		}

		// Write encoded bytes to response
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write(tbh_bytes); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			log.Printf("Error writing response: %v", err)
			return
		}

		//log.Printf("Successfully processed request for height %d", height)
	})

	mux.HandleFunc("/details", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		content := worker.BackAllStateInfo()
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

		running := worker.Running

		response := map[string]bool{"isRunning": running}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Printf("Error encoding response: %v", err)
			return
		}
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

		worker.CreatAccounts(amount)
		fmt.Printf("Worker %s has create %d accounts\n", worker.NodeName, amount)
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
