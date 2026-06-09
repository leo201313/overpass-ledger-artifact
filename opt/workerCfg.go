package opt

import "fmt"

type WorkerCfg struct {
	WorkerName                   string   `yaml:"worker_name"`
	WorkerAddress                string   `yaml:"worker_address"`
	PartyName                    string   `yaml:"party_name"`
	ShardNumber                  uint8    `yaml:"shard_number"`
	AssociatedCoordinatorAddress string   `yaml:"associated_coordinator_address"`
	OtherWorkerAddresses         []string `yaml:"other_worker_addresses"`

	MemDBMode    bool   `yaml:"mem_db_mode"`
	DatabasePath string `yaml:"database_path"`

	ContractEngine string `yaml:"contract_engine"`

	APIAddress string `yaml:"api_address"`
}

func LoadWorkerConfig(filename string) (*WorkerCfg, error) {
	config := &WorkerCfg{}
	if err := loadConfig(filename, config); err != nil {
		return nil, err
	}
	if err := config.SetDefault(); err != nil {
		return nil, err
	}
	return config, nil
}

func (cfg *WorkerCfg) SetDefault() error {
	if len(cfg.WorkerName) == 0 {
		return fmt.Errorf("unknown worker with no name")
	}
	if len(cfg.WorkerAddress) == 0 {
		return fmt.Errorf("unknown worker with no listen address")
	}
	if len(cfg.PartyName) == 0 {
		return fmt.Errorf("unknown worker with no party")
	}

	if !cfg.MemDBMode && len(cfg.DatabasePath) == 0 {
		return fmt.Errorf("memory database mode is off but no database path is set")
	}

	if len(cfg.ContractEngine) == 0 {
		cfg.ContractEngine = DEMO_ENGINE
	}
	if len(cfg.APIAddress) == 0 {
		cfg.APIAddress = DEFAULT_API_ADDR
	}

	return nil
}

func (cfg *WorkerCfg) Print() {
	fmt.Println("Worker Configuration:")
	fmt.Printf("  Worker Name: %s\n", cfg.WorkerName)
	fmt.Printf("  Worker Address: %s\n", cfg.WorkerAddress)
	fmt.Printf("  Party Name: %s\n", cfg.PartyName)
	fmt.Printf("  Shard Number: %d\n", cfg.ShardNumber)
	fmt.Printf("  Associated Coordinator Address: %s\n", cfg.AssociatedCoordinatorAddress)
	fmt.Printf("  Other Worker Addresses: %v\n", cfg.OtherWorkerAddresses)
	fmt.Printf("  Memory Database Mode: %v\n", cfg.MemDBMode)
	if !cfg.MemDBMode {
		fmt.Printf("  Database Path: %s\n", cfg.DatabasePath)
	}
	fmt.Printf("  Contract Engine: %s\n", cfg.ContractEngine)
	fmt.Printf("  API Listen Address: %s\n", cfg.APIAddress)
}
