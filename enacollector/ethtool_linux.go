// Copyright ProbeLab
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package enacollector

import (
	"net"

	"github.com/safchain/ethtool"
)

// ethtoolReader reads interface statistics with the ethtool ioctl. The ioctl
// calls that it uses do not need the CAP_NET_ADMIN capability.
type ethtoolReader struct {
	*ethtool.Ethtool
}

func newEthtoolReader() (statsReader, error) {
	e, err := ethtool.NewEthtool()
	if err != nil {
		return nil, err
	}
	return ethtoolReader{e}, nil
}

func (ethtoolReader) Interfaces() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		names = append(names, iface.Name)
	}
	return names, nil
}
