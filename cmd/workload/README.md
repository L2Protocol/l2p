## Workload Testing Tool

This tool performs RPC calls against a live node. Note the tests require a fully synced
node.

The test cases are read from JSON files, which you generate from a synced node (see
below). Point the tool at them with `--queries`, `--history-tests` and `--trace-tests`:

```shell
> ./workload test --queries queries/filter_queries.json --history-tests queries/history.json --trace-tests queries/trace.json http://host:8545
```

To run a specific test, use the `--run` flag to filter the test cases. Filtering works
similar to the `go test` command. For example, to run only tests for `eth_getBlockByHash`
and `eth_getBlockByNumber`, use this command:

```
> ./workload test --history-tests queries/history.json --run History/getBlockBy http://host:8545
```

### Generating tests

There is a facility for generating the tests from the chain. Run the following commands
(in this directory) against a synced node:

```shell
> go run . filtergen --queries queries/filter_queries.json http://host:8545
> go run . historygen --history-tests queries/history.json http://host:8545
> go run . tracegen --trace-tests queries/trace.json http://host:8545
```
