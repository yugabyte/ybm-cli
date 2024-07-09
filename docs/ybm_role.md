## ybm role

Manage roles

### Synopsis

Manage roles

```
ybm role [flags]
```

### Options

```
  -h, --help   help for role
```

### Options inherited from parent commands

```
  -a, --apiKey string      YBM Api Key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
      --host string        YBM Api hostname
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm](ybm.md)	 - ybm - Effortlessly manage your DB infrastructure on YugabyteDB Managed (DBaaS) from command line!
* [ybm role create](ybm_role_create.md)	 - Create a custom role
* [ybm role delete](ybm_role_delete.md)	 - Delete a custom role
* [ybm role describe](ybm_role_describe.md)	 - Describe a role
* [ybm role list](ybm_role_list.md)	 - List roles
* [ybm role update](ybm_role_update.md)	 - Update a custom role

