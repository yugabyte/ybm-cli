## ybm cluster create

Create a cluster

### Synopsis

Create a cluster

```
ybm cluster create [flags]
```

### Options

```
      --cluster-name string            [REQUIRED] Name of the cluster.
      --credentials stringToString     [REQUIRED] Credentials to login to the cluster. Please provide key value pairs username=<user-name>,password=<password>. (default [])
      --cloud-provider string          [OPTIONAL] The cloud provider where database needs to be deployed. AWS, AZURE or GCP. Default AWS.
      --cluster-tier string            [OPTIONAL] The tier of the cluster. Sandbox or Dedicated. Default Sandbox.
      --cluster-type string            [OPTIONAL] Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED. Default SYNCHRONOUS.
      --database-version string        [OPTIONAL] The database version of the cluster. Production, Innovation, Preview, or 'Early Access'. Default depends on cluster tier, Sandbox is Preview, Dedicated is Production.
      --encryption-spec string         [OPTIONAL] The customer managed key spec for the cluster.
                                       	Please provide key value pairs as follows:
                                       	For AWS: 
                                       	cloud-provider=AWS,aws-secret-key=<secret-key>,aws-access-key=<access-key>,aws-arn=<arn1>,aws-arn=<arn2> .
                                       	aws-access-key can be ommitted if the environment variable YBM_AWS_SECRET_KEY is set. If the environment variable is not set, the user will be prompted to enter the value.
                                       	For GCP:
                                       	cloud-provider=GCP,gcp-resource-id=<resource-id>,gcp-service-account-path=<service-account-path>.
                                       	For AZURE:
                                       	cloud-provider=AZURE,azu-client-id=<client-id>,azu-client-secret=<client-secret>,azu-tenant-id=<tenant-id>,azu-key-name=<key-name>,azu-key-vault-uri=<key-vault-uri>.
                                       	If specified, all parameters for that provider are mandatory.
      --fault-tolerance string         [OPTIONAL] Fault tolerance of the cluster. The possible values are NONE, NODE, ZONE, or REGION. Default NONE.
      --num-faults-to-tolerate int32   [OPTIONAL] The number of domain faults to tolerate for the level specified. The possible values are 0 for NONE, 1 for ZONE and [1-3] for anything else. Defaults to 0 for NONE, 1 otherwise.
      --region-info stringArray        Region information for the cluster, provided as key-value pairs. Arguments are region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name>,num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,disk-iops=<disk-iops> (AWS only). region, num-nodes, num-cores, disk-size-gb are required. Specify one --region-info flag for each region in the cluster.
      --preferred-region string        [OPTIONAL] The preferred region in a multi region cluster. The preferred region handles all read and write requests from clients.
      --default-region string          [OPTIONAL] The primary region in a partition-by-region cluster. The primary region is where all the tables not created in a tablespace reside.
  -h, --help                           help for create
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

