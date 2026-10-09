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
