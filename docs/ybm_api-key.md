## ybm api-key

Manage API Keys

### Synopsis

Manage API Keys in your YBM account

```
ybm api-key [flags]
```

### Options

```
  -h, --help   help for api-key
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

* [ybm](ybm.md)	 - ybm - Effortlessly manage your DB infrastructure on YugabyteDB Managed (DBaaS) from command line!
* [ybm api-key create](ybm_api-key_create.md)	 - Create an API Key
* [ybm api-key list](ybm_api-key_list.md)	 - List API Keys
* [ybm api-key revoke](ybm_api-key_revoke.md)	 - Revoke an API Key

