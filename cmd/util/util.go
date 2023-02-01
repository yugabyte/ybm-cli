package util

import (
	"errors"

	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func FindNetworkAllowList(nals []ybmclient.NetworkAllowListData, name string) (ybmclient.NetworkAllowListData, error) {
	for _, allowList := range nals {
		if allowList.Spec.Name == name {
			return allowList, nil
		}
	}
	return ybmclient.NetworkAllowListData{}, errors.New("Unable to find NetworkAllowList " + name)
}
