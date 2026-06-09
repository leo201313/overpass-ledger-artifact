# Overpass Ledger

Overpass Ledger is a Go implementation of the Overpass Ledger prototype used in
our research artifact. The repository contains the core ledger components,
networking layer, state managers, smart-contract execution logic, client
programs, and scripts for local or multi-machine experiments.

## Repository Layout

- `clients/`: command-line entry points for coordinators, workers, and the test
  manager.
- `common/`, `elements/`, `coes/`, `smartcontract/`, `stateManager/`: core data
  structures and execution/state-management logic.
- `entity/`: coordinator, worker, test manager, transaction pool, and supporting
  runtime entities.
- `network/` and `oplNet/`: peer-to-peer networking and network topology
  configuration.
- `database/`: storage abstraction and LevelDB-backed implementation.
- `deployScripts/`: scripts and configuration templates for local, general, and
  multi-machine experiments.
- `rlp/`: RLP encoding/decoding utilities.

## Requirements

- Go 1.22.2 or later in the Go 1.22 series.
- LevelDB dependencies are fetched through Go modules.
- For multi-machine experiments, Linux hosts with SSH/SCP access are expected by
  the deployment scripts.

## Build

Build the three main clients:

```bash
make
```

Or build a specific client:

```bash
make build-coordinator
make build-testmanager
make build-worker
```

Generated binaries are intentionally ignored by Git.

## Test

Compile all packages without running test bodies:

```bash
go test -run '^$' ./...
```

Run the full test suite with:

```bash
go test ./...
```

Some networking and entity tests open fixed local TCP ports. If your
environment blocks local listening sockets, or if those ports are already in
use, those tests may fail even though the packages build.

## Local Test Quick Start

The local test runs all OPL coordinator and worker processes on `127.0.0.1`
using the topology in `deployScripts/localtest/localtest.yaml`. By default, it
creates four parties and two worker shards.

Generate the local test directory:

```bash
make localtest
```

This creates `localtest/auto/`, including one directory per party and a
`manager_node` directory. Start all local coordinator and worker processes:

```bash
cd localtest/auto
./startAll.sh
```

Open another terminal and enter the manager directory:

```bash
cd localtest/auto/manager_node
```

Print the generated network configuration:

```bash
./tmClient printNetwork
```

Run a SmallBank workload:

```bash
./tmClient startTest
```

The test manager increases the offered load, waits for stable throughput, and
then prints metrics such as max/min/average TPS, latency, total transactions,
inherited transactions, re-executed transactions, and contention rate.

Other useful manager commands include:

```bash
./tmClient activeInfo
./tmClient startTransfer
./tmClient startSort -size 100000
./tmClient details -addr <node_api_address>
./tmClient requireState -addr <node_api_address>
```

Stop or reset the local deployment from `localtest/auto/`:

```bash
./stopAll.sh
./resetAll.sh
```

The generated `localtest/` directory is a build/runtime artifact and is ignored
by Git.

## Other Experiments

Create a general test directory:

```bash
make generaltest
```

Create a multi-machine experiment artifact:

```bash
make expbuild
```

The generated `generaltest/` and `expOPL/` directories are build outputs and
are ignored by Git. See `deployScripts/README.md` and
`deployScripts/expManager/README.md` for deployment details.

## Clean

Remove generated experiment folders and binaries:

```bash
make clean
```

## Citation

If you use this artifact in academic work, please cite the corresponding
Overpass Ledger paper.
