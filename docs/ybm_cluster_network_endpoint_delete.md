## ybm cluster network endpoint delete

Delete a network endpoint for a cluster

### Synopsis

Delete a network endpoint for a cluster

```
ybm cluster network endpoint delete [flags]
```

### Options

```
      --endpoint-id string   [REQUIRED] THe ID of the endpoint
  -f, --force                Bypass the prompt for non-interactive usage
  -h, --help                 help for delete
```

### Options inherited from parent commands

```
  -a, --apiKey string         YBM Api Key
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
      --config string         config file (default is $HOME/.ybm-cli.yaml)
      --debug                 Use debug mode, same as --logLevel debug
      --host string           YBM Api hostname
  -l, --logLevel string       Select the desired log level format(info). Default to info
      --no-color              Disable colors in output , default to false
  -o, --output string         Select the desired output format (table, json, pretty). Default to table
      --timeout duration      Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster network endpoint](ybm_cluster_network_endpoint.md)	 - Manage network endpoints for a cluster

