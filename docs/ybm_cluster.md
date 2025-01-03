## ybm cluster

Manage cluster operations

### Synopsis

Manage cluster operations

```
ybm cluster [flags]
```

### Options

```
  -h, --help   help for cluster
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
* [ybm cluster cert](ybm_cluster_cert.md)	 - Get the root CA certificate
* [ybm cluster create](ybm_cluster_create.md)	 - Create a cluster
* [ybm cluster db-audit-logging](ybm_cluster_db-audit-logging.md)	 - Configure Database Audit Logging for your Cluster.
* [ybm cluster db-query-logging](ybm_cluster_db-query-logging.md)	 - Configure Database Query Logging for your Cluster.
* [ybm cluster delete](ybm_cluster_delete.md)	 - Delete a cluster
* [ybm cluster describe](ybm_cluster_describe.md)	 - Describe a cluster
* [ybm cluster encryption](ybm_cluster_encryption.md)	 - Manage Encryption at Rest (EaR) for a cluster
* [ybm cluster list](ybm_cluster_list.md)	 - List clusters
* [ybm cluster namespace](ybm_cluster_namespace.md)	 - Manage Cluster Namespaces
* [ybm cluster network](ybm_cluster_network.md)	 - Manage network operations
* [ybm cluster node](ybm_cluster_node.md)	 - Manage nodes for a cluster
* [ybm cluster pause](ybm_cluster_pause.md)	 - Pause a cluster
* [ybm cluster read-replica](ybm_cluster_read-replica.md)	 - Manage Read Replicas
* [ybm cluster resume](ybm_cluster_resume.md)	 - Resume a cluster
* [ybm cluster update](ybm_cluster_update.md)	 - Update a cluster

