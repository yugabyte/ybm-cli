## ybm cluster create

Create a cluster

### Synopsis

Create a cluster

```
ybm cluster create [flags]
```

### Options

```
      --cluster-name string          [REQUIRED] Name of the cluster.
      --credentials stringToString   [REQUIRED] Credentials to login to the cluster. Please provide key value pairs username=<user-name>,password=<password>. (default [])
      --cloud-provider string        [OPTIONAL] The cloud provider where database needs to be deployed. AWS or GCP. Default AWS.
      --cluster-type string          [OPTIONAL] Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED. Default SYNCHRONOUS.
      --node-config stringToInt      [OPTIONAL] Configuration of the cluster nodes. Please provide key value pairs num-cores=<num-cores>,disk-size-gb=<disk-size-gb> as the value. If specified  num-cores is mandatory, disk-size-gb is optional. (default [])
      --region-info stringArray      [OPTIONAL] Region information for the cluster. Please provide key value pairs region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name> as the value. If specified, region and num-nodes are mandatory, vpc is optional. Information about multiple regions can be specified by using multiple --region-info arguments. Default if not specified is us-west-2 AWS region.
      --cluster-tier string          [OPTIONAL] The tier of the cluster. Sandbox or Dedicated. Default Sandbox.
      --fault-tolerance string       [OPTIONAL] The fault tolerance of the cluster. The possible values are NONE, ZONE and REGION. Default NONE.
      --database-version string      [OPTIONAL] The database version of the cluster. Stable or Preview. Default depends on cluster tier, Sandbox is Preview, Dedicated is Stable.
      --encryption-spec string       [OPTIONAL] The customer managed key spec for the cluster.
                                     	Please provide key value pairs cloud-provider=AWS,aws-secret-key=<secret-key>,aws-access-key=<access-key>,aws-arn=<arn1>,aws-arn=<arn2> .
                                     	aws-access-key can be ommitted if the environment variable YBM_AWS_SECRET_KEY is set. If the environment variable is not set, the user will be prompted to enter the value.
                                     	If specified, all parameters for that provider are mandatory.
  -h, --help                         help for create
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

* [ybm cluster](ybm_cluster.md)	 - Manage cluster operations
