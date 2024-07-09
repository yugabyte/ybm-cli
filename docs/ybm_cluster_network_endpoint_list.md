## ybm cluster network endpoint list

List network endpoints for a cluster

### Synopsis

List network endpoints for a cluster

```
ybm cluster network endpoint list [flags]
```

### Options

```
      --accessibility-type string   [OPTIONAL] Accessibility of the endpoint. Valid options are PUBLIC, PRIVATE and PRIVATE_SERVICE_ENDPOINT.
  -h, --help                        help for list
      --region string               [OPTIONAL] The region of the endpoint.
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

* [ybm cluster network endpoint](ybm_cluster_network_endpoint.md)	 - Manage network endpoints for a cluster

