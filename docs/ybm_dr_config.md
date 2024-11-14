## ybm dr config

Manage DR config

### Synopsis

Manage DR config

```
ybm dr config [flags]
```

### Options

```
  -h, --help   help for config
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

* [ybm dr](ybm_dr.md)	 - Manage DR for a cluster.
* [ybm dr config create](ybm_dr_config_create.md)	 - Create DR for a cluster
* [ybm dr config delete](ybm_dr_config_delete.md)	 - Delete DR
* [ybm dr config describe](ybm_dr_config_describe.md)	 - Describe DR
* [ybm dr config list](ybm_dr_config_list.md)	 - List DRs for a given cluster
* [ybm dr config update](ybm_dr_config_update.md)	 - Update DR for a cluster

