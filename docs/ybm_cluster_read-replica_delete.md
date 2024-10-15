## ybm cluster read-replica delete

Delete read replica

### Synopsis

Delete read replica from YugabyteDB Aeon

```
ybm cluster read-replica delete [flags]
```

### Options

```
  -f, --force   Bypass the prompt for non-interactive usage
  -h, --help    help for delete
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

