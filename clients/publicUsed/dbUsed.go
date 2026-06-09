package publicUsed

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var DB_USED = opt.Options{
	BlockCacheCapacity: 32 * 8 * opt.MiB, // Set block cache to 50 MB
	WriteBuffer:        8 * 8 * opt.MiB,  // Set write buffer size to 10 MB per buffer
	//OpenFilesCacheCapacity: 500,              // Increase file handle cache capacity
	//DisableBlockCache:      false,            // Ensure block cache is enabled
	//Filter:                   filter.NewBloomFilter(10), // Enable Bloom filter for optimized lookups
	//Compression:              opt.SnappyCompression, // Enable Snappy compression
	//DisableCompactionBackoff: false,                 // Enable compaction backoff
	//CompactionL0Trigger:    4,  // Set L0 compaction trigger threshold
	//WriteL0SlowdownTrigger: 8, // Set L0 slowdown trigger threshold
	//WriteL0PauseTrigger:    12, // Set L0 pause trigger threshold
}
