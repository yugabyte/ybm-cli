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
package readreplica_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	readreplica "github.com/yugabyte/ybm-cli/cmd/cluster/read-replica"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

var _ = Describe("Read Replica utils", func() {
	Context("When Getting default spec", func() {
		It("should returh the correct default spec", func() {
			vpcId := "12345698"
			n := int32(1)
			numReplicas := ybmclient.NewNullableInt32(&n)
			correctSpec := ybmclient.ReadReplicaSpec{
				PlacementInfo: ybmclient.PlacementInfo{
					CloudInfo: ybmclient.CloudInfo{
						Code:   "AWS",
						Region: "us-west2",
					},
					VpcId:       *ybmclient.NewNullableString(&vpcId),
					NumNodes:    1,
					NumReplicas: *numReplicas,
				},
			}
			nodeInfo := ybmclient.ClusterNodeInfo{
				NumCores: 2,
			}
			correctSpec.SetNodeInfo(nodeInfo)
			spec := readreplica.GetDefaultSpec(ybmclient.CLOUDENUM_AWS, vpcId)
			Expect(spec).To(BeEquivalentTo(correctSpec))
		})
	})
})
