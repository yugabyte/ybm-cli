## ybm vpc peering create

Create VPC peering

### Synopsis

Create VPC peering in YugabyteDB Aeon

```
ybm vpc peering create [flags]
```

### Options

```
      --name string                 [REQUIRED] Name for the VPC peering.
      --yb-vpc-name string          [REQUIRED] Name of the YugabyteDB Aeon VPC.
      --cloud-provider string       [REQUIRED] Cloud of the VPC with which to peer. AWS or GCP.
      --app-vpc-name string         [OPTIONAL] Name of the application VPC. Required for GCP. Not applicable for AWS.
      --app-vpc-project-id string   [OPTIONAL] Project ID of the application VPC. Required for GCP. Not applicable for AWS.
      --app-vpc-cidr string         [OPTIONAL] CIDR of the application VPC. Required for AWS. Optional for GCP.
      --app-vpc-account-id string   [OPTIONAL] Account ID of the application VPC. Required for AWS. Not applicable for GCP.
      --app-vpc-id string           [OPTIONAL] ID of the application VPC. Required for AWS. Not applicable for GCP.
      --app-vpc-region string       [OPTIONAL] Region of the application VPC. Required for AWS. Not applicable for GCP.
  -h, --help                        help for create
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

* [ybm vpc peering](ybm_vpc_peering.md)	 - Manage VPC Peerings

