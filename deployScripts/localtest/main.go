package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"opl/oplNet"
	"os"
	"path/filepath"
)

type Config struct {
	OPLNetwork oplNet.OPLNetwork `yaml:"opl_network"`
	MemDBMode  bool              `yaml:"mem_db_mode"`
	DialDelay  int               `yaml:"dial_delay"`
	BindDelay  int               `yaml:"bind_delay"`
}

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("usage: %s <localtest directory>", os.Args[0]))
	}
	localtestDir := os.Args[1]

	config, err := readConfig("deployScripts/localtest/localtest.yaml")

	if err != nil {
		panic(fmt.Errorf("error reading config: %v", err))
	}

	autoDir := filepath.Join(localtestDir, "auto")
	err = os.MkdirAll(autoDir, 0755)
	if err != nil {
		panic(fmt.Errorf("error creating auto directory: %v", err))
	}

	err = setParties(localtestDir, autoDir, config)
	if err != nil {
		panic(fmt.Errorf("error setting coordinators: %v", err))
	}
	err = setManager(localtestDir, autoDir, config)
	if err != nil {
		panic(fmt.Errorf("error setting manager: %v", err))
	}
	err = createScripts(autoDir, config)
	if err != nil {
		panic(fmt.Errorf("error creating start/stop scripts: %v", err))
	}
}

func setParties(localtestDir, autoDir string, cfg *Config) error {
	for i, party := range cfg.OPLNetwork.PartyIndex {
		partyDir := filepath.Join(autoDir, fmt.Sprintf("PG_%s", party))
		err := os.Mkdir(partyDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating nodes directory: %v", err)
		}

		err = setCdnt(localtestDir, partyDir, cfg, i)
		if err != nil {
			return err
		}
		err = setWorkers(localtestDir, partyDir, cfg, i)
		if err != nil {
			return err
		}
		createPartyScripts(partyDir, cfg)
	}
	return nil
}

func setCdnt(localtestDir, partyDir string, cfg *Config, partyIndex int) error {
	nodeDir := filepath.Join(partyDir, fmt.Sprintf("coordinator_node"))
	err := os.Mkdir(nodeDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating node directory: %v", err)
	}

	coordinator := cfg.OPLNetwork.CDNTNodes[partyIndex]

	dbPath := "db"

	if !cfg.MemDBMode {
		dbDir := filepath.Join(nodeDir, "db")
		err = os.Mkdir(dbDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating db directory: %v", err)
		}
	}

	coordinatorConfig := map[string]interface{}{
		"coordinator_name":            fmt.Sprintf("coordinator_%s", cfg.OPLNetwork.PartyIndex[partyIndex]),
		"coordinator_address":         coordinator.OPLAddr,
		"party_name":                  cfg.OPLNetwork.PartyIndex[partyIndex],
		"other_coordinator_addresses": getOtherCoordinatorAddresses(cfg.OPLNetwork.CDNTNodes, partyIndex),
		"worker_addresses":            getWorkerAddresses(cfg.OPLNetwork.ShardGroups, partyIndex),
		"shard_numbers":               cfg.OPLNetwork.ShardNOs,
		"mem_db_mode":                 cfg.MemDBMode,
		"database_path":               dbPath,
		"api_address":                 coordinator.APIAddr,
	}

	configData, err := yaml.Marshal(&coordinatorConfig)
	if err != nil {
		return fmt.Errorf("error marshaling coordinator config: %v", err)
	}

	err = os.WriteFile(filepath.Join(nodeDir, "coordinator_config.yaml"), configData, 0755)
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	err = copyFile(filepath.Join(localtestDir, "cdntClient"), filepath.Join(nodeDir, "cdntClient"))
	if err != nil {
		return fmt.Errorf("error copying client file: %v", err)
	}

	return nil
}

func setWorkers(localtestDir, partyDir string, cfg *Config, partyIndex int) error {
	for i, shardNO := range cfg.OPLNetwork.ShardNOs {
		nodeDir := filepath.Join(partyDir, fmt.Sprintf("worker_node_shard_%d", shardNO))
		err := os.Mkdir(nodeDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating node directory: %v", err)
		}

		worker := cfg.OPLNetwork.ShardGroups[i][partyIndex]

		dbPath := "db"

		if !cfg.MemDBMode {
			dbDir := filepath.Join(nodeDir, "db")
			err = os.Mkdir(dbDir, 0755)
			if err != nil {
				return fmt.Errorf("error creating db directory: %v", err)
			}
		}

		// Prepare the worker configuration
		workerConfig := map[string]interface{}{
			"worker_name":                    fmt.Sprintf("worker_%s_shard_%d", cfg.OPLNetwork.PartyIndex[partyIndex], shardNO),
			"worker_address":                 worker.OPLAddr,
			"party_name":                     cfg.OPLNetwork.PartyIndex[partyIndex],
			"shard_number":                   shardNO,
			"associated_coordinator_address": cfg.OPLNetwork.CDNTNodes[partyIndex].OPLAddr,
			"other_worker_addresses":         getOtherWorkerAddresses(cfg.OPLNetwork.ShardGroups[i], partyIndex),
			"mem_db_mode":                    cfg.MemDBMode,
			"database_path":                  dbPath,
			"api_address":                    worker.APIAddr,
		}

		// Marshal the worker configuration into YAML format
		configData, err := yaml.Marshal(&workerConfig)
		if err != nil {
			return fmt.Errorf("error marshaling worker config: %v", err)
		}

		// Write the worker configuration to a YAML file
		err = os.WriteFile(filepath.Join(nodeDir, "worker_config.yaml"), configData, 0755)
		if err != nil {
			return fmt.Errorf("error writing config file: %v", err)
		}

		// Copy the worker client binary from localtestDir to the worker node directory
		err = copyFile(filepath.Join(localtestDir, "workerClient"), filepath.Join(nodeDir, "workerClient"))
		if err != nil {
			return fmt.Errorf("error copying client file: %v", err)
		}
	}
	return nil
}

func setManager(localtestDir, autoDir string, cfg *Config) error {
	managerDir := filepath.Join(autoDir, "manager_node")

	err := os.Mkdir(managerDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating node directory: %v", err)
	}

	managerConfig := map[string]interface{}{
		"node_network": map[string]interface{}{
			"party_index":  cfg.OPLNetwork.PartyIndex,
			"cdnt_nodes":   cfg.OPLNetwork.CDNTNodes,
			"shard_nos":    cfg.OPLNetwork.ShardNOs,
			"shard_groups": cfg.OPLNetwork.ShardGroups,
		},
		"publish_number":     len(cfg.OPLNetwork.ShardNOs),
		"query_number":       len(cfg.OPLNetwork.ShardNOs),
		"tx_batch_size":      1024,
		"tx_send_per_second": 10000,
		"max_epoch_delay":    4,
		"query_interval":     100,
		"account_num":        100000,
	}

	configData, err := yaml.Marshal(&managerConfig)
	if err != nil {
		return fmt.Errorf("error marshaling manager config: %v", err)
	}

	err = os.WriteFile(filepath.Join(managerDir, "manager_config.yaml"), configData, 0755)
	if err != nil {
		return fmt.Errorf("error writing manager config file: %v", err)
	}

	err = copyFile(filepath.Join(localtestDir, "tmClient"), filepath.Join(managerDir, "tmClient"))
	if err != nil {
		return fmt.Errorf("error copying manager client file: %v", err)
	}
	return nil
}

func readConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	err = config.OPLNetwork.Check()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func getOtherCoordinatorAddresses(coordinators []oplNet.OPLNode, excludeIndex int) []string {
	var addresses []string
	for i, coordinator := range coordinators {
		if i != excludeIndex {
			addresses = append(addresses, coordinator.OPLAddr)
		}
	}
	return addresses
}

func getOtherWorkerAddresses(workers []oplNet.OPLNode, excludeIndex int) []string {
	var addresses []string
	for i, worker := range workers {
		if i != excludeIndex {
			addresses = append(addresses, worker.OPLAddr)
		}
	}
	return addresses
}

func getWorkerAddresses(shardGoups [][]oplNet.OPLNode, selfIndex int) []string {
	addresses := make([]string, len(shardGoups))
	for i := 0; i < len(shardGoups); i++ {
		addresses[i] = shardGoups[i][selfIndex].OPLAddr
	}
	return addresses
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("error reading source file: %v", err)
	}

	err = os.WriteFile(dst, data, 0755)
	if err != nil {
		return fmt.Errorf("error writing destination file: %v", err)
	}

	return nil
}

func createPartyScripts(partyDir string, cfg *Config) error {
	startScript := filepath.Join(partyDir, "startAll.sh")
	stopScript := filepath.Join(partyDir, "stopAll.sh")
	resetScript := filepath.Join(partyDir, "resetDB.sh")

	coordinatorStartContent := fmt.Sprintf(`#!/bin/bash
(cd coordinator_node && ./cdntClient -autoconn -dialDelay %d -bindDelay %d &)
`, cfg.DialDelay, cfg.BindDelay)

	workerStartContent := fmt.Sprintf(`#!/bin/bash
for d in $(find . -type d -name 'worker_node_shard_*'); do
    (cd $d && ./workerClient -autoconn -dialDelay %d -bindDelay %d &)
done
`, cfg.DialDelay, cfg.BindDelay)

	startContent := coordinatorStartContent + workerStartContent

	err := os.WriteFile(startScript, []byte(startContent), 0755)
	if err != nil {
		return fmt.Errorf("Error creating start script: %v", err)
	}

	err = os.WriteFile(stopScript, []byte(`#!/bin/bash
killall cdntClient workerClient
`), 0755)
	if err != nil {
		return fmt.Errorf("Error creating stop script: %v", err)
	}

	resetContent := `#!/bin/bash

# Process the coordinator node (assumes there is only one)
coordinator="coordinator_node"
if [ -d "$coordinator/db" ]; then
  echo "Processing $coordinator..."
  echo "Deleting $coordinator/db..."
  rm -rf "$coordinator/db"
fi
echo "Creating new $coordinator/db..."
mkdir -p "$coordinator/db"
echo "$coordinator db reset completed."

# Find and process all worker nodes dynamically
worker_nodes=$(find . -type d -name 'worker_node_shard_*')
for worker in $worker_nodes; do
  echo "Processing $worker..."

  # Define the path to the db folder within the worker node
  db_path="$worker/db"

  # Check if the db folder exists
  if [ -d "$db_path" ]; then
    echo "Deleting $db_path..."
    rm -rf "$db_path"
  fi

  # Create a new db folder
  echo "Creating new $db_path..."
  mkdir -p "$db_path"

  echo "$worker db reset completed."
done

echo "All nodes db reset completed."

`
	err = os.WriteFile(resetScript, []byte(resetContent), 0755)
	if err != nil {
		return fmt.Errorf("Error creating reset script: %v", err)
	}

	return nil
}

func createScripts(autoDir string, cfg *Config) error {
	startScript := filepath.Join(autoDir, "startAll.sh")
	stopScript := filepath.Join(autoDir, "stopAll.sh")
	resetScript := filepath.Join(autoDir, "resetAll.sh")

	partyStartContent := fmt.Sprintf(`#!/bin/bash
for d in $(find . -type d -name 'PG_*'); do
    (cd $d && ./startAll.sh &)
done
`)

	err := os.WriteFile(startScript, []byte(partyStartContent), 0755)
	if err != nil {
		return fmt.Errorf("Error creating start script: %v", err)
	}

	err = os.WriteFile(stopScript, []byte(`#!/bin/bash
killall cdntClient workerClient
`), 0755)
	if err != nil {
		return fmt.Errorf("Error creating stop script: %v", err)
	}

	partyResetContent := `#!/bin/bash
for d in $(find . -type d -name 'PG_*'); do
    (cd $d && ./resetDB.sh &)
done
`
	err = os.WriteFile(resetScript, []byte(partyResetContent), 0755)
	if err != nil {
		return fmt.Errorf("Error creating reset script: %v", err)
	}

	return nil
}
