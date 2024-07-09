## ybm region instance list

List the Instance Types for a region

### Synopsis

List the Instance Types for a region

```
ybm region instance list [flags]
```

### Options

```
      --cloud-provider string   [REQUIRED] The cloud provider for which the regions have to be fetched. AWS, AZURE or GCP.
      --region string           [REQUIRED] The region in the cloud provider for which the instance types have to fetched.
      --tier string             [OPTIONAL] Tier. Sandbox or Dedicated. (default "Dedicated")
      --show-disabled           [OPTIONAL] Whether to show disabled instance types. true or false. (default true)
  -h, --help                    help for list
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

* [ybm region instance](ybm_region_instance.md)	 - Manage instance types

