package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"opl/clients/publicUsed"
	"opl/clients/testManager/constructTxs"
	"opl/common"
	"opl/elements"
	"opl/oplNet"
	"opl/opt"
	"opl/rlp"
	"os"
	"sync"
	"time"
)

const default_manager_config_path = "./manager_config.yaml"

func main() {
	// Check if a command is provided
	if len(os.Args) < 2 {
		fmt.Println("Error: A command is required")
		printUsage()
		os.Exit(1)
	}

	// Extract the main command
	command := os.Args[1]

	// Handle the command with corresponding FlagSet
	switch command {
	case "printNetwork":
		handle_printNetwork(os.Args[2:])
	case "requireState":
		handle_requireState(os.Args[2:])
	case "requireNetwork":
		handle_requireNetwork(os.Args[2:])
	case "startTest":
		handle_startTest(os.Args[2:])
	case "startTransfer":
		handle_startTransfer(os.Args[2:])
	case "startSort":
		handle_startSort(os.Args[2:])
	case "details":
		handle_details(os.Args[2:])
	case "autoInit":
		handle_autoInit(os.Args[2:])
	case "activeInfo":
		handle_activeInfo(os.Args[2:])
	case "startDial":
		handle_startDial(os.Args[2:])
	case "startBind":
		handle_startBind(os.Args[2:])
	case "creatAccounts":
		handle_creatAccounts(os.Args[2:])

	default:
		fmt.Printf("Error: Unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	// Print usage information for the tool
	fmt.Println("Usage:")
	fmt.Println("  ./tmClient <command> [options]")
	fmt.Println("Commands:")

	// List the commands and their descriptions
	fmt.Println("  printNetwork    - Print the network status.")
	fmt.Println("  requireState    - Retrieve the required state from the node.")
	fmt.Println("  requireNetwork  - Retrieve the required network configuration from the node.")
	fmt.Println("  startTest       - Initiate a new test of SmallBank with parameters specified in manager_config.yaml.")
	fmt.Println("  startTransfer   - Initiate a new test of SmallBank but only has transfer operations with parameters specified in manager_config.yaml.")
	fmt.Println("  startSort       - Initiate a new test of CPUHeavy workload with parameters specified in manager_config.yaml.")
	fmt.Println("  details         - Display detailed system information, typically used for debugging.")
	fmt.Println("  autoInit        - Command all nodes to autonomously establish connections with each other.")
	fmt.Println("  activeInfo      - Query and display the running status of all nodes.")
	fmt.Println("  startDial       - Command all nodes to start dialing peers.")
	fmt.Println("  startBind       - Command all nodes to start binding peers.")
	fmt.Println("  creatAccounts   - Create a specified number of accounts (specified in manager_config.yaml) on all nodes in the network.")

	// Provide examples for how to use each command
	fmt.Println("\nExamples:")
	fmt.Println("  ./tmClient printNetwork")
	fmt.Println("  ./tmClient requireState -addr <node_address>")
	fmt.Println("  ./tmClient requireNetwork -addr <node_address>")
	fmt.Println("  ./tmClient startTest")
	fmt.Println("  ./tmClient startTransfer")
	fmt.Println("  ./tmClient startSort -size <default 100000>")
	fmt.Println("  ./tmClient details -addr <node_address>")
	fmt.Println("  ./tmClient autoInit -delay <seconds>")
	fmt.Println("  ./tmClient activeInfo")
	fmt.Println("  ./tmClient startDial -interval <seconds>")
	fmt.Println("  ./tmClient startBind")
	fmt.Println("  ./tmClient creatAccounts")
}

func handle_startDial(args []string) {
	// Define specific flags for startDial
	fs := flag.NewFlagSet("startDial", flag.ExitOnError)
	interval := fs.Int("interval", 1, "Interval time in seconds between successful startDial operations (default: 1s)")

	// Parse the provided arguments
	fs.Parse(args)

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v\n", err)
		return
	}

	allNodes := make([]oplNet.OPLNode, 0)
	allNodes = append(allNodes, cfg.NodeNetwork.CDNTNodes...)
	for _, shardNodes := range cfg.NodeNetwork.ShardGroups {
		allNodes = append(allNodes, shardNodes...)
	}

	// Iterate over all nodes and send /startDial request
	successCount := 0
	failCount := 0
	for i, node := range allNodes {
		// Construct the full URL for the /startDial endpoint
		fullURL := fmt.Sprintf("http://%s/startDial", node.APIAddr)

		// Create the request
		req, err := http.NewRequest(http.MethodPost, fullURL, nil)
		if err != nil {
			fmt.Printf("Error: Failed to create request for node %s: %v\n", node.APIAddr, err)
			failCount++
			continue
		}

		// Send the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: Failed to send request to node %s: %v\n", node.APIAddr, err)
			failCount++
			continue
		}

		// Ensure the response body is closed
		defer resp.Body.Close()

		// Check if the response status code is 200 OK
		if resp.StatusCode == http.StatusOK {
			successCount++
			fmt.Printf("Node %d (%s): startDial successful\n", i+1, node.APIAddr)

			// Wait the user-specified interval before proceeding to the next node
			time.Sleep(time.Duration(*interval) * time.Second)
		} else {
			failCount++
			fmt.Printf("Node %d (%s): startDial failed with status %d\n", i+1, node.APIAddr, resp.StatusCode)
		}
	}

	// Print the summary
	fmt.Printf("\nSummary:\n")
	fmt.Printf("Successful startDial: %d\n", successCount)
	fmt.Printf("Failed startDial: %d\n", failCount)
}

func handle_startBind(args []string) {
	// Define specific flags for startBind
	fs := flag.NewFlagSet("startBind", flag.ExitOnError)

	// Parse the provided arguments
	fs.Parse(args)

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v\n", err)
		return
	}

	allNodes := make([]oplNet.OPLNode, 0)
	allNodes = append(allNodes, cfg.NodeNetwork.CDNTNodes...)
	for _, shardNodes := range cfg.NodeNetwork.ShardGroups {
		allNodes = append(allNodes, shardNodes...)
	}

	// Iterate over all nodes and send /startBind request
	successCount := 0
	failCount := 0
	for i, node := range allNodes {
		// Construct the full URL for the /startBind endpoint
		fullURL := fmt.Sprintf("http://%s/startBind", node.APIAddr)

		// Create the request
		req, err := http.NewRequest(http.MethodPost, fullURL, nil)
		if err != nil {
			fmt.Printf("Error: Failed to create request for node %s: %v\n", node.APIAddr, err)
			failCount++
			continue
		}

		// Send the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: Failed to send request to node %s: %v\n", node.APIAddr, err)
			failCount++
			continue
		}

		// Ensure the response body is closed
		defer resp.Body.Close()

		// Check if the response status code is 200 OK
		if resp.StatusCode == http.StatusOK {
			successCount++
			fmt.Printf("Node %d (%s): startBind successful\n", i+1, node.APIAddr)
		} else {
			failCount++
			fmt.Printf("Node %d (%s): startBind failed with status %d\n", i+1, node.APIAddr, resp.StatusCode)
		}
	}

	// Print the summary
	fmt.Printf("\nSummary:\n")
	fmt.Printf("Successful startBind: %d\n", successCount)
	fmt.Printf("Failed startBind: %d\n", failCount)
}

func handle_activeInfo(args []string) {
	// Define specific flags for activeInfo
	fs := flag.NewFlagSet("autoDial", flag.ExitOnError)

	// Parse the provided arguments
	fs.Parse(args)

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v", err)
		return
	}

	allNodes := make([]oplNet.OPLNode, 0)
	allNodes = append(allNodes, cfg.NodeNetwork.CDNTNodes...)
	for _, shardNodes := range cfg.NodeNetwork.ShardGroups {
		allNodes = append(allNodes, shardNodes...)
	}

	// Initialize counters for running and not running nodes
	runningCount := 0
	notRunningCount := 0

	// Iterate over all nodes and query their running status
	for i, node := range allNodes {
		// Construct the full URL for the /isRunning endpoint
		fullURL := fmt.Sprintf("http://%s/isRunning", node.APIAddr)

		// Create the request
		req, err := http.NewRequest(http.MethodGet, fullURL, nil)
		if err != nil {
			fmt.Printf("Error: Failed to create request for node %s: %v\n", node.APIAddr, err)
			continue
		}

		// Send the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: Failed to send request to node %s: %v\n", node.APIAddr, err)
			continue
		}

		// Ensure the response body is closed
		defer resp.Body.Close()

		// Check if the response status code is 200 OK
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Warning: Node %s returned status %d\n", node.APIAddr, resp.StatusCode)
			continue
		}

		// Parse the JSON response
		var result map[string]bool
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error: Failed to decode response from node %s: %v\n", node.APIAddr, err)
			continue
		}

		// Extract and print the running status
		isRunning, exists := result["isRunning"]
		if !exists {
			fmt.Printf("Warning: Node %s response does not contain 'isRunning'\n", node.APIAddr)
			continue
		}

		// Update counters based on the running status
		if isRunning {
			runningCount++
		} else {
			notRunningCount++
		}

		fmt.Printf("Node %d (%s): isRunning = %v\n", i+1, node.APIAddr, isRunning)
	}

	// Print the summary of running and not running nodes
	fmt.Printf("\nSummary:\n")
	fmt.Printf("Running nodes: %d\n", runningCount)
	fmt.Printf("Not running nodes: %d\n", notRunningCount)
}

func handle_autoInit(args []string) {
	// Define specific flags for autoInit
	fs := flag.NewFlagSet("autoInit", flag.ExitOnError)

	delay := fs.Int("delay", 10, "Time seconds to wait for binding peers after dialing")

	// Parse the provided arguments
	fs.Parse(args)

	if *delay <= 0 {
		fmt.Println("Error: Delay should be over 0")
		os.Exit(1)
	}

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v", err)
		return
	}

	allNodes := make([]oplNet.OPLNode, 0)
	allNodes = append(allNodes, cfg.NodeNetwork.CDNTNodes...)
	for _, shardNodes := range cfg.NodeNetwork.ShardGroups {
		allNodes = append(allNodes, shardNodes...)
	}

	// Create a WaitGroup to manage concurrent requests
	var wg sync.WaitGroup

	for _, node := range allNodes {
		wg.Add(1)
		fmt.Printf("Start to dial node %s\n", node.OPLAddr)

		// Launch a Goroutine for each request
		go func(node oplNet.OPLNode) {
			defer wg.Done() // Mark as done when the Goroutine finishes

			// Construct the full URL for the /autoInit endpoint
			fullURL := fmt.Sprintf("http://%s/autoInit", node.APIAddr)

			// Create the request with the waitDelay parameter
			req, err := http.NewRequest(http.MethodGet, fullURL, nil)
			if err != nil {
				fmt.Printf("Error: Failed to create request for node %s: %v\n", node.APIAddr, err)
				return
			}

			// Add the waitDelay parameter to the query
			q := req.URL.Query()
			q.Add("waitDelay", fmt.Sprintf("%d", *delay))
			req.URL.RawQuery = q.Encode()

			// Send the request
			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				fmt.Printf("Error: Failed to send request to node %s: %v\n", node.APIAddr, err)
				return
			}
			defer resp.Body.Close() // Ensure the response body is closed

			// Handle the response
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error: Node %s responded with status code %d\n", node.APIAddr, resp.StatusCode)
			} else {
				fmt.Printf("Success: Node %s initialized successfully\n", node.APIAddr)
			}
		}(node)
	}

	// Wait for all Goroutines to complete
	wg.Wait()
	fmt.Println("All nodes processed.")

}

func handle_creatAccounts(args []string) {
	// Define specific flags for creatAccounts
	fs := flag.NewFlagSet("creatAccounts", flag.ExitOnError)

	// Parse the provided arguments
	fs.Parse(args)

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v", err)
		return
	}

	fmt.Printf("The total amount of accounts that should be created is %d\n", cfg.AccountNum)

	allNodes := make([]oplNet.OPLNode, 0)
	allNodes = append(allNodes, cfg.NodeNetwork.CDNTNodes...)
	for _, shardNodes := range cfg.NodeNetwork.ShardGroups {
		allNodes = append(allNodes, shardNodes...)
	}

	// Create a WaitGroup to manage concurrent requests
	var wg sync.WaitGroup

	for _, node := range allNodes {
		wg.Add(1)
		fmt.Printf("Start to inform node %s to creat accounts\n", node.OPLAddr)

		// Launch a Goroutine for each request
		go func(node oplNet.OPLNode) {
			defer wg.Done() // Mark as done when the Goroutine finishes

			// Construct the full URL for the /creatAccounts endpoint
			fullURL := fmt.Sprintf("http://%s/creatAccounts", node.APIAddr)

			// Create the request with the amount parameter
			req, err := http.NewRequest(http.MethodGet, fullURL, nil)
			if err != nil {
				fmt.Printf("Error: Failed to create request for node %s: %v\n", node.APIAddr, err)
				return
			}

			// Add the amount parameter to the query
			q := req.URL.Query()
			q.Add("amount", fmt.Sprintf("%d", cfg.AccountNum))
			req.URL.RawQuery = q.Encode()

			// Send the request
			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				fmt.Printf("Error: Failed to send request to node %s: %v\n", node.APIAddr, err)
				return
			}
			defer resp.Body.Close() // Ensure the response body is closed

			// Handle the response
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error: Node %s responded with status code %d\n", node.APIAddr, resp.StatusCode)
			} else {
				fmt.Printf("Success: Node %s creat accounts successfully\n", node.APIAddr)
			}
		}(node)
	}

	// Wait for all Goroutines to complete
	wg.Wait()
	fmt.Println("All nodes processed.")

}

func handle_details(args []string) {
	// Define specific flags for details
	fs := flag.NewFlagSet("details", flag.ExitOnError)

	addr := fs.String("addr", "nerd", "Used to contact with the node")

	// Parse the provided arguments
	fs.Parse(args)

	// Validate the Address
	if *addr == "" {
		fmt.Println("Error: Address is required")
		os.Exit(1)
	}

	// Construct the full URL for the /stateInfo endpoint
	fullURL := fmt.Sprintf("http://%s/details", *addr)

	// Make a GET request to the server's /stateInfo endpoint
	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Error: Failed to connect to server at %s: %v\n", fullURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check if the server returned an HTTP 200 status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned an error: %s\n", resp.Status)
		os.Exit(1)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: Failed to read server response: %v\n", err)
		os.Exit(1)
	}

	var detailInfo publicUsed.DETAIL_INFO
	if err := rlp.DecodeBytes(body, &detailInfo); err != nil {
		fmt.Printf("Error: Failed to decode server response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(detailInfo.Content)
}

func handle_printNetwork(args []string) {
	// Define specific flags for printNetwork
	fs := flag.NewFlagSet("printNetwork", flag.ExitOnError)

	// Parse the provided arguments
	fs.Parse(args)

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v", err)
		return
	}

	fmt.Println("The network is as follows:")
	fmt.Println(cfg.NodeNetwork.String())
}

func common_test_process(cfg *opt.ManagerCfg, qg *queryGroup, pg *publishGroup, publishFunc func(amount int, accountsAmount int) []elements.Transaction) {
	// Start testing
	minQPS, maxQPS := 500, cfg.TxSendPerSecond
	currentQPS := minQPS

	tp := NewTxPublisher(currentQPS, pg, cfg.TxBatchSize, cfg.AccountNum, publishFunc)
	tp.Start()
	defer tp.Stop()

	fmt.Println("The transactions publisher has started...")
	fmt.Println("Wait for testing....")
	time.Sleep(3 * time.Second)
	fmt.Println("Now begin test!")

	queryTxBatchSize := 1000

	maxRight := -1
	var maxLatency, minLatency, avgLatency time.Duration

	var increment int // increment will be calculated once we approach maxQPS

outer:
	for currentQPS < maxQPS {
		if currentQPS*2 <= maxQPS {
			// Continue doubling QPS until we are getting close to maxQPS
			currentQPS *= 2
		} else {
			// Calculate increment only when we're close to maxQPS
			if increment == 0 {
				increment = (maxQPS - currentQPS) / 5
				if increment < 1 {
					increment = 1 // Ensure the increment is at least 1
				}
			}
			// Start increasing QPS by the calculated increment
			currentQPS += increment
			// Ensure we don't exceed maxQPS
			if currentQPS > maxQPS {
				currentQPS = maxQPS
			}
		}

		tp.UpdateQPS(currentQPS)
		time.Sleep(time.Second) // wait the tx publisher to apply the qps

		txs := publishFunc(queryTxBatchSize, cfg.AccountNum)
		mtc := newMetricTrackContainer()
		go mtc.trackTXsPublish(pg, qg, cfg, txs)

		waitTimer := time.NewTimer(5 * time.Second)
		select {
		case <-mtc.finish:
			if mtc.delayedTxCount == 0 {
				maxRight = currentQPS
				maxLatency, minLatency, avgLatency = mtc.computeTransactionDelay()
			} else {
				break outer
			}
		case <-waitTimer.C:
			fmt.Printf("Over time when QPS is %d\n", currentQPS)
			break outer
		}
	}

	tp.UpdateQPS(maxRight)
	time.Sleep(5 * time.Second) // wait the tx publisher to apply the qps

	fmt.Printf("Now start to compute TPS, the stable QPS is: %d\n", maxRight)
	mtc := newMetricTrackContainer()
	avgTPS, maxTPS, minTPS, inheritCount, reExecCount := mtc.computeTPS(qg, cfg)
	totalCount := inheritCount + reExecCount
	contentionRate := float64(reExecCount) / float64(totalCount)

	// Print results
	fmt.Printf("%d publish clients and %d query clients are involved \n", cfg.PublishNumber, cfg.QueryNumber)
	fmt.Printf("The stable QPS used is %d\n", maxRight)
	fmt.Printf("Metrics:\n")
	fmt.Printf("- Max TPS: %.2f\n", maxTPS)
	fmt.Printf("- Min TPS: %.2f\n", minTPS)
	fmt.Printf("- Avg TPS: %.2f\n", avgTPS)

	fmt.Printf("- Max Latency: %v\n", maxLatency)
	fmt.Printf("- Min Latency: %v\n", minLatency)
	fmt.Printf("- Avg Latency: %v\n", avgLatency)

	fmt.Printf("- Total Transactions: %d\n", totalCount)
	fmt.Printf("- Inherited Transactions: %d\n", inheritCount)
	fmt.Printf("- Re-executed Transactions: %d\n", reExecCount)
	fmt.Printf("- Contention Rate: %.4f \n", contentionRate)
}

func common_test_process_CPUHeavy(cfg *opt.ManagerCfg, qg *queryGroup, pg *publishGroup, size int, publishFunc func(amount int, size int) []elements.Transaction) {
	// Start testing
	minQPS, maxQPS := 100, cfg.TxSendPerSecond
	currentQPS := minQPS

	tp := NewTxPublisher(currentQPS, pg, cfg.TxBatchSize, size, publishFunc)
	tp.Start()
	defer tp.Stop()

	fmt.Println("The transactions publisher has started...")
	fmt.Println("Wait for testing....")
	time.Sleep(3 * time.Second)
	fmt.Println("Now begin test!")

	queryTxBatchSize := 100

	maxRight := -1
	var maxLatency, minLatency, avgLatency time.Duration

	var increment int // increment will be calculated once we approach maxQPS

outer:
	for currentQPS < maxQPS {
		if currentQPS*2 <= maxQPS {
			// Continue doubling QPS until we are getting close to maxQPS
			currentQPS *= 2
		} else {
			// Calculate increment only when we're close to maxQPS
			if increment == 0 {
				increment = (maxQPS - currentQPS) / 5
				if increment < 1 {
					increment = 1 // Ensure the increment is at least 1
				}
			}
			// Start increasing QPS by the calculated increment
			currentQPS += increment
			// Ensure we don't exceed maxQPS
			if currentQPS > maxQPS {
				currentQPS = maxQPS
			}
		}

		tp.UpdateQPS(currentQPS)
		time.Sleep(time.Second) // wait the tx publisher to apply the qps

		txs := publishFunc(queryTxBatchSize, size)
		mtc := newMetricTrackContainer()
		go mtc.trackTXsPublish(pg, qg, cfg, txs)

		waitTimer := time.NewTimer(5 * time.Second)
		select {
		case <-mtc.finish:
			if mtc.delayedTxCount == 0 {
				maxRight = currentQPS
				maxLatency, minLatency, avgLatency = mtc.computeTransactionDelay()
			} else {
				break outer
			}
		case <-waitTimer.C:
			fmt.Printf("Over time when QPS is %d\n", currentQPS)
			break outer
		}
	}

	tp.UpdateQPS(maxRight)
	time.Sleep(5 * time.Second) // wait the tx publisher to apply the qps

	fmt.Printf("Now start to compute TPS, the stable QPS is: %d\n", maxRight)
	mtc := newMetricTrackContainer()
	avgTPS, maxTPS, minTPS, inheritCount, reExecCount := mtc.computeTPS(qg, cfg)
	totalCount := inheritCount + reExecCount
	contentionRate := float64(reExecCount) / float64(totalCount)

	// Print results
	fmt.Printf("%d publish clients and %d query clients are involved \n", cfg.PublishNumber, cfg.QueryNumber)
	fmt.Printf("The stable QPS used is %d\n", maxRight)
	fmt.Printf("Metrics:\n")
	fmt.Printf("- Max TPS: %.2f\n", maxTPS)
	fmt.Printf("- Min TPS: %.2f\n", minTPS)
	fmt.Printf("- Avg TPS: %.2f\n", avgTPS)

	fmt.Printf("- Max Latency: %v\n", maxLatency)
	fmt.Printf("- Min Latency: %v\n", minLatency)
	fmt.Printf("- Avg Latency: %v\n", avgLatency)

	fmt.Printf("- Total Transactions: %d\n", totalCount)
	fmt.Printf("- Inherited Transactions: %d\n", inheritCount)
	fmt.Printf("- Re-executed Transactions: %d\n", reExecCount)
	fmt.Printf("- Contention Rate: %.4f \n", contentionRate)
}

func handle_startTransfer(args []string) {
	// Define specific flags for startTransfer
	fs := flag.NewFlagSet("startTransfer", flag.ExitOnError)

	// Parse the provided arguments
	fs.Parse(args)

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v", err)
		return
	}

	pg, qg := initializeClients(cfg)

	common_test_process(cfg, qg, pg, constructTxs.GenerateTransactions_Transfer)

	fmt.Println("startTransfer command is done")
}

func handle_startTest(args []string) {
	// Define specific flags for startTest
	fs := flag.NewFlagSet("startTest", flag.ExitOnError)

	// Parse the provided arguments
	fs.Parse(args)

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v", err)
		return
	}

	pg, qg := initializeClients(cfg)

	common_test_process(cfg, qg, pg, constructTxs.GenerateTransactions_SmallBank)

	fmt.Println("startTest command is done")
}

func handle_startSort(args []string) {
	// Define specific flags for startSort
	fs := flag.NewFlagSet("startSort", flag.ExitOnError)
	size := fs.Int("size", 100000, "Size is length of the unsorted array (default: 100000)")

	// Parse the provided arguments
	fs.Parse(args)

	cfg, err := opt.LoadManagerConfig(default_manager_config_path)
	if err != nil {
		fmt.Printf("Error: Fail in loading config file %v", err)
		return
	}

	pg, qg := initializeClients(cfg)

	common_test_process_CPUHeavy(cfg, qg, pg, *size, constructTxs.GenerateTransactions_CPUHeavy)

	fmt.Println("startSort command is done")
}

type metricTrackContainer struct {
	finish           chan struct{}
	startTime        time.Time
	startHeight      int
	epochHeights     []int
	epochGotTime     []time.Time
	epochTransations []int
	epochInheritTxs  []int
	txMap            map[common.Hash]int
	txPublishTime    []time.Time
	txGotTime        []time.Time
	txCount          int
	txDelaySum       time.Duration
	delayedTxCount   int
}

func newMetricTrackContainer() *metricTrackContainer {
	return &metricTrackContainer{
		finish:           make(chan struct{}),
		startTime:        time.Time{},
		startHeight:      0,
		epochHeights:     make([]int, 0),
		epochGotTime:     make([]time.Time, 0),
		epochTransations: make([]int, 0),
		epochInheritTxs:  make([]int, 0),
		txMap:            make(map[common.Hash]int),
		txPublishTime:    make([]time.Time, 0),
		txGotTime:        make([]time.Time, 0),
		txCount:          0,
		txDelaySum:       0,
		delayedTxCount:   0,
	}
}

const waitHeights = 21

func (mtc *metricTrackContainer) computeTPS(qg *queryGroup, cfg *opt.ManagerCfg) (averageTPS, maxTPS, minTPS float64, inheritCount, reExecCount int) {

	startHeight, err := qg.getHeight()
	if err != nil {
		panic(err)
	}

	endHeights := startHeight + waitHeights
	nextHeight := startHeight + 1

	queryTicker := time.NewTicker(time.Duration(cfg.QueryInterval) * time.Millisecond)
	for nextHeight <= endHeights {
		<-queryTicker.C
		txIDs, _, got := qg.queryEpoch(nextHeight)
		if got {
			nowTime := time.Now()
			fmt.Printf("got Epoch with height %d, it contain %d transactions\n", nextHeight, len(txIDs))
			mtc.epochGotTime = append(mtc.epochGotTime, nowTime)
			mtc.epochTransations = append(mtc.epochTransations, len(txIDs))
			nextHeight += 1

		}
	}

	// Check if there are enough data points to compute metrics
	if len(mtc.epochGotTime) < 2 || len(mtc.epochTransations) < 2 {
		fmt.Println("Insufficient data to compute metrics")
		return 0, 0, 0, 0, 0
	}

	// Initialize metrics
	var totalTXCount int
	maxTPS = 0
	minTPS = float64(10000000)

	for i := 1; i < len(mtc.epochGotTime); i++ {
		epochDuration := mtc.epochGotTime[i].Sub(mtc.epochGotTime[i-1]).Seconds() // Duration in seconds
		if epochDuration > 0 {
			totalTXCount += mtc.epochTransations[i]
			tps := float64(mtc.epochTransations[i]) / epochDuration
			if tps > maxTPS {
				maxTPS = tps
			}
			if tps < minTPS {
				minTPS = tps
			}
		}
	}

	totalDuration := mtc.epochGotTime[len(mtc.epochGotTime)-1].Sub(mtc.epochGotTime[0]).Seconds()

	averageTPS = float64(totalTXCount) / float64(totalDuration)

	time.Sleep(1 * time.Second)
	fmt.Println("Begin gathering data on transaction processing types.")

	for i := startHeight + 1; i <= endHeights; i++ {
		txIDs := make([]common.Hash, 0)
		txTypes := make([]uint8, 0)
		var got = false

		for try := 0; try < 3; try++ {
			txIDs, txTypes, got = qg.queryEpoch(i)
			if !got {
				time.Sleep(1 * time.Second)
				continue
			} else {
				break
			}
		}

		if !got {
			panic("fail in gathering data on transaction processing types, an observed epoch cannot be got.")
		}

		if len(txIDs) != len(txTypes) {
			panic("fail in gathering data on transaction processing types, the txIDs' length does not match txTypes.")
		}
		for _, txT := range txTypes {
			if txT == 0 {
				inheritCount += 1
			} else if txT == 1 {
				reExecCount += 1
			} else {
				panic(fmt.Errorf("unknown transaction processed type %d is got", txT))
			}
		}
	}

	return averageTPS, maxTPS, minTPS, inheritCount, reExecCount
}

func (mtc *metricTrackContainer) computeTransactionDelay() (time.Duration, time.Duration, time.Duration) {
	var (
		maxLatency    time.Duration
		minLatency    time.Duration = time.Hour // Initialize to a large value
		totalLatency  time.Duration
		latencyCounts int
	)

	for i := 0; i < len(mtc.txPublishTime); i++ {
		if !mtc.txPublishTime[i].IsZero() && !mtc.txGotTime[i].IsZero() {
			latency := mtc.txGotTime[i].Sub(mtc.txPublishTime[i])
			totalLatency += latency
			latencyCounts++
			if latency > maxLatency {
				maxLatency = latency
			}
			if latency < minLatency {
				minLatency = latency
			}
		}
	}

	// Compute average latency
	var avgLatency time.Duration
	if latencyCounts > 0 {
		avgLatency = totalLatency / time.Duration(latencyCounts)
	} else {
		fmt.Println("No valid transactions to compute latency")
	}

	return maxLatency, minLatency, avgLatency
}

func (mtc *metricTrackContainer) trackTXsPublish(pg *publishGroup, qg *queryGroup, cfg *opt.ManagerCfg, transactions []elements.Transaction) {
	for i, tx := range transactions {
		mtc.txMap[tx.TxID] = i
	}
	mtc.txGotTime = make([]time.Time, len(transactions))

	startHeight, err := qg.getHeight()
	if err != nil {
		panic(err)
	}

	fmt.Printf("The start height is %d\n", startHeight)

	mtc.startTime = time.Now()

	txsPubTime := make(chan []time.Time)

	go func() {
		txsPubTime <- pg.publishTxs(cfg.TxBatchSize, transactions)
	}()

	mtc.startHeight = startHeight
	nextHeight := startHeight + 1

	queryTicker := time.NewTicker(time.Duration(cfg.QueryInterval) * time.Millisecond)
	for {
		fmt.Printf("Now nextHeight is %d\n", nextHeight)
		<-queryTicker.C
		txIDs, _, got := qg.queryEpoch(nextHeight)
		if got {
			nowTime := time.Now()
			for _, txID := range txIDs {
				index, have := mtc.txMap[txID]
				if have {
					mtc.txGotTime[index] = nowTime
					delete(mtc.txMap, txID)
					if (nextHeight - startHeight) > cfg.MaxEpochDelay {
						mtc.delayedTxCount += 1
					}
				} else {
					// do nothing
				}
			}

			nextHeight += 1
		} else {
			// do nothing
		}

		if len(mtc.txMap) == 0 { // all transactions have got
			break
		}
	}

	mtc.txPublishTime = <-txsPubTime
	mtc.finish <- struct{}{}
}

type TxPublisher struct {
	QPS              int
	mu               sync.Mutex
	stopChan         chan struct{}
	wg               sync.WaitGroup
	pg               *publishGroup
	batchSize        int
	accountNumOrSize int
	publishFunc      func(amount int, accountNum int) []elements.Transaction
}

func NewTxPublisher(initialQPS int, pg *publishGroup, batchSize int, accountNumOrSize int, publishFunc func(amount int, accountNumOrSize int) []elements.Transaction) *TxPublisher {
	return &TxPublisher{
		QPS:              initialQPS,
		mu:               sync.Mutex{},
		stopChan:         make(chan struct{}),
		wg:               sync.WaitGroup{},
		pg:               pg,
		batchSize:        batchSize,
		accountNumOrSize: accountNumOrSize,
		publishFunc:      publishFunc,
	}
}

func (tp *TxPublisher) Start() {
	tp.wg.Add(1)
	go func() {
		defer tp.wg.Done()
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-tp.stopChan:
				return
			case <-ticker.C:
				tp.mu.Lock()
				amount := tp.QPS / 4
				tp.mu.Unlock()

				txs := tp.publishFunc(amount, tp.accountNumOrSize)
				tp.pg.publishTxs(tp.batchSize, txs)
			}
		}
	}()
}

func (tp *TxPublisher) Stop() {
	close(tp.stopChan)
	tp.wg.Wait()
}

// UpdateQPS updates the QPS for the TransactionManager.
func (tp *TxPublisher) UpdateQPS(newQPS int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	fmt.Printf("Updating QPS from %d to %d\n", tp.QPS, newQPS)
	tp.QPS = newQPS
}

func handle_requireState(args []string) {
	// Define specific flags for requireState
	fs := flag.NewFlagSet("requireState", flag.ExitOnError)

	addr := fs.String("addr", "nerd", "Used to contact with the node")

	// Parse the provided arguments
	fs.Parse(args)

	// Validate the Address
	if *addr == "" {
		fmt.Println("Error: Address is required")
		os.Exit(1)
	}

	// Construct the full URL for the /stateInfo endpoint
	fullURL := fmt.Sprintf("http://%s/stateInfo", *addr)

	// Make a GET request to the server's /stateInfo endpoint
	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Error: Failed to connect to server at %s: %v\n", fullURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check if the server returned an HTTP 200 status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned an error: %s\n", resp.Status)
		os.Exit(1)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: Failed to read server response: %v\n", err)
		os.Exit(1)
	}

	var cdntStateInfo publicUsed.CDNT_STATE_INFO
	if err := rlp.DecodeBytes(body, &cdntStateInfo); err == nil {
		fmt.Println("The node required is coordinator, and the state is as follows:")
		fmt.Println(cdntStateInfo.String())
		return
	}

	var workerStateInfo publicUsed.WORKER_STATE_INFO
	if err := rlp.DecodeBytes(body, &workerStateInfo); err == nil {
		fmt.Println("The node required is worker, and the state is as follows:")
		fmt.Println(workerStateInfo.String())
		return
	}

	fmt.Println("Error: Filed to parse the response into a known state format")
}

func handle_requireNetwork(args []string) {
	// Define specific flags for requireNetwork
	fs := flag.NewFlagSet("requireNetwork", flag.ExitOnError)

	addr := fs.String("addr", "nerd", "Used to contact with the node")

	// Parse the provided arguments
	fs.Parse(args)

	// Validate the Address
	if *addr == "" {
		fmt.Println("Error: Address is required")
		os.Exit(1)
	}

	// Construct the full URL for the /networkInfo endpoint
	fullURL := fmt.Sprintf("http://%s/networkInfo", *addr)

	// Make a GET request to the server's /stateInfo endpoint
	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Error: Failed to connect to server at %s: %v\n", fullURL, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check if the server returned an HTTP 200 status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned an error: %s\n", resp.Status)
		os.Exit(1)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: Failed to read server response: %v\n", err)
		os.Exit(1)
	}

	var cdntNetworkInfo publicUsed.CDNT_NETWORK_INFO
	if err := rlp.DecodeBytes(body, &cdntNetworkInfo); err == nil {
		fmt.Println("The node required is coordinator, and the network is as follows:")
		fmt.Println(cdntNetworkInfo.String())
		return
	}

	var workerNetworkInfo publicUsed.WORKER_NETWORK_INFO
	if err := rlp.DecodeBytes(body, &workerNetworkInfo); err == nil {
		fmt.Println("The node required is worker, and the network is as follows:")
		fmt.Println(workerNetworkInfo.String())
		return
	}
	fmt.Println("Error: Filed to parse the response into a known network format")
}

type txPublishClient struct {
	apiAddr string
	client  *http.Client
}

func newTxPublishClient(addr string) *txPublishClient {
	return &txPublishClient{
		apiAddr: addr,
		client:  &http.Client{},
	}
}

type epochQueryClient struct {
	apiAddr string
	client  *http.Client
}

func newEpochQueryClient(addr string) *epochQueryClient {
	return &epochQueryClient{
		apiAddr: addr,
		client:  &http.Client{},
	}
}

func (tpc *txPublishClient) publishTransactions(txs []elements.Transaction) error {
	// Construct the full URL for the /publishTXs endpoint
	fullURL := fmt.Sprintf("http://%s/publishTXs", tpc.apiAddr)

	txGroup := publicUsed.TX_GROUP_MSG{TXS: txs}
	// RLP encode the TX_GROUP_MSG
	encodedData, err := rlp.EncodeToBytes(txGroup)
	if err != nil {
		log.Fatalf("Failed to RLP encode TX_GROUP_MSG: %v", err)
		return err
	}

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewBuffer(encodedData))
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := tpc.client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send HTTP request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Handle the response
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server returned an error: %s", resp.Status)
	}

	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Failed to decode server response: %v", err)
		return err
	}

	// Print the server response
	//fmt.Printf("Response: %v\n", response)

	return nil
}

func (eqc *epochQueryClient) queryEpoch(height int) (txIDs []common.Hash, txTypes []uint8, got bool) {
	// Construct the full URL for the /getTxIDsByHeight endpoint
	fullURL := fmt.Sprintf("http://%s/getTxIDsByHeight?height=%d", eqc.apiAddr, height)

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
		return nil, nil, false
	}

	resp, err := eqc.client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send HTTP request: %v", err)
		return nil, nil, false
	}
	defer resp.Body.Close()

	// Handle the response
	if resp.StatusCode != http.StatusOK {
		log.Printf("Server returned an error: %s", resp.Status)
		return nil, nil, false
	}

	// Read and decode the RLP-encoded response
	var tbh publicUsed.TXIDS_BY_HEIGHT
	if err := rlp.Decode(resp.Body, &tbh); err != nil {
		log.Printf("Failed to decode RLP response: %v", err)
		return nil, nil, false
	}

	// Convert GOT field to boolean
	got = tbh.GOT == 1

	//log.Printf("The txtypes are: %v", tbh.TXTYPES)
	// Return transaction IDs and got flag
	return tbh.TXIDS, tbh.TXTYPES, got
}

// Initialize publish and query clients
func initializeClients(cfg *opt.ManagerCfg) (*publishGroup, *queryGroup) {
	publishClients := make([]*txPublishClient, cfg.PublishNumber)
	queryClients := make([]*epochQueryClient, cfg.QueryNumber)

	totIndex := 0
	shardAmount := len(cfg.NodeNetwork.ShardNOs)
	partyAmount := len(cfg.NodeNetwork.PartyIndex)

	for i := 0; i < cfg.PublishNumber; i++ {
		partyIndex := (totIndex / shardAmount) % partyAmount
		shardIndex := totIndex % shardAmount
		publishClients[i] = newTxPublishClient(cfg.NodeNetwork.ShardGroups[shardIndex][partyIndex].APIAddr)
		totIndex++
	}

	for i := 0; i < cfg.QueryNumber; i++ {
		partyIndex := (totIndex / shardAmount) % partyAmount
		shardIndex := totIndex % shardAmount
		queryClients[i] = newEpochQueryClient(cfg.NodeNetwork.ShardGroups[shardIndex][partyIndex].APIAddr)
		totIndex++
	}

	pg := newPublishGroup(publishClients)
	qg := newQueryGroup(queryClients)
	return pg, qg

}

func (tpc *txPublishClient) getHeight() (int, error) {
	// Construct the full URL for the /stateInfo endpoint
	fullURL := fmt.Sprintf("http://%s/stateInfo", tpc.apiAddr)

	// Make a GET request to the server's /stateInfo endpoint
	resp, err := http.Get(fullURL)
	if err != nil {
		return -1, fmt.Errorf("failed to connect to server at %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	// Check if the server returned an HTTP 200 status
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("server returned an error (HTTP %d): %s", resp.StatusCode, resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("failed to read server response: %w", err)
	}

	// Decode the response into a worker state information
	var workerStateInfo publicUsed.WORKER_STATE_INFO
	if err = rlp.DecodeBytes(body, &workerStateInfo); err != nil {
		return -1, fmt.Errorf("failed to decode response into WORKER_STATE_INFO: %w", err)
	}

	// Return the height
	return int(workerStateInfo.NowHeight), nil
}

func (eqc *epochQueryClient) getHeight() (int, error) {
	// Construct the full URL for the /stateInfo endpoint
	fullURL := fmt.Sprintf("http://%s/stateInfo", eqc.apiAddr)

	// Make a GET request to the server's /stateInfo endpoint
	resp, err := http.Get(fullURL)
	if err != nil {
		return -1, fmt.Errorf("failed to connect to server at %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	// Check if the server returned an HTTP 200 status
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("server returned an error (HTTP %d): %s", resp.StatusCode, resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("failed to read server response: %w", err)
	}

	// Decode the response into a worker state information
	var workerStateInfo publicUsed.WORKER_STATE_INFO
	if err = rlp.DecodeBytes(body, &workerStateInfo); err != nil {
		return -1, fmt.Errorf("failed to decode response into WORKER_STATE_INFO: %w", err)
	}

	// Return the height
	return int(workerStateInfo.NowHeight), nil
}

type publishGroup struct {
	clients []*txPublishClient
}

func newPublishGroup(clients []*txPublishClient) *publishGroup {
	return &publishGroup{clients: clients}
}

type queryGroup struct {
	nowTurn int
	clients []*epochQueryClient
}

func newQueryGroup(clients []*epochQueryClient) *queryGroup {
	return &queryGroup{
		nowTurn: 0,
		clients: clients,
	}
}

func (pg *publishGroup) publishTxs(maxBatchSize int, txs []elements.Transaction) (txsPubTime []time.Time) {
	var wg sync.WaitGroup
	var mu sync.Mutex // Protect access to txsPubTime

	// Initialize the result slice
	txsPubTime = make([]time.Time, len(txs))

	// Calculate how many transactions each client should handle
	numClients := len(pg.clients)
	if numClients == 0 {
		fmt.Println("No clients available")
		return
	}

	// Distribute transactions to each client
	batchSize := len(txs) / numClients
	if len(txs)%numClients != 0 {
		batchSize++ // Handle the remainder
	}

	for i, client := range pg.clients {
		startIdx := i * batchSize

		if startIdx >= len(txs) {
			continue
		}

		endIdx := startIdx + batchSize
		if endIdx > len(txs) {
			endIdx = len(txs)
		}

		clientTxs := txs[startIdx:endIdx]

		for len(clientTxs) > 0 {
			currentBatchSize := maxBatchSize
			if len(clientTxs) < maxBatchSize {
				currentBatchSize = len(clientTxs)
			}

			wg.Add(1)
			go func(txsBatch []elements.Transaction, batchStartIdx int) {
				defer wg.Done()

				// Record the publish time for each transaction in this batch
				publishTime := time.Now()
				err := client.publishTransactions(txsBatch)
				if err != nil {
					fmt.Println("Error publishing transactions:", err)
					return
				}

				// Save the publish times
				mu.Lock()
				for idx := 0; idx < len(txsBatch); idx++ {
					txsPubTime[batchStartIdx+idx] = publishTime
				}
				mu.Unlock()
			}(clientTxs[:currentBatchSize], startIdx)

			clientTxs = clientTxs[currentBatchSize:]
		}
	}

	wg.Wait()
	return txsPubTime
}

// queryEpoch queries the epoch for the specified height using the client determined by nowTurn
func (qg *queryGroup) queryEpoch(height int) (txIDs []common.Hash, txTypes []uint8, got bool) {
	// Ensure there are clients in the group
	if len(qg.clients) == 0 {
		fmt.Println("No clients available for querying.")
		return nil, nil, false
	}

	// Select the client based on nowTurn, and wrap around using modulo
	client := qg.clients[qg.nowTurn]

	// Query the epoch from the selected client
	txIDs, txTypes, got = client.queryEpoch(height)

	// Update nowTurn for the next query (circular increment)
	qg.nowTurn = (qg.nowTurn + 1) % len(qg.clients)

	return txIDs, txTypes, got
}

func (qg *queryGroup) getHeight() (int, error) {
	// Ensure there are clients in the group
	if len(qg.clients) == 0 {
		return 0, fmt.Errorf("No clients available for querying.")
	}
	// Select the client based on nowTurn, and wrap around using modulo
	client := qg.clients[qg.nowTurn]

	// Update nowTurn for the next query (circular increment)
	qg.nowTurn = (qg.nowTurn + 1) % len(qg.clients)

	height, err := client.getHeight()
	if err != nil {
		return 0, err
	}

	return height, nil

}
