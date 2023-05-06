## ybm vpc peering

Manage VPC Peerings

### Synopsis

Manage VPC Peerings

```
ybm vpc peering [flags]
```

### Options

```
  -h, --help   help for peering
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

* [ybm vpc](ybm_vpc.md)	 - Manage VPCs
* [ybm vpc peering create](ybm_vpc_peering_create.md)	 - Create VPC peering
* [ybm vpc peering delete](ybm_vpc_peering_delete.md)	 - Delete VPC peering
* [ybm vpc peering list](ybm_vpc_peering_list.md)	 - List VPC peerings

