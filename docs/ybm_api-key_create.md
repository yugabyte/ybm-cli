## ybm api-key create

Create an API Key

### Synopsis

Create an API Key

```
ybm api-key create [flags]
```

### Options

```
      --name string                  [REQUIRED] The name of the API Key.
      --duration int32               [REQUIRED] The duration for which the API Key will be valid. 0 denotes that the key will never expire.
      --unit string                  [REQUIRED] The time units for which the API Key will be valid. Available options are Hours, Days, and Months.
      --description string           [OPTIONAL] Description of the API Key to be created.
      --network-allow-lists string   [OPTIONAL] The network allow lists(comma separated names) to assign to the API key.
      --role-name string             [OPTIONAL] The name of the role to be assigned to the API Key. If not provided, an Admin API Key will be generated.
  -f, --force                        Bypass the prompt for non-interactive usage
  -h, --help                         help for create
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm api-key](ybm_api-key.md)	 - Manage API Keys

