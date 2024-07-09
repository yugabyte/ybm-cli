## ybm network-allow-list list

List network allow lists in YugabyteDB Managed

### Synopsis

List network allow lists in YugabyteDB Managed

```
ybm network-allow-list list [flags]
```

### Options

```
  -h, --help          help for list
  -n, --name string   [OPTIONAL] The name of the Network Allow List.
```

### Options inherited from parent commands

```
  -a, --apiKey string      YBM Api Key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm network-allow-list](ybm_network-allow-list.md)	 - Manage Network Allow Lists

