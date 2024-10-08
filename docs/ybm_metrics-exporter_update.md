## ybm metrics-exporter update

Update Metrics Exporter Config

### Synopsis

Update Metrics Exporter Config

```
ybm metrics-exporter update [flags]
```

### Options

```
      --config-name string              [REQUIRED] The name of the metrics exporter configuration
      --type string                     [REQUIRED] The type of third party metrics sink
      --new-config-name string          [OPTIONAL] The new name of the metrics exporter configuration
      --datadog-spec stringToString     Configuration for Datadog. 
                                        	Please provide key value pairs as follows: 
                                        	api-key=<your-datadog-api-key>,site=<your-datadog-site-parameters> (default [])
      --grafana-spec stringToString     Configuration for Grafana. 
                                        	Please provide key value pairs as follows: 
                                        	access-policy-token=<your-grafana-token>,zone=<your-grafana-zone-parameter>,instance-id=<your-grafana-instance-id>,org-slug=<your-grafana-org-slug> (default [])
      --sumologic-spec stringToString   Configuration for sumologic. 
                                        	Please provide key value pairs as follows: 
                                        	access-key=<your-sumologic-access-key>,access-id=<your-sumologic-access-id>,installation-token=<your-sumologic-installation-token> (default [])
  -h, --help                            help for update
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

* [ybm metrics-exporter](ybm_metrics-exporter.md)	 - Manage Metrics Exporter

