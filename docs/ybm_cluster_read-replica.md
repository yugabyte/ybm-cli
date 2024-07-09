## ybm cluster read-replica

Manage Read Replicas

### Synopsis

Manage Read Replicas

```
ybm cluster read-replica [flags]
```

### Options

```
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
  -h, --help                  help for read-replica
```

### Options inherited from parent commands

```
  -a, --apiKey string      YBM Api Key
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
* [ybm cluster read-replica create](ybm_cluster_read-replica_create.md)	 - Create read replica
* [ybm cluster read-replica delete](ybm_cluster_read-replica_delete.md)	 - Delete read replica
* [ybm cluster read-replica list](ybm_cluster_read-replica_list.md)	 - List read replicas
* [ybm cluster read-replica update](ybm_cluster_read-replica_update.md)	 - Edit read replica

