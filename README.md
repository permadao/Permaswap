# Permaswap

[whitepaper](https://mirror.xyz/permaswap.eth/kdg0iXx1jB-vXYEc_WEAeTNX_sGjv8BXksHxcFdoKjo)

## Setup

To install the Go package:

```
go mod tidy
```

## Run

You can operate as either a **Router** or **LP**. Alternatively, you might choose to run a **HALO Node**, which only retrieves halo transactions and computes the current state.

The EVM private key refers to an Ethereum account private key. When operating as a Router, it's utilized to sign router transactions that are submitted to Everpay or Arweave. In contrast, when running a HALO Node, it's not required; simply use a test account private key.

#### Router

1. Stake a minimum of 80,000 HALO test tokens.
2. Prepare the router configuration file:
```
cd cmd/router
```

Edit the example.toml configuration file to suit your needs.

3. compile router

```
go build
```

4. start a router
```
./router --config example.toml
```

#### LP

For documentation, refer to this guide [this guide](https://permadao.notion.site/Golang-LP-client-configuration-tutorial-0c8b65f06eed4add880dad0f29d89d37)


#### HALO Node
```
cd cmd/halo
cp run_example.sh run.sh
```

Update the run.sh script with your MySQL DSN, EVM private key, and Genesis transaction.

```
source run.sh
```

#### GENESIS Tx

Genesis tx [0x91be83007f1b642d328ab01a7759f38b75f89a61079a998c0fce834fc36f7b91](https://scan.everpay.io/tx/0x91be83007f1b642d328ab01a7759f38b75f89a61079a998c0fce834fc36f7b91)