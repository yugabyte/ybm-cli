## ybm integration create

Create Integration

### Synopsis

Create Integration

```
ybm integration create [flags]
```

### Options

```
      --config-name string              [REQUIRED] The name of the Integration
      --type string                     [REQUIRED] The type of third party Integration sink
      --datadog-spec stringToString     Configuration for Datadog. 
                                        	Please provide key value pairs as follows: 
                                        	api-key=<your-datadog-api-key>,site=<your-datadog-site-parameters> (default [])
      --grafana-spec stringToString     Configuration for Grafana. 
                                        	Please provide key value pairs as follows: 
                                        	access-policy-token=<your-grafana-token>,zone=<your-grafana-zone-parameter>,instance-id=<your-grafana-instance-id>,org-slug=<your-grafana-org-slug> (default [])
      --sumologic-spec stringToString   Configuration for sumologic. 
                                        	Please provide key value pairs as follows: 
                                        	access-key=<your-sumologic-access-key>,access-id=<your-sumologic-access-id>,installation-token=<your-sumologic-installation-token> (default [])
  -h, --help                            help for create
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

* [ybm integration](ybm_integration.md)	 - Manage Integration

