# Bitcoin CPU Miner

A simple, educational Bitcoin CPU miner written in Go that interacts with a local Bitcoin full node.

---

## Getting Started

Follow these guides in the `docs` directory to set up, configure, and run the project:

1. **[Bitcoin Node Setup Guide](docs/bitcoin_node_setup.md)**: Instructions to install and run a local Bitcoin full node (`bitcoind`) on macOS using Homebrew, covering MainNet, TestNet, RegTest, and SimNet configurations.
2. **[Address & Key Generation Guide](docs/bitcoin_address_generation.md)**: Instructions on generating valid public addresses and WIF private keys for all network environments, including using the offline automated generator script.
3. **[Miner Pipeline Architecture](docs/bitcoin_cpu_miner_instructions.md)**: A step-by-step conceptual walkthrough of how a Bitcoin CPU miner works under the hood and how it maps to the codebase packages and functions.

---

## Configuration & Usage

### 1. Configure the Miner
Create your local configuration by copying the template config:
```bash
cp cmd/config/config.yaml cmd/config/config.local.yaml
```
Open `cmd/config/config.local.yaml` and update the host, port, RPC credentials, and mining payout address corresponding to your local Bitcoin node.

### 2. Run the Address Generator (Optional)
Generate random test addresses and WIF private keys for any environment offline:
```bash
go run cmd/keygen/main.go
```

### 3. Run Quality Verification & Unit Tests
Run the project's test suite:
```bash
make test
```

### 4. Run the CPU Miner
Start mining block templates retrieved from your local node:
```bash
make run
```
*(By default, it will look for `cmd/config/config.local.yaml` as the config path. Pass a custom path using `go run cmd/miner/main.go --config=<path>` if needed).*