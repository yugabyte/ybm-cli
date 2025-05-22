# About this project
This is a cli tool that enables you to interact with YugabyteDB Aeon accounts using commands in your command-line shell. It's like aws cli but for YugabyteDB Managed offering cloud.
With minimal configuration, you can start running commands from the command line that implement functionality equivalent to that provided by the browser-based YugabyteDB Aeon interface.

# Tech Stack
## Core
This project uses Go/Golang. The golang libs used are defined in @go.mod file.
For cli, spf13/cobra lib is used. The HTTP calls are made using the open api generated
go client of YBM backend.

## Tests:
For testing we use Ginkgo and Gomega. All mock Http responses are defined in directly @cmd/tests/fixture/<api>.json.
Try to reuse them wherever applicable, for e.g. accounts.json or projects.json

## Project Structure:
- ~/<project-root>/internal/ has internal code like the client, formatter etc.
- ~/<Project-root>/cmd/ has code for all cli commands exposed in the CLI. Each command has it's own
  directory, generally same as package name. E.g. cmd/cluster/cluster.go
- ~/<Project-root>/docs is the docs directory, auto generated via cobra in markdown format.

## Working:
Each time you run a cli command e.g. $ ybm cluster pause ..., cobra calls the relevant function.
The implementation then does the field validation etc, calls the YBM api via the open api generated client
and outputs the revelant stuff for the commands based on HTTP response it received. It can call multiple api's
based on the command.

### Command Output:
Each command can emit 3 types of output: JSON raw, Table or pretty json.
User can specify that via: -o or --output flag e.g. -o json. Default output format is table.

### Adding a new command:
Generally when adding a new command, we need to do 3 things:
1. Create a new directory under cmd/ for the new command.
2. Define a root command for it via cobra. Add appropriate subcommands and implement their logic.
3. Decide the columns to show in output table and define a formatter for the new command output in internal/formatter/ directory.

The info about a command can be fetched from API docs that mentions about the functionality of an
API. E.g. if we want to provide functionality of enable database query logging on a cluster. We can refer
these API docs:
1. https://api-docs.yugabyte.com/docs/managed-apis/badf5234abfd9-create-a-database-query-logging-configuration-and-associate-with-cluster
2. https://api-docs.yugabyte.com/docs/managed-apis/b6fccb276393f-list-database-query-logging-configurations-associated-with-a-cluster
3. https://api-docs.yugabyte.com/docs/managed-apis/ed35b3ae3711f-update-database-query-logging-configuration-associated-with-a-cluster
4. https://api-docs.yugabyte.com/docs/managed-apis/be0da833721b6-remove-database-query-logging-configuration-associated-with-a-cluster

And our cli command may look like below with given 4 subcommands:
```cli
> ybm cluster db-query-logging --help
Configure Database Query Logging for your Cluster.

Usage:
  ybm cluster db-query-logging [flags]
  ybm cluster db-query-logging [command]

Available Commands:
  describe    Describe Database Query Logging config
  disable     Disable Database Query Logging
  enable      Enable Database Query Logging
  update      Update Database Query Logging config
```

The resource here is `cluster` and it's subcommand is `db-query-logging`. The verbs describe, disable etc describes what action needs to be performed.

## API Client Interaction

Commands interact with the YugabyteDB Aeon API via a generated Go client located in `internal/client`. Here's the typical pattern:

1.  **Initialization:**
    *   Inside the command's `Run` or `RunE` function, initialize the client using `authApi, err := ybmAuthClient.NewAuthApiClient()`.
    *   Immediately check for errors during initialization and log fatally using `logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))`. This ensures consistent error reporting.
    *   The `NewAuthApiClient` function handles loading the relevant api keys and backend url automatically.

2.  **Authentication Check (Standard Practice):**
    *   After successful initialization, call `authApi.GetInfo("", "")`. This step initializes & fetches account information and mandatory to run.

3.  **Making API Calls:**
    *   Use the methods on the initialized `authApi` object to perform API operations.
    *   @client.go file is a wrapper on openApi generated go client. Any new APi must be added in client.go file
        first and then used.

4.  **Handling API Errors:**
    *   Check the error returned by `.Execute()` immediately.
    *   If an error is present:
        *   Optionally log the full HTTP response for debugging: `logrus.Debugf("Full HTTP response: %v", r)`
        *   Use the `ybmAuthClient.GetApiErrorDetails(err)` helper function to extract user-friendly error details from the API response.
        *   Log the processed error details and exit fatally: `logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))`

5.  **Example (`create integration`):**
    ```go
    // Inside Run function:
    authApi, err := ybmAuthClient.NewAuthApiClient()
    if err != nil {
        logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
    }
    authApi.GetInfo("", "") // Standard check

    // prepare relevant spec for HTTP call

    resp, r, err := authApi.relevantApiCall..Execute()

    if err != nil {
        logrus.Debugf("Full HTTP response: %v", r) // Optional debug log
        logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err)) // Process and log error
    }

    // ... process successful response (resp) ...
    ```
6. **Output format**:
  For each command the, define a formatter in directory internal/formatter/<cmd>.go, refer implementation
  of any one existing formatter to understand more.
