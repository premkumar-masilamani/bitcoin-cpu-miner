# Project Rules & Learnings - Bitcoin CPU Miner

Guidelines and learnings captured during the CPU miner rewrite and test development.

## Codebase Architecture & Libraries
- **Standard Libraries**: Always use the standard `btcsuite` packages (`github.com/btcsuite/btcd/wire`, `github.com/btcsuite/btcutil`, `github.com/btcsuite/btcd/txscript`) instead of manual hex serialization hacks.
- **Dynamic Chains**: Always query the active network environment (`chain`) dynamically from the node via `GetBlockChainInfo` to automatically configure `*chaincfg.Params`. Never hardcode `MainNetParams` for local mining testing.
- **BIP-34 Coinbase Rules**: Coinbase transactions must serialize the block height as the first item in the input scriptSig using length-prefixed little-endian bytes.

## Testing Conventions
- **Test File Naming**: Every test file must strictly match the name of the functional code file with a `_test` suffix (e.g., `pkg/miner/minerservice.go` ➡️ `pkg/miner/minerservice_test.go`).
- **RPC Client Mocking**: Do not mock `rpcclient.Client` directly (concrete structure). Instead, define a minimal `BitcoinRPCClient` interface containing needed methods and inject it.
- **PoW Timeout Sim**: For mining loop timeout unit tests, define `MiningDurationInSeconds` as a package-level variable instead of a constant so it can be overridden to a negative value during tests.

## Node Commands & Workflows
- **Custom Config Path**: When running `bitcoind` with a custom config file path via `-conf=/path/to/bitcoin.conf`, always pass the same `-conf` parameter to `bitcoin-cli`.
- **Local Key Generation**: Use `go run cmd/keygen/main.go` to generate random keypairs and P2PKH addresses offline for MainNet, TestNet, RegTest, and SimNet testing.
