## ybm cluster read-replica create

Create read replica

### Synopsis

Create read replica in YugabyteDB Aeon

```
ybm cluster read-replica create [flags]
```

### Options

```
  -h, --help                  help for create
  -r, --replica stringArray   [OPTIONAL] Region information for the cluster. Please provide key value pairs num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,disk-iops=<disk-iops>,cloud-provider=<GCP or AWS>,region=<region>,num-nodes=<num-nodes>,vpc=<vpc-name>,num-replicas=<num-replicas>,multi-zone=<multi-zone>.
```

### Options inherited from parent commands

```
  -a, --apiKey string         YugabyteDB Aeon account API key
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
      --config string         config file (default is $HOME/.ybm-cli.yaml)
      --debug                 Use debug mode, same as --logLevel debug
      --host string           YugabyteDB Aeon Api hostname
  -l, --logLevel string       Select the desired log level format(info). Default to info
      --no-color              Disable colors in output , default to false
  -o, --output string         Select the desired output format (table, json, pretty). Default to table
      --timeout duration      Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster read-replica](ybm_cluster_read-replica.md)	 - Manage Read Replicas

