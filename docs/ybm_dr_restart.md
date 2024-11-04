## ybm dr restart

Restart DR for a cluster

### Synopsis

Restart DR for a cluster

```
ybm dr restart [flags]
```

### Options

```
      --databases stringArray   [OPTIONAL] Databases to be restarted. Please provide a comma separated list of database names <db-name-1>,<db-name-2>.
      --dr-name string          [REQUIRED] Name of the DR configuration.
  -h, --help                    help for restart
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

* [ybm dr](ybm_dr.md)	 - Manage DR for a cluster.

