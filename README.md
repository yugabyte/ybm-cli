# ybm-cli
A CLI implementation for YugabyteDB Aeon.

# Install with brew

`brew install yugabyte/tap/ybm`

Find more details on installation and configuration [here](https://docs.yugabyte.com/preview/yugabyte-cloud/managed-automation/managed-cli/managed-cli-overview/)

## Commands list

You can find all the commands documented [here](https://docs.yugabyte.com/preview/yugabyte-cloud/managed-automation/managed-cli/managed-cli-reference/)


## Example workflows 

You can find the example workflows documented [here](https://docs.yugabyte.com/preview/yugabyte-cloud/managed-automation/managed-cli/managed-cli-examples/)

## Running unit tests locally:
This project uses [GinkGo](https://onsi.github.io/ginkgo/) for testing.

### Pre-requisite
- **Golang** must be installed on your machine.

### Running All Tests:
To run all the tests, navigate to the `cmd` directory and use the following command:

```bash
# This is mandatory, if you run $ go test from root of your project, it will run 0 tests.
cd cmd

# Run all tests
go test
```

### Running tests in a file:
You need to first install ginkgo via:
```bash
go install github.com/onsi/ginkgo/v2/ginkgo
```

This will install ginkgo where `go` is installed e.g. `/Users/alice/go/`. You can check the go installation path by running `go env GOPATH`. It's generally at `$HOME/go`.

To run a test, go to the directory where your test resides i.e. `ybm-cli/cmd` in this case.

If your tests in a file say `cluster_test.go` are like this.
```go 
var _ = Describe("Cluster", func() {
    Describe("Pausing cluster", func() { ... }
    Describe("Resuming cluster", func() { ... }
    Describe("Get Cluster", func() { ... }
}
```

You can run the tests like:
```bash
$HOME/go/bin/ginkgo -v -focus="Cluster" # To run all tests related to Cluster.
$HOME/go/bin/ginkgo -v -focus="Pausing cluster|Resuming cluster" # To run all tests related to Pausing/Resuming cluster
```
It can be a regex as well inside focus or text from either of `Context`, `It` or `Describe` block. You can read more about it [here](https://onsi.github.io/ginkgo/#description-based-filtering).
