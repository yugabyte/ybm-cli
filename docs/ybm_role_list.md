## ybm role list

List roles

### Synopsis

List roles in YugabyteDB Aeon

```
ybm role list [flags]
```

### Options

```
      --role-name string   [OPTIONAL] To filter by role name.
      --type string        [OPTIONAL] To filter by role type. BUILT-IN and CUSTOM options are available to list only built-in or custom roles.
  -h, --help               help for list
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon Api Key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm role](ybm_role.md)	 - Manage roles

