## ybm cluster encryption update

Update Encryption at Rest (EaR) configurations for a cluster

### Synopsis

Update Encryption at Rest (EaR) configurations for a cluster

```
ybm cluster encryption update [flags]
```

### Options

```
      --encryption-spec string   [REQUIRED] The customer managed key spec for the cluster.
                                 	Please provide key value pairs as follows:
                                 	For AWS: 
                                 	cloud-provider=AWS,aws-secret-key=<secret-key>,aws-access-key=<access-key>,aws-arn=<arn1>,aws-arn=<arn2> .
                                 	aws-access-key can be ommitted if the environment variable YBM_AWS_SECRET_KEY is set. If the environment variable is not set, the user will be prompted to enter the value.
                                 	For GCP:
                                 	cloud-provider=GCP,gcp-resource-id=<resource-id>,gcp-service-account-path=<service-account-path>.
                                 	For AZURE:
                                 	cloud-provider=AZURE,azu-client-id=<client-id>,azu-client-secret=<client-secret>,azu-tenant-id=<tenant-id>,azu-key-name=<key-name>,azu-key-vault-uri=<key-vault-uri>.
                                 	
  -h, --help                     help for update
```

### Options inherited from parent commands

```
  -a, --apiKey string         YugabyteDB Aeon account API key
  -c, --cluster-name string   [REQUIRED] The name of the cluster.
      --config string         config file (default is $HOME/.ybm-cli.yaml)
      --debug                 Use debug mode, same as --logLevel debug
      --host string           YugabyteDB Aeon Api hostname
  -l, --logLevel string       Select the desired log level format(info). Default to info
      --no-color              Disable colors in output , default to false
  -o, --output string         Select the desired output format (table, json, pretty). Default to table
      --timeout duration      Wait command timeout, example: 5m, 1h. (default 168h0m0s)
      --wait                  Wait until the task is completed, otherwise it will exit immediately, default to false
```

### SEE ALSO

* [ybm cluster encryption](ybm_cluster_encryption.md)	 - Manage Encryption at Rest (EaR) for a cluster

