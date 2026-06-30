# Step-by-Step Guide: Writing a Bitcoin CPU Miner

This document provides detailed step-by-step instructions to implement a Bitcoin CPU Miner in Go. It also maps each instruction to the exact package and function in the codebase.

---

## Task Checklist & Code Mapping

### Phase 1: Communication & Template Retrieval

#### 1. Connect to Bitcoin Full Node RPC Server
- **Description:** Initialize an RPC client configured to make HTTP POST requests with standard credentials and port numbers to the local `bitcoind` daemon. Verify node connectivity by querying blockchain information.
- **Code Location:** [`pkg/miner/minerservice.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/miner/minerservice.go) -> `NewMinerService`
- **Status:** [x] Completed

#### 2. Fetch Block Work (Block Template)
- **Description:** Call the `getblocktemplate` JSON-RPC method with SegWit rules enabled. The node returns a template containing the height, version, previous block hash, difficulty target (`bits`), transaction data, and coinbase subsidy.
- **Code Location:** [`pkg/miner/minerservice.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/miner/minerservice.go) -> `MinerService.MineBlock` calling `RPCClient.GetBlockTemplate`
- **Status:** [x] Completed

---

### Phase 2: Transaction & Merkle Tree Preparation

#### 3. Resolve Network Parameters & Miner address scriptPubKey
- **Description:** Query the local node to discover the active network (MainNet, TestNet, RegTest, or SimNet). Decode the miner's payout address under that network and convert it into a standard output script (`scriptPubKey`) such as P2PKH.
- **Code Location:**
  - Dynamic discovery: [`pkg/miner/minerservice.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/miner/minerservice.go) -> `MinerService.MineBlock`
  - Address script generation: [`pkg/util/convertor.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/util/convertor.go) -> `ConvertBitcoinAddressToP2PKHScript` (needs cleanup)
- **Status:** [/] In Progress / Revision Required (support dynamic parameters instead of hardcoding `MainNetParams`)

#### 4. Construct Coinbase Transaction
- **Description:** Construct the unique first transaction of the block (`coinbase`). It has:
  - Input: Spends a null hash (all `0`s) and input index `0xffffffff`.
  - Coinbase ScriptSig: Must contain the block height encoded as a little-endian byte slice prefixed with its byte length (BIP-34 consensus rule) and optional miner signature string.
  - Output: Pays the block reward subsidy + transaction fees to the miner's output script (`scriptPubKey`).
- **Code Location:** [`pkg/util/util.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/util/util.go) -> `CreateCoinbaseTx` (needs implementation)
- **Status:** [/] In Progress / Revision Required

#### 5. Deserialize Memory-Pool Transactions
- **Description:** Convert hex-serialized transactions provided in the template's `Data` field into structured transaction objects.
- **Code Location:** [`pkg/util/util.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/util/util.go) -> `MineBlock` (needs implementation)
- **Status:** [/] In Progress / Revision Required

#### 6. Build Merkle Tree & Calculate Merkle Root
- **Description:** Group the coinbase transaction and all deserialized transaction hashes into a list (with coinbase at index 0). Pair and double-SHA256 hash them iteratively in a binary tree format to produce a single 32-byte Merkle root hash.
- **Code Location:** [`pkg/util/util.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/util/util.go) -> `CalcMerkleRoot` (needs implementation)
- **Status:** [/] In Progress / Revision Required

---

### Phase 3: Block Header Assembly & Proof of Work

#### 7. Assemble Block Header
- **Description:** Concatenate the six fields of the block header into an 80-byte buffer:
  - Version (4 bytes)
  - Previous Block Hash (32 bytes)
  - Merkle Root (32 bytes)
  - Timestamp (4 bytes)
  - Bits / Difficulty (4 bytes)
  - Nonce (4 bytes)
- **Code Location:** [`pkg/util/util.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/util/util.go) -> `MineBlock` (needs implementation)
- **Status:** [/] In Progress / Revision Required

#### 8. Parse Difficulty Target
- **Description:** Convert the compact 32-bit `bits` value (e.g. `1d00ffff`) into a full 256-bit big integer. This represents the target threshold that the block hash must satisfy.
- **Code Location:** [`pkg/util/util.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/util/util.go) -> `CompactToBig` (needs implementation)
- **Status:** [/] In Progress / Revision Required

#### 9. Solve Proof of Work (Mining Loop)
- **Description:** Iterate through nonce values `0` to `0xffffffff`. For each nonce, update the header, compute the double-SHA256 block hash, and compare it as a 256-bit integer against the target. If the hash is less than or equal to the target, mining is successful.
- **Code Location:** [`pkg/util/util.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/util/util.go) -> `MineBlock` (needs implementation)
- **Status:** [/] In Progress / Revision Required

---

### Phase 4: Submission

#### 10. Submit Block to local Bitcoin Node
- **Description:** Serialize the solved block (header + transactions) and send it using the `submitblock` JSON-RPC method to the node for verification and consensus propagation.
- **Code Location:** [`pkg/miner/minerservice.go`](file:///Users/premkumar/Code/bitcoin-cpu-miner/pkg/miner/minerservice.go) -> `MinerService.MineBlock` calling `RPCClient.SubmitBlock`
- **Status:** [x] Completed
