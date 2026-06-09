package opt

import "fmt"

type CdntCfg struct {
	CoordinatorName           string   `yaml:"coordinator_name"`
	CoordinatorAddress        string   `yaml:"coordinator_address"`
	PartyName                 string   `yaml:"party_name"`
	OtherCoordinatorAddresses []string `yaml:"other_coordinator_addresses"`
	WorkerAddresses           []string `yaml:"worker_addresses"`
	ShardNumbers              []uint8  `yaml:"shard_numbers"`

	MemDBMode    bool   `yaml:"mem_db_mode"`
	DatabasePath string `yaml:"database_path"`

	ContractEngine string `yaml:"contract_engine"`

	APIAddress string `yaml:"api_address"`
}

func LoadCoordinatorConfig(filename string) (*CdntCfg, error) {
	config := &CdntCfg{}
	if err := loadConfig(filename, config); err != nil {
		return nil, err
	}
	if err := config.SetDefault(); err != nil {
		return nil, err
	}
	return config, nil
}

func (cfg *CdntCfg) SetDefault() error {
	if len(cfg.CoordinatorName) == 0 {
		return fmt.Errorf("unknown coordinator with no name")
	}
	if len(cfg.CoordinatorAddress) == 0 {
		return fmt.Errorf("unknown coordinator with no listen address")
	}
	if len(cfg.PartyName) == 0 {
		return fmt.Errorf("unknown coordinator with no party")
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

func (cfg *CdntCfg) Print() {
	fmt.Println("Coordinator Configuration:")
	fmt.Printf("  Coordinator Name: %s\n", cfg.CoordinatorName)
	fmt.Printf("  Coordinator Address: %s\n", cfg.CoordinatorAddress)
	fmt.Printf("  Party Name: %s\n", cfg.PartyName)
	fmt.Printf("  Other Coordinator Addresses: %v\n", cfg.OtherCoordinatorAddresses)
	fmt.Printf("  Worker Addresses: %v\n", cfg.WorkerAddresses)
	fmt.Printf("  Shard Numbers: %v\n", cfg.ShardNumbers)
	fmt.Printf("  Memory Database Mode: %v\n", cfg.MemDBMode)
	if !cfg.MemDBMode {
		fmt.Printf("  Database Path: %s\n", cfg.DatabasePath)
	}
	fmt.Printf("  Contract Engine: %s\n", cfg.ContractEngine)
	fmt.Printf("  API Listen Address: %s\n", cfg.APIAddress)
}
