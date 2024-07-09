## ybm network-allow-list create

Create network allow lists in YugabyteDB Aeon

### Synopsis

Create network allow lists in YugabyteDB Aeon

```
ybm network-allow-list create [flags]
```

### Options

```
  -i, --ip-addr strings      [REQUIRED] IP addresses included in the Network Allow List.
  -n, --name string          [REQUIRED] The name of the Network Allow List.
  -d, --description string   [OPTIONAL] Description of the Network Allow List.
  -h, --help                 help for create
```

### Options inherited from parent commands

```
  -a, --apiKey string      YBM Api Key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
      --host string        YBM Api hostname
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm network-allow-list](ybm_network-allow-list.md)	 - Manage Network Allow Lists

