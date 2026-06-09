package oplNet

import "fmt"

type OPL_NODE_TYPE uint8

const CDNT_TYPE OPL_NODE_TYPE = 0
const WORKER_TYPE OPL_NODE_TYPE = 1

type OPLNode struct {
	NodeType OPL_NODE_TYPE `yaml:"node_type"`
	OPLAddr  string        `yaml:"opl_addr"`
	APIAddr  string        `yaml:"api_addr"`
}

type OPLNetwork struct {
	PartyIndex  []string    `yaml:"party_index"`
	CDNTNodes   []OPLNode   `yaml:"cdnt_nodes"`
	ShardNOs    []uint8     `yaml:"shard_nos"`
	ShardGroups [][]OPLNode `yaml:"shard_groups"`
}

func (on *OPLNetwork) Check() error {
	if len(on.PartyIndex) == 0 {
		return fmt.Errorf("the party is empty")
	}
	if len(on.PartyIndex) != len(on.CDNTNodes) {
		return fmt.Errorf("unmatched paties and coordinators")
	}

	for i, node := range on.CDNTNodes {
		if node.NodeType != CDNT_TYPE {
			return fmt.Errorf("CDNTNode at index %d is not of type CDNT_TYPE", i)
		}
	}

	if len(on.ShardNOs) == 0 {
		return fmt.Errorf("the shard numbers is empty")
	}
	if len(on.ShardNOs) != len(on.ShardGroups) {
		return fmt.Errorf("unmatched shard numbers and shard groups")
	}

	for i := 0; i < len(on.ShardNOs); i++ {
		if len(on.ShardGroups[i]) != len(on.PartyIndex) {
			return fmt.Errorf("unmatched parties and workers in shard group %d", i)
		}
	}

	for i, group := range on.ShardGroups {
		for j, node := range group {
			if node.NodeType != WORKER_TYPE {
				return fmt.Errorf("node in ShardGroups[%d][%d] is not of type WORKER_TYPE", i, j)
			}
		}
	}

	return nil
}

func (on *OPLNetwork) String() string {
	var result string

	// Print Coordinators (CDNTNodes) with PartyIndex
	result += "Coordinators:\n"
	for i, node := range on.CDNTNodes {
		result += fmt.Sprintf("  Party [%s]: {NodeType: %d, OPLAddr: %s, APIAddr: %s}\n",
			on.PartyIndex[i], node.NodeType, node.OPLAddr, node.APIAddr)
	}

	// Print ShardGroups with PartyIndex
	result += "\nShardGroups:\n"
	for shardIndex, group := range on.ShardGroups {
		result += fmt.Sprintf("  Shard %d:\n", on.ShardNOs[shardIndex])
		for partyIndex, node := range group {
			result += fmt.Sprintf("    Party [%s]: {NodeType: %d, OPLAddr: %s, APIAddr: %s}\n",
				on.PartyIndex[partyIndex], node.NodeType, node.OPLAddr, node.APIAddr)
		}
	}

	return result
}
