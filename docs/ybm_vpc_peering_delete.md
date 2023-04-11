## ybm vpc peering delete

Delete VPC peering

### Synopsis

Delete VPC peering in YugabyteDB Managed

```
ybm vpc peering delete [flags]
```

### Options

```
  -f, --force         Bypass the prompt for non-interactive usage
  -h, --help          help for delete
      --name string   [REQUIRED] Name for the VPC peering.
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

* [ybm vpc peering](ybm_vpc_peering.md)	 - Manage VPC Peerings

