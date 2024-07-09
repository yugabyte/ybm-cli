## ybm cluster cert download

Download the root CA certificate

### Synopsis

Download the root CA certificate

```
ybm cluster cert download [flags]
```

### Options

```
  -f, --force        [OPTIONAL] Overwrite the output file if it exists
  -h, --help         help for download
      --out string   [OPTIONAL] Output file name (default: stdout)
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster cert](ybm_cluster_cert.md)	 - Get the root CA certificate

