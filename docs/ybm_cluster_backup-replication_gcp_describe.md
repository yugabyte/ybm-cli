## ybm cluster backup-replication gcp describe

Describe the current GCP backup replication configuration

### Synopsis

Describe the current GCP backup replication configuration for the specified cluster

```
ybm cluster backup-replication gcp describe [flags]
```

### Options

```
  -h, --help       help for describe
      --show-all   Show all configurations including scheduled for expiry and removed regions
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

