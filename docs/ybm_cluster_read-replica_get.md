## ybm cluster read-replica get

Get read replica

### Synopsis

Get read replica in YugabyteDB Managed

```
ybm cluster read-replica get [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -a, --apiKey string         YBM Api Key
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
      --config string         config file (default is $HOME/.ybm-cli.yaml)
      --debug                 Use debug mode, same as --logLevel debug
  -l, --logLevel string       Select the desired log level format(info). Default to info
      --no-color              Disable colors in output , default to false
  -o, --output string         Select the desired output format (table, json, pretty). Default to table
      --timeout duration      Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster read-replica](ybm_cluster_read-replica.md)	 - Manage Read Replicas
