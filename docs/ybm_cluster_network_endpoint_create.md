## ybm cluster network endpoint create

Create a new network endpoint for a cluster

### Synopsis

Create a new network endpoint for a cluster

```
ybm cluster network endpoint create [flags]
```

### Options

```
      --accessibility-type string    [REQUIRED] The accessibility of the endpoint. Valid options are PUBLIC, PRIVATE and PRIVATE_SERVICE_ENDPOINT.
  -h, --help                         help for create
      --region string                [REQUIRED] The region of the endpoint.
      --security-principals string   [OPTIONAL] The list of security principals that have access to this endpoint. Required for private service endpoints.  Accepts a comma separated list. E.g.: arn:aws:iam::account_id1:root,arn:aws:iam::account_id2:root
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

* [ybm cluster network endpoint](ybm_cluster_network_endpoint.md)	 - Manage network endpoints for a cluster

