package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"opl/deployScripts/expManager/expUtils"
	"opl/oplNet"
	"os"
	"path/filepath"
	"strings"
)

const cfgFilePath = "./initExp.yaml"

type Config struct {
	// describe the network
	EntityIPList      []string `yaml:"entity_ip_list"`
	WorkerGroupNumber int      `yaml:"worker_group_number"`
	OPLPort           int      `yaml:"opl_port"`
	APIPort           int      `yaml:"api_port"`

	// to creat nodes
	CoordinatorClientPath string `yaml:"coordinator_client_path"`
	WorkerClientPath      string `yaml:"worker_client_path"`
	TestManagerClientPath string `yaml:"test_manager_client_path"`
	NodesCreatPath        string `yaml:"nodes_creat_path"`

	// for scp and ssh use
	UserName string `yaml:"user_name"`
	HomePath string `yaml:"home_path"`
	PemPath  string `yaml:"pem_path"`

	oplNetwork oplNet.OPLNetwork
}

func main() {
	config, err := readConfig(cfgFilePath)
	if err != nil {
		panic(fmt.Errorf("error reading config: %v", err))
	}
	fmt.Println("[INFO] Successfully loaded the initial configuration.")
	fmt.Println("[INFO] Starting to create node directories...")
	creatDir := config.NodesCreatPath
	err = os.MkdirAll(creatDir, 0755)
	if err != nil {
		panic(fmt.Errorf("error creating auto directory: %v", err))
	}

	err = setParties(creatDir, config)
	if err != nil {
		panic(fmt.Errorf("error setting coordinators: %v", err))
	}

	err = setManager(config.TestManagerClientPath, "./", config)
	if err != nil {
		panic(fmt.Errorf("error setting manager: %v", err))
	}

	fmt.Println("[INFO] Starting to create scripts...")
	creatScripts("./", config)
	fmt.Println("[INFO] Successfully initialize experiments.")

	return
}

func setParties(autoDir string, cfg *Config) error {
	for i, party := range cfg.oplNetwork.PartyIndex {
		partyDir := filepath.Join(autoDir, fmt.Sprintf("%s", party))
		fmt.Printf("[INFO] Now create node %s.\n", party)
		err := os.Mkdir(partyDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating nodes directory: %v", err)
		}

		err = setCdnt(cfg.CoordinatorClientPath, partyDir, cfg, i)
		if err != nil {
			return err
		}
		err = setWorkers(cfg.WorkerClientPath, partyDir, cfg, i)
		if err != nil {
			return err
		}
		err = createPartyScripts(partyDir, cfg)
		if err != nil {
			return err
		}

		// Compress the partyDir into a .tar.gz file
		err = expUtils.CompressDirToTarGz(partyDir, autoDir)
		if err != nil {
			return fmt.Errorf("error compressing party directory %s: %v", partyDir, err)
		}

		fmt.Printf("[INFO] Node %s is done.\n", party)
	}

	return nil
}

func setCdnt(cdntPath, partyDir string, cfg *Config, partyIndex int) error {
	nodeDir := filepath.Join(partyDir, fmt.Sprintf("coordinator_node"))
	err := os.Mkdir(nodeDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating node directory: %v", err)
	}

	coordinator := cfg.oplNetwork.CDNTNodes[partyIndex]

	dbPath := "db"

	dbDir := filepath.Join(nodeDir, dbPath)
	err = os.Mkdir(dbDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating db directory: %v", err)
	}

	coordinatorConfig := map[string]interface{}{
		"coordinator_name":            fmt.Sprintf("coordinator_%s", cfg.oplNetwork.PartyIndex[partyIndex]),
		"coordinator_address":         coordinator.OPLAddr,
		"party_name":                  cfg.oplNetwork.PartyIndex[partyIndex],
		"other_coordinator_addresses": getOtherCoordinatorAddresses(cfg.oplNetwork.CDNTNodes, partyIndex),
		"worker_addresses":            getWorkerAddresses(cfg.oplNetwork.ShardGroups, partyIndex),
		"shard_numbers":               cfg.oplNetwork.ShardNOs,
		"mem_db_mode":                 false,
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

	err = copyFile(cdntPath, filepath.Join(nodeDir, "cdntClient"))
	if err != nil {
		return fmt.Errorf("error copying client file: %v", err)
	}

	return nil
}

func setWorkers(workerPath, partyDir string, cfg *Config, partyIndex int) error {
	for i, shardNO := range cfg.oplNetwork.ShardNOs {
		nodeDir := filepath.Join(partyDir, fmt.Sprintf("worker_node_shard_%d", shardNO))
		err := os.Mkdir(nodeDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating node directory: %v", err)
		}

		worker := cfg.oplNetwork.ShardGroups[i][partyIndex]

		dbPath := "db"

		dbDir := filepath.Join(nodeDir, dbPath)
		err = os.Mkdir(dbDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating db directory: %v", err)
		}

		// Prepare the worker configuration
		workerConfig := map[string]interface{}{
			"worker_name":                    fmt.Sprintf("worker_%s_shard_%d", cfg.oplNetwork.PartyIndex[partyIndex], shardNO),
			"worker_address":                 worker.OPLAddr,
			"party_name":                     cfg.oplNetwork.PartyIndex[partyIndex],
			"shard_number":                   shardNO,
			"associated_coordinator_address": cfg.oplNetwork.CDNTNodes[partyIndex].OPLAddr,
			"other_worker_addresses":         getOtherWorkerAddresses(cfg.oplNetwork.ShardGroups[i], partyIndex),
			"mem_db_mode":                    false,
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
		err = copyFile(workerPath, filepath.Join(nodeDir, "workerClient"))
		if err != nil {
			return fmt.Errorf("error copying client file: %v", err)
		}
	}
	return nil
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

func (cfg *Config) ConstructNetwork() error {
	if len(cfg.EntityIPList) == 0 {
		return fmt.Errorf("the entity cannot be zero")
	}

	if cfg.WorkerGroupNumber <= 0 {
		return fmt.Errorf("the worker_group_number should over zero")
	}

	partyName := cfg.EntityIPList
	shardNOs := make([]uint8, cfg.WorkerGroupNumber)

	for i := 0; i < len(shardNOs); i++ {
		shardNOs[i] = uint8(i + 1)
	}

	cdnts := make([]oplNet.OPLNode, len(partyName))
	shardGroups := make([][]oplNet.OPLNode, len(shardNOs))
	for i := 0; i < len(shardGroups); i++ {
		shardGroups[i] = make([]oplNet.OPLNode, len(partyName))
	}

	for i := 0; i < len(partyName); i++ {
		ip := cfg.EntityIPList[i]
		cdnt := oplNet.OPLNode{
			NodeType: oplNet.CDNT_TYPE,
			OPLAddr:  expUtils.CatIPPort(ip, cfg.OPLPort),
			APIAddr:  expUtils.CatIPPort(ip, cfg.APIPort),
		}
		cdnts[i] = cdnt

		for j := 0; j < len(shardGroups); j++ {
			worker := oplNet.OPLNode{
				NodeType: oplNet.WORKER_TYPE,
				OPLAddr:  expUtils.CatIPPort(ip, cfg.OPLPort+j+1),
				APIAddr:  expUtils.CatIPPort(ip, cfg.APIPort+j+1),
			}
			shardGroups[j][i] = worker
		}
	}

	newOplNetWork := oplNet.OPLNetwork{
		PartyIndex:  partyName,
		CDNTNodes:   cdnts,
		ShardNOs:    shardNOs,
		ShardGroups: shardGroups,
	}
	if err := newOplNetWork.Check(); err != nil {
		return err
	}

	cfg.oplNetwork = newOplNetWork
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

	err = config.ConstructNetwork()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func createPartyScripts(partyDir string, cfg *Config) error {
	startScript := filepath.Join(partyDir, "startAll.sh")
	stopScript := filepath.Join(partyDir, "stopAll.sh")
	resetScript := filepath.Join(partyDir, "resetDB.sh")

	coordinatorStartContent := fmt.Sprintf(`#!/bin/bash
(cd coordinator_node && ./cdntClient &)
`)

	workerStartContent := fmt.Sprintf(`#!/bin/bash
for d in $(find . -type d -name 'worker_node_shard_*'); do
    (cd $d && ./workerClient &)
done
`)

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

func setManager(tmPath, autoDir string, cfg *Config) error {
	//managerDir := filepath.Join(autoDir, "manager_node")
	managerDir := autoDir

	// Check if the managerDir already exists
	if _, err := os.Stat(managerDir); err == nil {
		// If the directory exists, skip creation
		fmt.Println("[INFO] Test manager directory already exists. Skipping creation.")
	} else if os.IsNotExist(err) {
		// If the directory does not exist, create it
		err := os.Mkdir(managerDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating test manager directory: %v", err)
		}
		fmt.Println("[INFO] Test manager directory created successfully.")
	} else {
		// Handle other errors (e.g., permission issues)
		return fmt.Errorf("error checking directory: %v\n", err)
	}

	fmt.Println("[INFO] Now begin to config test manager...")

	managerConfig := map[string]interface{}{
		"node_network": map[string]interface{}{
			"party_index":  cfg.oplNetwork.PartyIndex,
			"cdnt_nodes":   cfg.oplNetwork.CDNTNodes,
			"shard_nos":    cfg.oplNetwork.ShardNOs,
			"shard_groups": cfg.oplNetwork.ShardGroups,
		},
		"publish_number":     len(cfg.oplNetwork.ShardNOs),
		"query_number":       len(cfg.oplNetwork.ShardNOs),
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

	targetPath := filepath.Join(managerDir, "tmClient")

	if _, err := os.Stat(targetPath); err == nil {
		fmt.Println("[INFO] Test manager already exists at target path. Skipping copy.")
		fmt.Println("[INFO] Test manager is done.")
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking target file: %v\n", err)
	}

	err = copyFile(tmPath, targetPath)
	if err != nil {
		return fmt.Errorf("error copying manager client file: %v", err)
	}
	fmt.Println("[INFO] Test manager is done.")
	return nil
}

func creatScripts(dir string, cfg *Config) error {
	deployScript := filepath.Join(dir, "deployAll.sh")
	startScript := filepath.Join(dir, "startAll.sh")
	stopScript := filepath.Join(dir, "stopAll.sh")
	resetScirpt := filepath.Join(dir, "resetAll.sh")

	// ----------------------------------------------------------------- deploy script
	// -------------------------------------------------------------------------------------------------
	deployContent := fmt.Sprintf(`#!/bin/bash

# Define the array of node IP addresses
nodes_ip_array=(%s)

# Define the path to the private key
key_path="%s"

# Define the target directory on the remote node
target_dir="%s"  # You can change this to any desired path

# Define the username for SSH and SCP
username="%s"

# Check if the private key file exists
if [ ! -f "${key_path}" ]; then
    echo "Error: Key file ${key_path} does not exist. Please provide the correct path."
    exit 1
fi

# Check if the array of node IPs is defined
if [ -z "${nodes_ip_array[*]}" ]; then
    echo "Error: nodes_ip_array is empty. Please define the array before running the script."
    exit 1
fi

# Loop through each IP address in the array
for ip in "${nodes_ip_array[@]}"; do
    echo "Deploying node at IP: ${ip}"
    
    # Ensure the target directory exists on the remote node
    ssh -i "${key_path}" "${username}@${ip}" "mkdir -p ${target_dir}" || {
        echo "Failed to create directory ${target_dir} on ${ip}"
        continue
    }

    # Check and remove existing tarball and extracted folder on the remote node
    tarball_path="${target_dir}/${ip}.tar.gz"
    extracted_folder="${target_dir}/${ip}"
    ssh -i "${key_path}" "${username}@${ip}" "
        if [ -f '${tarball_path}' ]; then
            echo 'Existing tarball found on ${ip}, removing...'
            rm -f '${tarball_path}'
        fi
        if [ -d '${extracted_folder}' ]; then
            echo 'Existing extracted folder found on ${ip}, removing...'
            rm -rf '${extracted_folder}'
        fi
    " || {
        echo "Failed to clean up old files on ${ip}"
        continue
    }
    
    # Copy the tarball to the remote node
    scp -i "${key_path}" "%s/${ip}.tar.gz" "${username}@${ip}:${target_dir}/" || {
        echo "Failed to copy ${ip}.tar.gz to ${ip}:${target_dir}"
        continue
    }

    # Extract the tarball on the remote node
    ssh -i "${key_path}" "${username}@${ip}" "tar -zxf ${tarball_path} -C ${target_dir}/" || {
        echo "Failed to extract ${ip}.tar.gz in ${target_dir} on ${ip}"
        continue
    }
    
    echo "Successfully deployed node at IP: ${ip}"

done

echo "Deployment completed."
`, strings.Join(cfg.EntityIPList, " "), cfg.PemPath, cfg.HomePath, cfg.UserName, cfg.NodesCreatPath)

	err := os.WriteFile(deployScript, []byte(deployContent), 0755)
	if err != nil {
		return fmt.Errorf("Error creating deploy script: %v", err)
	}

	// ----------------------------------------------------------------- start script
	// -------------------------------------------------------------------------------------------------
	startContent := fmt.Sprintf(`#!/bin/bash

# Define the array of node IP addresses
nodes_ip_array=(%s)

# Define the path to the private key
key_path="%s"

# Define the target directory on the remote node where files were extracted
target_dir="%s"  # Change this if a different directory was used during deployment

# Define the username for SSH
username="%s"

# Check if the private key file exists
if [ ! -f "${key_path}" ]; then
    echo "Error: Key file ${key_path} does not exist. Please provide the correct path."
    exit 1
fi

# Loop through each IP address in the array
for ip in "${nodes_ip_array[@]}"; do
    echo "Starting node at IP: ${ip}"
    
	# Connect to the remote node, navigate to the extracted directory, and execute startAll.sh in the background
    ssh -i "${key_path}" "${username}@${ip}" "cd ${target_dir}/${ip} && nohup ./startAll.sh > startAll.log 2>&1 &" || {
        echo "Failed to start node at ${ip}"
        continue
    }
    
    echo "Successfully started node at IP: ${ip}"
done

echo "All nodes processed."
`, strings.Join(cfg.EntityIPList, " "), cfg.PemPath, cfg.HomePath, cfg.UserName)

	err = os.WriteFile(startScript, []byte(startContent), 0755)
	if err != nil {
		return fmt.Errorf("Error creating start script: %v", err)
	}

	// ----------------------------------------------------------------- stop script
	// -------------------------------------------------------------------------------------------------
	stopContent := fmt.Sprintf(`#!/bin/bash

# Define the array of node IP addresses
nodes_ip_array=(%s)

# Define the path to the private key
key_path="%s"

# Define the target directory on the remote node where files were extracted
target_dir="%s"  # Change this if a different directory was used during deployment

# Define the username for SSH
username="%s"

# Check if the private key file exists
if [ ! -f "${key_path}" ]; then
    echo "Error: Key file ${key_path} does not exist. Please provide the correct path."
    exit 1
fi

# Loop through each IP address in the array
for ip in "${nodes_ip_array[@]}"; do
    echo "Stopping node at IP: ${ip}"
    
    # Connect to the remote node, navigate to the extracted directory, and execute stopAll.sh
    ssh -i "${key_path}" "${username}@${ip}" "cd ${target_dir}/${ip} && ./stopAll.sh" || {
        echo "Failed to stop node at ${ip}"
        continue
    }
    
    echo "Successfully stopped node at IP: ${ip}"
done

echo "All nodes processed."
`, strings.Join(cfg.EntityIPList, " "), cfg.PemPath, cfg.HomePath, cfg.UserName)

	err = os.WriteFile(stopScript, []byte(stopContent), 0755)
	if err != nil {
		return fmt.Errorf("Error creating stop script: %v", err)
	}

	// ----------------------------------------------------------------- reset script
	// -------------------------------------------------------------------------------------------------
	resetContent := fmt.Sprintf(`#!/bin/bash

# Define the array of node IP addresses
nodes_ip_array=(%s)

# Define the path to the private key
key_path="%s"

# Define the target directory on the remote node where files were extracted
target_dir="%s"  # Change this if a different directory was used during deployment

# Define the username for SSH
username="%s"

# Check if the private key file exists
if [ ! -f "${key_path}" ]; then
    echo "Error: Key file ${key_path} does not exist. Please provide the correct path."
    exit 1
fi

# Loop through each IP address in the array
for ip in "${nodes_ip_array[@]}"; do
    echo "Resetting node at IP: ${ip}"
    
    # Connect to the remote node, navigate to the extracted directory, and execute resetDB.sh
    ssh -i "${key_path}" "${username}@${ip}" "cd ${target_dir}/${ip} && ./resetDB.sh" || {
        echo "Failed to reset node at ${ip}"
        continue
    }
    
    echo "Successfully reset node at IP: ${ip}"
done

echo "All nodes processed."
`, strings.Join(cfg.EntityIPList, " "), cfg.PemPath, cfg.HomePath, cfg.UserName)

	err = os.WriteFile(resetScirpt, []byte(resetContent), 0755)
	if err != nil {
		return fmt.Errorf("Error creating reset script: %v", err)
	}

	return nil
}
