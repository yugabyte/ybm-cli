## ybm cluster encryption list

List Encryption at Rest (EaR) configurations for a cluster

### Synopsis

List Encryption at Rest (EaR) configurations for a cluster

```
ybm cluster encryption list [flags]
```

### Options

```
  -h, --help   help for list
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
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster encryption](ybm_cluster_encryption.md)	 - Manage Encryption at Rest (EaR) for a cluster

