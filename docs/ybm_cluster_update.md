## ybm cluster update

Update a cluster

### Synopsis

Update a cluster

```
ybm cluster update [flags]
```

### Options

```
      --cloud-provider string     [OPTIONAL] The cloud provider where database needs to be deployed. AWS, AZURE or GCP.
      --cluster-name string       [REQUIRED] Name of the cluster.
      --cluster-tier string       [OPTIONAL] The tier of the cluster. Sandbox or Dedicated.
      --cluster-type string       [OPTIONAL] Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED.
      --database-version string   [OPTIONAL] The database version of the cluster. Production, Innovation, Preview, or 'Early Access'.
      --fault-tolerance string    [OPTIONAL] Fault tolerance of the cluster. The possible values are NONE, NODE, ZONE, or REGION. Default NONE.
  -h, --help                      help for update
      --new-name string           [OPTIONAL] The new name to be given to the cluster.
      --region-info stringArray   Region information for the cluster, provided as key-value pairs. Arguments are region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name>,num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,disk-iops=<disk-iops> (AWS only). region, num-nodes, num-cores, disk-size-gb are required. Specify one --region-info flag for each region in the cluster.
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

* [ybm cluster](ybm_cluster.md)	 - Manage cluster operations

