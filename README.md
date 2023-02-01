# ybm-cli
A CLI implementation for YB Managed.

# Global configuration
This CLI support 3 possibles configurations: 
* Passing values as flags 
  ```shell
  ybm --apikey AWERDFSSS --host cloud.yugabyte.com get cluster
  ```
* Using a configuration file called `.ybm-cli.yaml` under your `$HOME` directory.
  You can use the command `ybm configure` to help to setup the file

* Using environment variables (all need to start with `YBM_`)
 ```shell
 export YBM_APIKEY=AWERDFSSS
 export YBM_HOST=cloud.yugabyte.com
 ybm get cluster
 ```

By default, `https` will be added to the host if no scheme are provided if you want to use `http` 
just add it to the host `http://cloud.yugabyte.com`


## Sample Create Commands:

### Cluster
Minimal Required Configurations
```sh
ybm cluster create 
    --cluster-name=test-cluster
    --credentials=username=anonymous,password=password123
```
All possibilities
```sh
ybm cluster create
    --cluster-name=test-cluster
    --credentials=username=anonymous,password=password123
    --cloud-type=[AWS or GCP]
    --cluster-type=[SYNCHRONOUS or GEO_PARTITIONED]
    --node-config=num-cores=<num-cores>,disk-size-gb=<disk-size-gb>
    --region-info=region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name>
    --cluster-tier=[FREE or PAID]
    --fault-tolerance=[NONE or ZONE or Region]
    --database-track=[STABLE or PREVIEW]
```

### Network Allow List
```sh
ybm network-allow-list create
     --name=admins
     --description="admin allow list"
     --ip-addr=0.0.0.0/12, 1.1.1.1/12
```

### Read Replica
```sh
ybm read-replica create
     --replica=num-cores=<num-cores>,memory-mb=<memory-mb>,disk-size-gb=<disk-size-gb>,code=<GCP or AWS>,region=<region>,num-nodes=<num-nodes>,vpc=<vpc-name>,num-replicas=<num-replicas>,multi-zone=<multi-zone>
```
Additional replica can becreated by repeating the `replica` flag but each replica must be in a distinct region. 

### VPC
```sh
ybm vpc create
    --name=demo-vpc
    --cloud=GCP
    --global-cidr=10.0.0.0/18
```

```sh
ybm vpc create
    --name=demo-vpc
    --cloud=GCP
    --region=us-west1,us-west2
    --cidr=10.0.0.0/18,12.0.0.0/18
```


### VPC Peering
```sh
ybm vpc-peering create
    --name=demo-peer
    --vpc-name=demo-vpc
    --cloud=GCP
    --project=project
    --vpc=vpc-name
    --region=us-west1
    --cidr=10.0.0.0/18
```

### CDC Sink

```sh
ybm cdc-sink create
    --name=sink-2 
    --hostname=kafka.self.us 
    --auth-type=BASIC 
    --cdc-sink-type=KAFKA 
    --username=something 
    --password=something
```

```sh
./ybm cdc-sink get
    --name=sink-2
```

```sh
./ybm cdc-sink update
    --name=sink-2 
    --new-name=new-sink-2
```

```sh
./ybm cdc-sink delete
    --name=sink-2 
```




## Developer Workflow
- `cd ~/code`
- `git clone git@github.com:yugabyte/yugabyte-cloud.git`
- `cd yugabyte-cloud/apiserver`
- `git checkout -b sample-api`
-  Develop a new API - Sample PR: https://github.com/yugabyte/yugabyte-cloud/pull/7078/files
- `sbt cleanCloud`
- `sbt compileCloud` - This will generate the go client
- `cd ..`
- `git clone git@github.com:yugabyte/yugabytedb-managed-go-client-internal.git`
- `cd yugabytedb-managed-go-client-internal`
- `git checkout -b client-with-sample-api`
- `cp -r ~/code/yugabyte-cloud/apiserver/openapi-clients/clients/public/go/generated/ .` - Copy the go public client to the `client-with-sample-api` branch in `yugabytedb-managed-go-client-internal` repository
- Set the module name in go.mod file name so that the first line of the file looks like `module github.com/yugabyte/yugabytedb-managed-go-client-internal`
- `git push origin client-with-sample-api` - Sample PR: https://github.com/yugabyte/yugabytedb-managed-go-client-internal/pull/1/files
- `cd ..`
- `git clone git@github.com:yugabyte/ybm-cli.git`
- `cd ybm-cli`
- `export GOPRIVATE=github.com/yugabyte`
- `go get github.com/yugabyte/yugabytedb-managed-go-client-internal@client-with-sample-api` - Notice the branch name after `@`. It should be the same as the branch in `yugabytedb-managed-go-client-internal` repository where we copied the autogenerated client code from the `yugabyte-cloud` repository after adding the sample API.
- Make the changes to the ybm cli utlizing the internal go client - Sample PR: https://github.com/yugabyte/ybm-cli/pull/12/files
- `export YBM_HOST=localhost:9000`
- `make build`
- `./ybm greet`

## Update CLI
- make update-cli
