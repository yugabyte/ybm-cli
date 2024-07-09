## ybm integration

Manage Integration

### Synopsis

Manage Integration

```
ybm integration [flags]
```

### Options

```
  -h, --help   help for integration
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

* [ybm](ybm.md)	 - ybm - Effortlessly manage your DB infrastructure on YugabyteDB Aeon (DBaaS) from command line!
* [ybm integration create](ybm_integration_create.md)	 - Create Integration
* [ybm integration delete](ybm_integration_delete.md)	 - Delete Integration
* [ybm integration list](ybm_integration_list.md)	 - List Integration

