## ybm network-allow-list delete

Delete network allow list from YugabyteDB Aeon

### Synopsis

Delete network allow list from YugabyteDB Aeon

```
ybm network-allow-list delete [flags]
```

### Options

```
  -f, --force         Bypass the prompt for non-interactive usage
  -h, --help          help for delete
  -n, --name string   [REQUIRED] The name of the Network Allow List.
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

* [ybm network-allow-list](ybm_network-allow-list.md)	 - Manage Network Allow Lists

