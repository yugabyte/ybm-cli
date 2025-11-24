## ybm cluster backup-replication gcp enable

Enable GCP backup replication for a cluster

### Synopsis

Enable GCP backup replication for all backup regions in the cluster

```
ybm cluster backup-replication gcp enable [flags]
```

### Options

```
  -h, --help                        help for enable
      --region-target stringArray   [REQUIRED] Specify region and bucket pairs for backup replication. Format: region=<region-name>,bucket-name=<bucket-name>. Must be specified for each backup region in the cluster.
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

* [ybm cluster backup-replication gcp](ybm_cluster_backup-replication_gcp.md)	 - Manage replication of cluster backups to GCP buckets

