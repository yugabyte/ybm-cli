## ybm vpc

Manage VPCs

### Synopsis

Manage VPCs

```
ybm vpc [flags]
```

### Options

```
  -h, --help   help for vpc
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon Api Key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm](ybm.md)	 - ybm - Effortlessly manage your DB infrastructure on YugabyteDB Aeon (DBaaS) from command line!
* [ybm vpc create](ybm_vpc_create.md)	 - Create a VPC in YugabyteDB Aeon
* [ybm vpc delete](ybm_vpc_delete.md)	 - Delete a VPC in YugabyteDB Aeon
* [ybm vpc list](ybm_vpc_list.md)	 - List VPCs in YugabyteDB Aeon
* [ybm vpc peering](ybm_vpc_peering.md)	 - Manage VPC Peerings

