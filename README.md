# ybm-cli
A CLI implementation for YugabyteDB Aeon.

# Install with brew

`brew install yugabyte/tap/ybm`

For details on installation and configuration, see [YugabyteDB Aeon CLI](https://docs.yugabyte.com/stable/yugabyte-cloud/managed-automation/managed-cli/) in the YugabyteDB Aeon documentation.

## Commands list

The command reference is in the [docs](https://github.com/yugabyte/ybm-cli/tree/main/docs/ybm.md) directory of this repository.

## Example workflows 

- [Create cluster](https://docs.yugabyte.com/stable/yugabyte-cloud/managed-automation/managed-cli/managed-cli-example-create/)
- [Other short examples](https://docs.yugabyte.com/stable/yugabyte-cloud/managed-automation/managed-cli/managed-cli-example-quick/)

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

## New Command Checklist for YBM CLI

This document serves as a checklist to guide the process of adding a new command/subcommand to the YBM CLI tool. 
Follow each step to ensure the command is implemented, tested, and documented consistently with YBM CLI standards.

### 1. Command Implementation

The following for each command should work(of course including the intended functionality):
- [ ] **Output Format Support**:
  - [ ] Ensure support for output formats: **table**, **JSON**, and **pretty**. Ideally, this will be taken care automatically when you implement your formatter using `internal/formatter` package.
    - **Table**: Display data in a structured table format, suitable for terminal viewing. Refer commands like `$ ybm cluster list`. THis is default output format.
    - **JSON**: Provide output in raw JSON format for easy parsing.
    - **Pretty JSON**: Format JSON output in a human-readable way, with indentation.
  - [ ] Test output in each format to verify correct display and structure.
- [ ] **Async Operation Support**:
  - [ ] For commands with asynchronous operations (e.g. `cluster pause` or `resume` etc), implement support for the `--wait` flag.
    - When `--wait` is specified, the command should wait until the operation completes and display the final status. This flag is globally added to all commands, to implement it you need to call [this](https://github.com/yugabyte/ybm-cli/blob/c24aca2700307de5d8f91735e9e75659f1c25847/internal/client/client.go#L1327) i.e. `authApi.WaitForTaskCompletion(...)`. Refer [this](https://github.com/yugabyte/ybm-cli/blob/c24aca2700307de5d8f91735e9e75659f1c25847/cmd/cluster/log-exporter/query_log_exporter.go#L251) for example.

### 2. Testing

- **Unit Tests**: Write unit tests for the command including but not limited to:
  - [ ] A test case for required params not provided. 
  - [ ] A test to assert Table output format. 
  - [ ] A test to assert Json output format. 
  - [ ] When mocking an API call, verify that the call was actually made. For example, use `Expect(server.ReceivedRequests()).Should(HaveLen(1))`. For `POST`, `PUT`, or `PATCH` requests, you can also assert the request body content as follows:
    ```shell
    server.AppendHandlers(
      ghttp.CombineHandlers(
        ghttp.VerifyRequest(http.MethodPut, "/some/path"),
        # Here we are asserting on the expected Request Body
        ghttp.VerifyJSON(`{"expectedField1":"val1",...}`),
        ghttp.RespondWithJSONEncodedPtr(&statusCode, resp),
      )
    )
    ```
    This approach is useful for confirming that all necessary fields were included when building the request model, helping catch any omissions in the request structure.

- **Run Tests**: Execute the test suite and verify all tests pass.

### 3. Documentation
To generate doc, run:
```shell
cd ~/code/ybm-cli/
make doc
```
It will automatically generate document for each subcommand.
