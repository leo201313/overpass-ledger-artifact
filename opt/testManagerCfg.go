package opt

import (
	"fmt"
	"opl/oplNet"
)

type ManagerCfg struct {
	NodeNetwork     oplNet.OPLNetwork `yaml:"node_network"`
	PublishNumber   int               `yaml:"publish_number"`
	QueryNumber     int               `yaml:"query_number"`
	TxBatchSize     int               `yaml:"tx_batch_size"`
	TxSendPerSecond int               `yaml:"tx_send_per_second"`
	MaxEpochDelay   int               `yaml:"max_epoch_delay"`
	QueryInterval   int               `yaml:"query_interval"`
	AccountNum      int               `yaml:"account_num"`
}

func (cfg *ManagerCfg) SetDefault() error {
	if err := cfg.NodeNetwork.Check(); err != nil {
		return err
	}

	if cfg.PublishNumber == 0 {
		cfg.PublishNumber = len(cfg.NodeNetwork.ShardNOs)
	}

	if cfg.QueryNumber == 0 {
		cfg.QueryNumber = 1
	}

	if len(cfg.NodeNetwork.ShardNOs) > cfg.PublishNumber {
		return fmt.Errorf("the client number is too small")
	}

	if len(cfg.NodeNetwork.PartyIndex)*len(cfg.NodeNetwork.ShardNOs) < cfg.PublishNumber {
		return fmt.Errorf("the client number is too big")
	}

	if cfg.AccountNum <= 0 {
		cfg.AccountNum = 1000
	}

	if cfg.TxBatchSize == 0 {
		return fmt.Errorf("the transaction batch size cannot be 0")
	}

	return nil
}

func LoadManagerConfig(filename string) (*ManagerCfg, error) {
	config := &ManagerCfg{}
	if err := loadConfig(filename, config); err != nil {
		return nil, err
	}
	if err := config.SetDefault(); err != nil {
		return nil, err
	}
	return config, nil
}
