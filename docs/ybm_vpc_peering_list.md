## ybm vpc peering list

List VPC peerings

### Synopsis

List VPC peerings in YugabyteDB Aeon

```
ybm vpc peering list [flags]
```

### Options

```
  -h, --help          help for list
      --name string   [OPTIONAL] Name for the VPC peering.
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

* [ybm vpc peering](ybm_vpc_peering.md)	 - Manage VPC Peerings

