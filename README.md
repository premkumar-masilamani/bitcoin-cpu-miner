# Bitcoin Miner
- Simple bitcoin miner written in go.
- Interacts with the Bitcoin Full Node running in the local machine.

## Pre-requisites

1. Install bitcoin full node. Please follow the [instructions](https://bitcoin.org/en/full-node).
   - If you are running the macOS, please compile from the source code. Please follow the [instructions](https://github.com/bitcoin/bitcoin/blob/master/doc/build-osx.md).
   - I prefer to run `make install` at the end, to copy the binaries to the `/usr/local/bin` folder. You can leave the binaries the `src` folder and add to the $PATH as well.

2. Create a bitcoin config file with the name `bitcoin.conf` inside the folder which has bitcoin blocks.
```
# server=1 tells Bitcoin-Qt and bitcoind to accept JSON-RPC commands
server=1

# RPC User name and password
rpcuser=bitcoinrpcuser
rpcpassword=hard-to-guess-rpc-password

# Listen for RPC connections on this TCP port:
rpcport=8332
```

3. Run bitcoin full node in daemon mode. I keep the data in a separate flash drive, and it looks like this.
```
bitcoind -disablewallet -datadir=/Volumes/Prem/Bitcoin/FullNode
```

3a. You can run the bitcoin core in daemon mode as well.
```
bitcoind -daemon -datadir=/Volumes/Prem/Bitcoin/FullNode
```

## Running Bitcoin Miner

1. Create the configuration file. You can copy the file `cmd/config/config.yaml` to `cmd/config/config.local.yaml`. Please update the relevant secrets.

| Config Name      | Default Value | Required |
|:-----------------|:--------------|:---------|
| bitcoin.host     | "localhost"   | Yes      |
| bitcoin.port     | "8332"        | Yes      |
| bitcoin.username | ""            | Yes      |
| bitcoin.password | ""            | Yes      |

2. Install the dependencies
```
go get github.com/btcsuite/btcd/rpcclient
```

3. Run the program
```
make run
```