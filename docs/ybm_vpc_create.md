## ybm vpc create

Create a VPC in YugabyteDB Managed

### Synopsis

Create a VPC in YugabyteDB Managed

```
ybm vpc create [flags]
```

### Options

```
      --name string             [REQUIRED] Name for the VPC.
      --cloud-provider string   [REQUIRED] Cloud provider for the VPC: AWS, AZURE or GCP.
      --global-cidr string      [OPTIONAL] Global CIDR for the VPC.
      --region strings          [OPTIONAL] Region of the VPC.
      --cidr strings            [OPTIONAL] CIDR of the VPC.
  -h, --help                    help for create
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

* [ybm vpc](ybm_vpc.md)	 - Manage VPCs

