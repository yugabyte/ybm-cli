## ybm

ybm - Effortlessly manage your DB infrastructure on YugabyteDB Aeon (DBaaS) from command line!

### Synopsis

ybm - Effortlessly manage your DB infrastructure on YugabyteDB Aeon (DBaaS) from command line!

```
ybm [flags]
```

### Options

```
  -a, --apiKey string      YugabyteDB Aeon account API key
      --config string      config file (default is $HOME/.ybm-cli.yaml)
      --debug              Use debug mode, same as --logLevel debug
  -h, --help               help for ybm
      --host string        YugabyteDB Aeon Api hostname
  -l, --logLevel string    Select the desired log level format(info). Default to info
      --no-color           Disable colors in output , default to false
  -o, --output string      Select the desired output format (table, json, pretty). Default to table
      --timeout duration   Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait               Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm api-key](ybm_api-key.md)	 - Manage API Keys
* [ybm auth](ybm_auth.md)	 - Authenticate ybm CLI
* [ybm backup](ybm_backup.md)	 - Manage backup operations of a cluster
* [ybm cluster](ybm_cluster.md)	 - Manage cluster operations
* [ybm completion](ybm_completion.md)	 - Generate the autocompletion script for the specified shell
* [ybm integration](ybm_integration.md)	 - Manage Integration
* [ybm metrics-exporter](ybm_metrics-exporter.md)	 - Manage Metrics Exporter
* [ybm network-allow-list](ybm_network-allow-list.md)	 - Manage Network Allow Lists
* [ybm permission](ybm_permission.md)	 - View available permissions for roles
* [ybm region](ybm_region.md)	 - Manage cloud regions
* [ybm role](ybm_role.md)	 - Manage roles
* [ybm signup](ybm_signup.md)	 - Open a browser to sign up for YugabyteDB Aeon
* [ybm usage](ybm_usage.md)	 - Billing usage for the account in YugabyteDB Aeon
* [ybm user](ybm_user.md)	 - Manage users
* [ybm vpc](ybm_vpc.md)	 - Manage VPCs

