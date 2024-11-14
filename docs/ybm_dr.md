## ybm dr

Manage DR for a cluster.

### Synopsis

Manage DR for a cluster.

```
ybm dr [flags]
```

### Options

```
  -h, --help   help for dr
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

* [ybm](ybm.md)	 - ybm - Effortlessly manage your DB infrastructure on YugabyteDB Aeon (DBaaS) from command line!
* [ybm dr config](ybm_dr_config.md)	 - Manage DR config
* [ybm dr failover](ybm_dr_failover.md)	 - Failover DR for a cluster
* [ybm dr pause](ybm_dr_pause.md)	 - Pause DR for a cluster
* [ybm dr restart](ybm_dr_restart.md)	 - Restart DR for a cluster
* [ybm dr resume](ybm_dr_resume.md)	 - Resume DR for a cluster
* [ybm dr switchover](ybm_dr_switchover.md)	 - Switchover DR for a cluster

