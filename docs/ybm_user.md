## ybm user

Manage users

### Synopsis

Manage users in your YugabyteDB Aeon account

```
ybm user [flags]
```

### Options

```
  -h, --help   help for user
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

* [ybm](ybm.md)	 - ybm - Effortlessly manage your DB infrastructure on YugabyteDB Aeon (DBaaS) from command line!
* [ybm user delete](ybm_user_delete.md)	 - Delete a user
* [ybm user invite](ybm_user_invite.md)	 - Invite a user
* [ybm user list](ybm_user_list.md)	 - List users
* [ybm user update](ybm_user_update.md)	 - Modify role of a user

