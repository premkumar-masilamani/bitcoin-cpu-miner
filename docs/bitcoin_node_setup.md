# Bitcoin Node Setup on macOS

This guide provides step-by-step instructions to install and run a local Bitcoin full node on macOS using Homebrew.

## Installation

We recommend using Homebrew to manage your Bitcoin installation.

### Option 1: Command Line Daemon (`bitcoind` & `bitcoin-cli`)
If you only need the headless daemon and command-line control tools (ideal for development and running background tasks):
```bash
brew install bitcoin
```

### Option 2: Graphical User Interface (`Bitcoin-Qt`)
If you prefer a visual wallet and node dashboard:
```bash
brew install --cask bitcoin-core
```
This installs the macOS application to your Applications directory, which includes `bitcoin-qt` as well as the command-line tools.

---

## Network Environments Explained

Bitcoin Core supports multiple separate network environments, allowing developers to test their applications safely without using real money.

| Network | Description | Use Case |
| :--- | :--- | :--- |
| **MainNet** | The real, live production Bitcoin network. Blocks require significant energy to mine (ASICs). | Production use. Transactions cost actual money. |
| **TestNet (TestNet3)** | A global public test network where coins have no value. Hashing difficulty is low but requires syncing block headers from peer nodes. | Testing network protocols and multi-party workflows on a public blockchain. |
| **RegTest (Regression Test)** | A completely private, local network. Blocks can be instantly mined on-demand with zero computational effort. | Local testing, unit testing, and fast-paced application development. **(Recommended for CPU Miner development)** |
| **SimNet (Simulation Network)** | Similar to RegTest, but runs with slightly different default configurations (e.g., address prefixes, genesis block) optimized for simulating network environments. | Simulated multi-node networks and consensus rule testing. |

---

## Configurations (`bitcoin.conf`)

By default, Bitcoin Core looks for configuration files at `~/Library/Application Support/Bitcoin/bitcoin.conf`. Below are templates for the different network configurations.

### Configuration Template for RegTest / SimNet (Development Mode)
Create or edit your `bitcoin.conf`:
```ini
# server=1 tells Bitcoin Core to accept JSON-RPC commands
server=1

# Run the node in the background as a daemon
daemon=1

# Enable RPC authentication
rpcuser=bitcoinrpcuser
rpcpassword=hard-to-guess-rpc-password

# Network selection (choose one of these)
regtest=1
# simnet=1

# Configure listen options
[regtest]
rpcport=18443
rpcbind=127.0.0.1
rpcallowip=127.0.0.1
```

### Configuration Template for TestNet (Public Test)
```ini
server=1
daemon=1
rpcuser=bitcoinrpcuser
rpcpassword=hard-to-guess-rpc-password

testnet=1

[test]
rpcport=18332
rpcbind=127.0.0.1
rpcallowip=127.0.0.1
```

### Configuration Template for MainNet (Production Node)
```ini
server=1
daemon=1
rpcuser=bitcoinrpcuser
rpcpassword=hard-to-guess-rpc-password

[main]
rpcport=8332
rpcbind=127.0.0.1
rpcallowip=127.0.0.1
```

### Using a Custom `bitcoin.conf` Path
By default, Bitcoin Core looks in standard directory paths. However, you can store your config file anywhere (for example, in your project directory) and pass it explicitly to both `bitcoind` and `bitcoin-cli` using the `-conf` option:

```bash
# Start bitcoind with a custom config file path:
bitcoind -conf=/absolute/path/to/your/bitcoin.conf

# Interact with the node using bitcoin-cli with the same custom config:
bitcoin-cli -conf=/absolute/path/to/your/bitcoin.conf getblockchaininfo
```

> [!IMPORTANT]
> If you start `bitcoind` with a custom config path, you **must** also pass the exact same `-conf` parameter to `bitcoin-cli` so it knows which RPC credentials, host, and port to use.

---

## Running the Bitcoin Node

Depending on how you installed Bitcoin, run the commands below:

### 1. Running RegTest Node
To start the daemon on RegTest:
```bash
bitcoind -regtest -daemon
```

To verify the daemon is running and check its status:
```bash
bitcoin-cli -regtest getblockchaininfo
```

To stop the node:
```bash
bitcoin-cli -regtest stop
```

### 2. Generating Blocks Instantly (RegTest Exclusive)
Because RegTest does not have a real network hashing block generation loop, you must generate blocks on-demand to test your mining rewards or transactions:
```bash
# Generate 101 blocks to an address to mature coinbase rewards:
bitcoin-cli -regtest generatetoaddress 101 <your_bitcoin_address>
```

### 3. Running SimNet Node
To start:
```bash
bitcoind -simnet -daemon
```

Check status:
```bash
bitcoin-cli -simnet getblockchaininfo
```

### 4. Running TestNet Node
Start TestNet to sync with the public testnet network:
```bash
bitcoind -testnet -daemon
```
Note: Initial synchronization may take hours depending on network speeds.

### 5. Running MainNet Node
Start syncing the main Bitcoin blockchain:
```bash
bitcoind -daemon
```
> [!CAUTION]
> Syncing the mainnet blockchain requires hundreds of gigabytes of disk space and a long download duration.
