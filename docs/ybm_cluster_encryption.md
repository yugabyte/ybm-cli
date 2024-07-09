## ybm cluster encryption

Manage Encryption at Rest (EaR) for a cluster

### Synopsis

Manage Encryption at Rest (EaR) for a cluster

```
ybm cluster encryption [flags]
```

### Options

```
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
  -h, --help                  help for encryption
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon Api Key
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
* [ybm cluster encryption list](ybm_cluster_encryption_list.md)	 - List Encryption at Rest (EaR) configurations for a cluster
* [ybm cluster encryption update](ybm_cluster_encryption_update.md)	 - Update Encryption at Rest (EaR) configurations for a cluster
* [ybm cluster encryption update-state](ybm_cluster_encryption_update-state.md)	 - Update Encryption at Rest (EaR) state for a cluster

