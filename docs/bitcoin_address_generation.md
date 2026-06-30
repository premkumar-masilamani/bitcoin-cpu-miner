# Bitcoin Address Generation Guide

This guide details how to generate valid Bitcoin addresses and private keys for testing the CPU Miner app across all environment configurations.

---

## Method 1: Automated Offline Keygen Script (Recommended)

We have provided a fully automated offline script in the codebase that generates private keys and derives valid addresses for all 4 network environments without requiring a running node.

### Running the Generator

Execute the script from the project root:
```bash
go run cmd/keygen/main.go
```

### Example Output
```
================================================================================
             Generated Offline Bitcoin Keypairs & Addresses                     
================================================================================
Environment:   MainNet
WIF Private:   Kz8oq4qPFwrRHmnEchihzNrRU4ps8kGtP7aL7mjueMW4NAyKVaUu
P2PKH Address: 19yKWub6YNU5VWx3Bq8Pccbs9Z3hbCGHkq
--------------------------------------------------------------------------------
Environment:   TestNet3
WIF Private:   cQVoHyqEh1YgTDFW17XqMhMV6J8GoCNaT9ioECCR9UA4cv5asidz
P2PKH Address: mpVGoxg5MPuLGdReuQ6mSXpC1YeQQmiYFK
--------------------------------------------------------------------------------
Environment:   RegTest
WIF Private:   cQVoHyqEh1YgTDFW17XqMhMV6J8GoCNaT9ioECCR9UA4cv5asidz
P2PKH Address: mpVGoxg5MPuLGdReuQ6mSXpC1YeQQmiYFK
--------------------------------------------------------------------------------
Environment:   SimNet
WIF Private:   Fqu9pYkCMeCa1HQj8afhhWCnVuCbm2gXpk8gT43FtxQ8UuPtFZff
P2PKH Address: SWGKYkNFGjfH1pjVjG7UAWkRoLH8KApc7q
--------------------------------------------------------------------------------
```

### How the Script Works
The code (`cmd/keygen/main.go`) uses the `btcec` elliptic curve package to generate a secp256k1 private key, encodes it in Wallet Import Format (WIF), and uses `btcutil` to derive a compressed legacy P2PKH public key address under the network parameters for MainNet, TestNet3, RegTest, and SimNet.

---

## Method 2: Node Address Generation (`bitcoin-cli`)

If you have a running local Bitcoin node (`bitcoind`), you can generate wallet-backed addresses directly via RPC.

### 1. Create a Wallet (Required on modern nodes)
Bitcoin Core disables default wallet creation. You must create one first:
```bash
# RegTest:
bitcoin-cli -regtest createwallet "testwallet"

# TestNet:
bitcoin-cli -testnet createwallet "testwallet"
```

### 2. Generate a Legacy Address (P2PKH)
To generate a legacy address matching the format supported by the miner:
```bash
# RegTest:
bitcoin-cli -regtest getnewaddress "minerlabel" legacy

# TestNet:
bitcoin-cli -testnet getnewaddress "minerlabel" legacy

# MainNet:
bitcoin-cli getnewaddress "minerlabel" legacy
```

### 3. Generate a Bech32 SegWit Address
Modern wallets and nodes default to Native SegWit (Bech32) addresses:
```bash
# RegTest:
bitcoin-cli -regtest getnewaddress "minerlabel" bech32
```

---

## Address Prefix Reference

To confirm that your generated address matches the selected environment, inspect the leading character:

| Environment | Address Type | Starting Prefix | Example |
| :--- | :--- | :--- | :--- |
| **MainNet** | Legacy (P2PKH) | `1` | `19yKWub6YNU5VWx3Bq8Pccbs...` |
| **MainNet** | SegWit (Bech32) | `bc1` | `bc1qxyz...` |
| **TestNet3 / RegTest** | Legacy (P2PKH) | `m` or `n` | `mpVGoxg5MPuLGdReuQ6mS...` |
| **TestNet3 / RegTest** | SegWit (Bech32) | `tb1` | `tb1qxyz...` |
| **SimNet** | Legacy (P2PKH) | `S` | `SWGKYkNFGjfH1pjVjG7UA...` |
