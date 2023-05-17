## ybm role update

Update a custom role

### Synopsis

Update a custom role in YB Managed

```
ybm role update [flags]
```

### Options

```
      --description string        [OPTIONAL] New description of the role to be updated.
  -h, --help                      help for update
      --new-name string           [OPTIONAL] New name of the role to be updated.
      --permissions stringArray   [REQUIRED] Permissions for the role. Please provide key value pairs resource-type=<resource-type>,operation-group=<operation-group> as the value. Both resource-type and operation-group are mandatory. Information about multiple permissions can be specified by using multiple --permissions arguments.
      --role-name string          [REQUIRED] Name of the role.
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

* [ybm role](ybm_role.md)	 - Manage roles

