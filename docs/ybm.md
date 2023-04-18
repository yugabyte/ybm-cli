## ybm

ybm - Effortlessly manage your DB infrastructure on YugabyteDB Managed (DBaaS) from command line!

### Synopsis

ybm - Effortlessly manage your DB infrastructure on YugabyteDB Managed (DBaaS) from command line!

```
ybm [flags]
```

### Options

```
  -a, --apiKey string      YBM Api Key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -h, --help               help for ybm
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
  -t, --toggle             Help message for toggle
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm auth](ybm_auth.md)	 - Authenticate ybm CLI
* [ybm backup](ybm_backup.md)	 - Manage backup operations of a cluster
* [ybm cluster](ybm_cluster.md)	 - Manage cluster operations
* [ybm completion](ybm_completion.md)	 - Generate the autocompletion script for the specified shell
* [ybm network-allow-list](ybm_network-allow-list.md)	 - Manage Network Allow Lists
* [ybm region](ybm_region.md)	 - Manage cloud regions
* [ybm signup](ybm_signup.md)	 - Open a browser to sign up for YugabyteDB Managed
* [ybm tools](ybm_tools.md)	 - Tools command
* [ybm vpc](ybm_vpc.md)	 - Manage VPCs

