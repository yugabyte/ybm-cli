# ybm-cli
A CLI implementation for YB Managed.

## Sample Create Commands:

### Cluster
Minimal Required Configurations
```sh
ybm create cluster 
    --cluster-name=test-cluster
    --credentials=username=anonymous,password=password123
```
All possibilities
```sh
ybm create cluster 
    --cluster-name=test-cluster
    --credentials=username=anonymous,password=password123
    --cloud-type=[AWS or GCP]
    --cluster-type=[SYNCHRONOUS or GEO_PARTITIONED]
    --node-config=num_cores=2
    --region-info=region=<region-name>,num_nodes=<number-of-nodes>,vpc_id=<vpc-id>
    --cluster-tier=[FREE or PAID]
    --fault-tolerance=[NONE or ZONE or Region]
    --database-track=[STABLE or PREVIEW]
```

### Network Allow List
```sh
ybm create network-allow-list 
     --name=admins
     --description="admin allow list"
     --ip_addr=0.0.0.0/12, 1.1.1.1/12
```

### Read Replica
```sh
ybm create read-replica
     --replica=num_cores=<region-num_cores>,memory_mb=<memory_mb>,disk_size_gb=<disk_size_gb>,code=<GCP or AWS>,region=<region>,num_nodes=<num_nodes>,vpc_id=<vpc_id>,num_replicas=<num_replicas>,multi_zone=<multi_zone>
```
Additional replica can becreated by repeating the `replica` flag but each replic must be in a distinct region. 

### VPC
```sh
ybm create vpc
    --name=demo-vpc
    --cloud=GCP
    --global-cidr=10.0.0.0/18
```

```sh
ybm create vpc
    --name=demo-vpc
    --cloud=GCP
    --region=us-west1,uswest2
    --cidr=10.0.0.0/18,12.0.0.0/18
```


### VPC Peering
```sh
ybm create vpc
    --name=demo-peer
    --vpc-name=demo-vpc
    --cloud=GCP
    --project=project
    --vpc-id=10.0.0.0/12
    --region=us-west1
    --cidr=10.0.0.0/18
```
