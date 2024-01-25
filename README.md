# Permaswap

[whitepaper](https://mirror.xyz/permaswap.eth/kdg0iXx1jB-vXYEc_WEAeTNX_sGjv8BXksHxcFdoKjo)

## Setup

install go package

```
go mod tidy
```

## Run

Run as **Router** or **LP**.

You also could just run **HALO Node** , which fetch halo txs and calculate current state.

### Router

```
cd cmd/router
cp run_example.sh run.sh
```

Fill in your MySQL DSN / ECC private key/ Genesis tx in run.sh

```
source run.sh
```

### LP

[doc](https://permadao.notion.site/Golang-LP-client-configuration-tutorial-0c8b65f06eed4add880dad0f29d89d37)


### HALO Node
```
cd cmd/halo
cp run_example.sh run.sh
```

Fill in your MySQL DSN / ECC private key/ Genesis tx in run.sh

```
source run.sh
```
