## L2P

L2P is an EVM-compatible blockchain secured by a set of validators running the Parlia
consensus engine. It is a fork of [bsc](https://github.com/bnb-chain/bsc), which is itself a
fork of [go-ethereum](https://github.com/ethereum/go-ethereum). Because of that lineage you
will find that many tools, binaries and docs carry Ethereum names — most notably the client
binary itself, which is still called `geth`.

[![Build Test](https://github.com/L2Protocol/l2p/actions/workflows/build-test.yml/badge.svg)](https://github.com/L2Protocol/l2p/actions)

The chain is:

- **EVM-compatible**: existing Ethereum tooling, contracts and libraries work unchanged.
- **Validator-secured**: a limited set of validators takes turns producing blocks, with
  double-sign detection and slashing to guarantee security, stability and finality.
- **Fast**: block times step down through the fork schedule, from 3 seconds to 1.5 seconds.

Network parameters:

|                |         |
| -------------- | ------- |
| Chain ID       | 12216   |
| P2P port       | 31398   |
| HTTP-RPC port  | 8545    |
| WS-RPC port    | 8546    |

## Release types

There are three types of release, each with a clear purpose and version scheme:

- **Stable release**: production-ready builds for the vast majority of users. Format: `v<Major>.<Minor>.<Patch>`.
- **Feature release**: early access to a single feature without affecting the core product. Format: `v<Major>.<Minor>.<Patch>-feature-<FeatureName>`.
- **Preview release**: bleeding-edge builds for users who want the latest code. Format: `v<Major>.<Minor>.<Patch>-<Meta>`, where Meta indicates maturity: alpha (experimental), beta (largely complete), rc (release candidate).

## Consensus: Parlia

Proof-of-Work is a proven way to secure a decentralised network, but it is unfriendly to the
environment and needs a large number of participants to stay secure. Proof-of-Authority
defends against a 51% attack with far better efficiency, but is criticised for being less
decentralised: the validators that take turns producing blocks hold all the authority.

Parlia combines staking-based election with Proof-of-Authority:

1. Blocks are produced by a limited set of validators.
2. Validators take turns producing blocks, similar to Ethereum's Clique engine.
3. The validator set is elected in and out through staking-based governance on-chain.
4. The engine interacts with a set of system contracts to handle liveness slashing, reward
   distribution and validator set renewal.

## Building the source

For prerequisites and detailed build instructions please read the upstream
[installation instructions](https://geth.ethereum.org/docs/getting-started/installing-geth).

Building `geth` requires both Go (version 1.27 or later) and a C compiler (GCC 5 or higher).
You can install them using your favourite package manager. Once the dependencies are
installed, run

```shell
make geth
```

or, to build the full suite of utilities:

```shell
make all
```

If you get this error when running a self-built binary:

```shell
Caught SIGILL in blst_cgo_init, consult <blst>/bindinds/go/README.md.
```

add the following environment variables and build again:

```shell
export CGO_CFLAGS="-O -D__BLST_PORTABLE__"
export CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"
```

## Executables

The project comes with several wrappers/executables found in the `cmd` directory.

|    Command    | Description                                                                                                                                                                                                                                                                                            |
| :-----------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
|  **`geth`**   | Main L2P client binary. It is the entry point into the network, capable of running as a full node (default) or an archive node (retaining all historical state). It exposes JSON-RPC endpoints over HTTP, WebSocket and IPC. See `geth --help` and the [CLI page](https://geth.ethereum.org/docs/interface/command-line-options) for command line options. |
|    `clef`     | Stand-alone signing tool, which can be used as a backend signer for `geth`.                                                                                                                                                                                                                             |
|   `devp2p`    | Utilities to interact with nodes on the networking layer, without running a full blockchain.                                                                                                                                                                                                            |
|   `abigen`    | Source code generator to convert contract definitions into easy to use, compile-time type-safe Go packages. It operates on plain [contract ABIs](https://docs.soliditylang.org/en/develop/abi-spec.html) and also accepts Solidity source files.                                                        |
|  `bootnode`   | Stripped down client that only takes part in the network node discovery protocol, without running any higher level application protocols. Useful as a lightweight bootstrap node.                                                                                                                       |
|     `evm`     | Developer utility version of the EVM, capable of running bytecode snippets within a configurable environment and execution mode. Allows isolated, fine-grained debugging of EVM opcodes (e.g. `evm --code 60ff60ff --debug run`).                                                                        |
|  `extradump`  | Decodes the `extraData` field of a block header into its validator set and seal.                                                                                                                                                                                                                       |
|   `rlpdump`   | Converts binary RLP dumps to a user-friendlier hierarchical representation (e.g. `rlpdump --hex CE0183FFFFFFC4C304050583616263`).                                                                                                                                                                       |

## Running `geth`

Going through all the possible command line flags is out of scope here, but a few common
combinations will get you up to speed quickly.

### Steps to run a full node

#### 1. Download the pre-built binaries

```shell
# Linux
wget $(curl -s https://api.github.com/repos/L2Protocol/l2p/releases/latest |grep browser_ |grep geth_linux |cut -d\" -f4)
mv geth_linux geth
chmod -v u+x geth

# macOS
wget $(curl -s https://api.github.com/repos/L2Protocol/l2p/releases/latest |grep browser_ |grep geth_mac |cut -d\" -f4)
mv geth_macos geth
chmod -v u+x geth
```

#### 2. Download the config files

```shell
wget $(curl -s https://api.github.com/repos/L2Protocol/l2p/releases/latest |grep browser_ |grep mainnet |cut -d\" -f4)
unzip mainnet.zip
```

This gives you `config.toml` and `genesis.json`.

#### 3. Initialise and start the node

```shell
./geth --datadir ./node init ./genesis.json
./geth --config ./config.toml --datadir ./node --cache 8000 --rpc.allow-unprotected-txs --history.transactions 0
```

By default this runs with the path-based storage scheme and inline state pruning, keeping
the latest 90000 blocks of history state. If you want higher performance and care little
about state consistency, add `--tries-verify-mode none`.

#### 4. Monitor node status

The bundled `config.toml` writes the log to **./node/l2p.log** rather than to standard
output. When the node has started syncing you should see lines like:

```shell
t=2026-01-08T15:00:27+0000 lvl=info msg="Imported new chain segment" blocks=1 txs=177 mgas=17.317 elapsed=31.131ms number=1,234 hash=0x42e6b5…
```

#### 5. Interact with the node

Start `geth`'s built-in interactive [JavaScript console](https://geth.ethereum.org/docs/interface/javascript-console)
with the trailing `console` subcommand, or attach to an already running instance with
`geth attach`.

*Note: always use separate accounts for testing and for real funds.*

### Configuration

As an alternative to passing numerous flags to the `geth` binary, you can pass a
configuration file:

```shell
$ geth --config /path/to/your_config.toml
```

To get an idea of what the file should look like, use the `dumpconfig` subcommand to export
your existing configuration:

```shell
$ geth --your-favourite-flags dumpconfig
```

### Programmatically interfacing with nodes

As a developer you will want to interact with the network from your own programs rather than
through the console. `geth` has built-in support for JSON-RPC based APIs
([standard APIs](https://ethereum.org/en/developers/docs/apis/json-rpc/),
[`geth` specific APIs](https://geth.ethereum.org/docs/interacting-with-geth/rpc), and the
[JSON-RPC API reference](rpc/json-rpc-api.md)). These can be exposed via HTTP, WebSockets
and IPC (UNIX sockets on UNIX based platforms, named pipes on Windows).

The IPC interface is enabled by default and exposes all APIs. The HTTP and WS interfaces
must be enabled manually and only expose a subset of APIs for security reasons.

HTTP based JSON-RPC API options:

  * `--http` Enable the HTTP-RPC server
  * `--http.addr` HTTP-RPC server listening interface (default: `localhost`)
  * `--http.port` HTTP-RPC server listening port (default: `8545`)
  * `--http.api` API's offered over the HTTP-RPC interface (default: `eth,net,web3`)
  * `--http.corsdomain` Comma separated list of domains from which to accept cross-origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--ws.addr` WS-RPC server listening interface (default: `localhost`)
  * `--ws.port` WS-RPC server listening port (default: `8546`)
  * `--ws.api` API's offered over the WS-RPC interface (default: `eth,net,web3`)
  * `--ws.origins` Origins from which to accept WebSocket requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You need to speak [JSON-RPC](https://www.jsonrpc.org/specification) on all transports. You
can reuse the same connection for multiple requests.

**Note: understand the security implications of opening up an HTTP/WS based transport before
doing so. Nodes with exposed APIs are actively targeted. Browser tabs can also reach locally
running servers, so malicious web pages could try to subvert locally available APIs.**

## Running a bootnode

Bootnodes are lightweight nodes that are not behind a NAT and run only the discovery
protocol. On startup a node logs its enode, the public identifier others use to connect to it.

A bootnode needs a key, which is created with:

```shell
bootnode -genkey boot.key
```

That key is then used to run the bootnode:

```shell
bootnode -nodekey boot.key -addr :31398 -network l2p
```

The choice of port passed to `-addr` is arbitrary, but keeping it at 31398 matches the rest
of the network.

## Contribution

Thank you for considering helping out with the source code. We welcome contributions from
anyone on the internet, and are grateful for even the smallest of fixes.

If you'd like to contribute, please fork, fix, commit and send a pull request for the
maintainers to review and merge into the main code base. If you wish to submit more complex
changes though, please check up with the core developers first to ensure those changes are
in line with the general philosophy of the project and/or get some early feedback, which can
make both your efforts much lighter as well as our review and merge procedures quick and
simple.

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting)
   guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary)
   guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"

## License

The library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.

The binaries (i.e. all code inside of the `cmd` directory) are licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also
included in our repository in the `COPYING` file.
