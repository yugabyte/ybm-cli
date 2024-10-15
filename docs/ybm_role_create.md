## ybm role create

Create a custom role

### Synopsis

Create a custom role in YugabyteDB Aeon

```
ybm role create [flags]
```

### Options

```
      --role-name string          [REQUIRED] Name of the role to be created.
      --permissions stringArray   [REQUIRED] Permissions for the role. Please provide key value pairs resource-type=<resource-type>,operation-group=<operation-group> as the value. Both resource-type and operation-group are mandatory. Information about multiple permissions can be specified by using multiple --permissions arguments.
      --description string        [OPTIONAL] Description of the role to be created.
  -f, --force                     Bypass the prompt for non-interactive usage
  -h, --help                      help for create
```

### Options inherited from parent commands

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
      --host string        YugabyteDB Aeon Api hostname
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm role](ybm_role.md)	 - Manage roles

