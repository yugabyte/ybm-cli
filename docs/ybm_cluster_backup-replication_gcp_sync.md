## ybm cluster backup-replication gcp sync

Trigger resync of GCP backup replication for a cluster

### Synopsis

Trigger resync of backup data for a cluster. Creates one-time transfer operations for existing transfer jobs to synchronize backup data to target GCS buckets.

```
ybm cluster backup-replication gcp sync [flags]
```

### Options

```
  -h, --help   help for sync
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

