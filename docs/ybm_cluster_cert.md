## ybm cluster cert

Get the root CA certificate

### Synopsis

Get the root CA certificate for your YBM clusters

```
ybm cluster cert [flags]
```

### Options

```
  -h, --help   help for cert
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
* [ybm cluster cert download](ybm_cluster_cert_download.md)	 - Download the root CA certificate

