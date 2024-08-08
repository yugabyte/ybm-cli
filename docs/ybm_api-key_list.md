## ybm api-key list

List API Keys

### Synopsis

List API Keys in your YugabyteDB Aeon account

```
ybm api-key list [flags]
```

### Options

```
      --name string     [OPTIONAL] To filter by API Key name.
      --status string   [OPTIONAL] To filter by API Key status. Available options are ACTIVE, EXPIRED, REVOKED.
  -h, --help            help for list
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
      --host string        YugabyteDB Aeon Api hostname
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm api-key](ybm_api-key.md)	 - Manage API Keys

