## ybm region instance

Manage instance types

### Synopsis

Manage instance types for your YugabyteDB Aeon clusters

```
ybm region instance [flags]
```

### Options

```
  -h, --help   help for instance
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

* [ybm region](ybm_region.md)	 - Manage cloud regions
* [ybm region instance list](ybm_region_instance_list.md)	 - List the Instance Types for a region

