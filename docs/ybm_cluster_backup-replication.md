## ybm cluster backup-replication

Manage backup replication configuration for a cluster

### Synopsis

Manage backup replication configuration for a cluster. Supports cloud provider GCP.

```
ybm cluster backup-replication [flags]
```

### Options

```
  -h, --help   help for backup-replication
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

* [ybm cluster](ybm_cluster.md)	 - Manage cluster operations
* [ybm cluster backup-replication gcp](ybm_cluster_backup-replication_gcp.md)	 - Manage replication of cluster backups to GCP buckets

