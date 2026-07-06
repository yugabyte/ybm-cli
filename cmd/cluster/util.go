// Licensed to Yugabyte, Inc. under one or more contributor license
// agreements. See the NOTICE file distributed with this work for
// additional information regarding copyright ownership. Yugabyte
// licenses this file to you under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cluster

import (
	"fmt"
	"strings"

	"github.com/yugabyte/ybm-cli/cmd/util"
)

func regionInfoFlagDescription(includeBackupReplication bool) string {
	args := []string{
		"region=<region-name>",
		"num-nodes=<number-of-nodes>",
		"vpc=<vpc-name>",
		"num-cores=<num-cores>",
	}
	if util.IsFeatureFlagEnabled(util.MULTI_ZONE_SUPPORT) {
		args = append(args, "num-zones=<num-zones> (Multi-region SYNCHRONOUS with ZONE fault tolerance only)")
	}
	args = append(args, "disk-size-gb=<disk-size-gb>", "disk-iops=<disk-iops> (AWS only)")
	if includeBackupReplication && util.IsFeatureFlagEnabled(util.BACKUP_REPLICATION_GCP_TARGET) {
		args = append(args, "backup-replication-gcp-target=<gcp-target>")
	}
	return fmt.Sprintf(
		"Region information for the cluster, provided as key-value pairs. Arguments are %s. region, num-nodes, num-cores, disk-size-gb are required. Specify one --region-info flag for each region in the cluster.",
		strings.Join(args, ","),
	)
}
