## ybm user delete

Delete a user

### Synopsis

Delete a user from your YugabyteDB Aeon account

```
ybm user delete [flags]
```

### Options

```
      --email string   [REQUIRED] The email of the user.
  -f, --force          Bypass the prompt for non-interactive usage
  -h, --help           help for delete
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

* [ybm user](ybm_user.md)	 - Manage users

