## ybm region list

List Cloud Provider Regions

### Synopsis

List Cloud Provider Regions

```
ybm region list [flags]
```

### Options

```
      --cloud-provider string   [REQUIRED] The cloud provider for which the regions have to be fetched. AWS, AZURE or GCP.
  -h, --help                    help for list
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

* [ybm region](ybm_region.md)	 - Manage cloud regions

