## ybm cluster backup-replication gcp

Manage replication of cluster backups to GCP buckets

### Synopsis

Manage replication of cluster backups to GCP buckets

```
ybm cluster backup-replication gcp [flags]
```

### Options

```
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
  -h, --help                  help for gcp
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
      --host string        YugabyteDB Aeon Api hostname
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster backup-replication](ybm_cluster_backup-replication.md)	 - Manage backup replication configuration for a cluster
* [ybm cluster backup-replication gcp describe](ybm_cluster_backup-replication_gcp_describe.md)	 - Describe the current GCP backup replication configuration
* [ybm cluster backup-replication gcp enable](ybm_cluster_backup-replication_gcp_enable.md)	 - Enable GCP backup replication for a cluster
* [ybm cluster backup-replication gcp update](ybm_cluster_backup-replication_gcp_update.md)	 - Update GCP backup replication configuration for a cluster

