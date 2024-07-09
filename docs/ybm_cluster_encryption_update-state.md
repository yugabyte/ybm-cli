## ybm cluster encryption update-state

Update Encryption at Rest (EaR) state for a cluster

### Synopsis

Update Encryption at Rest (EaR) state for a cluster

```
ybm cluster encryption update-state [flags]
```

### Options

```
      --disable   Disable EAR
      --enable    Enable EAR
  -h, --help      help for update-state
```

### Options inherited from parent commands

```
  -a, --apiKey string         YugabyteDB Aeon Api Key
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

* [ybm cluster encryption](ybm_cluster_encryption.md)	 - Manage Encryption at Rest (EaR) for a cluster

