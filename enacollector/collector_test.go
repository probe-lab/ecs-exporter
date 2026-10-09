package enacollector

import (
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeIface struct {
	driver    string
	driverErr error
	stats     map[string]uint64
	statsErr  error
}

type fakeReader struct {
	ifaces    map[string]fakeIface
	order     []string
	ifacesErr error
	closed    bool
}

func (r *fakeReader) Interfaces() ([]string, error) { return r.order, r.ifacesErr }

func (r *fakeReader) DriverName(iface string) (string, error) {
	return r.ifaces[iface].driver, r.ifaces[iface].driverErr
}

func (r *fakeReader) Stats(iface string) (map[string]uint64, error) {
	return r.ifaces[iface].stats, r.ifaces[iface].statsErr
}

func (r *fakeReader) Close() { r.closed = true }

func newTestCollector(reader *fakeReader, readerErr error) *collector {
	return &collector{
		newReader: func() (statsReader, error) {
			if readerErr != nil {
				return nil, readerErr
			}
			return reader, nil
		},
		logger: slog.New(slog.DiscardHandler),
	}
}

// enaStats is a subset of the statistics that the ENA driver reports.
var enaStats = map[string]uint64{
	"tx_timeout":                    0,
	"queue_0_tx_cnt":                12345,
	"bw_in_allowance_exceeded":      1,
	"bw_out_allowance_exceeded":     2,
	"pps_allowance_exceeded":        3,
	"conntrack_allowance_exceeded":  4,
	"linklocal_allowance_exceeded":  5,
	"conntrack_allowance_available": 136812,
}

const allMetricsEth1 = `
# HELP ecs_network_bw_in_allowance_exceeded_total Number of packets queued or dropped because the inbound aggregate bandwidth exceeded the maximum for the instance.
# TYPE ecs_network_bw_in_allowance_exceeded_total counter
ecs_network_bw_in_allowance_exceeded_total{interface="eth1"} 1
# HELP ecs_network_bw_out_allowance_exceeded_total Number of packets queued or dropped because the outbound aggregate bandwidth exceeded the maximum for the instance.
# TYPE ecs_network_bw_out_allowance_exceeded_total counter
ecs_network_bw_out_allowance_exceeded_total{interface="eth1"} 2
# HELP ecs_network_conntrack_allowance_available Number of tracked connections that the instance can establish before it reaches the connections tracked allowance of its instance type. Only available on Nitro-based instances.
# TYPE ecs_network_conntrack_allowance_available gauge
ecs_network_conntrack_allowance_available{interface="eth1"} 136812
# HELP ecs_network_conntrack_allowance_exceeded_total Number of packets dropped because connection tracking exceeded the maximum for the instance and new connections could not be established.
# TYPE ecs_network_conntrack_allowance_exceeded_total counter
ecs_network_conntrack_allowance_exceeded_total{interface="eth1"} 4
# HELP ecs_network_linklocal_allowance_exceeded_total Number of packets dropped because the packets per second to local proxy services (Amazon DNS, instance metadata, Amazon Time Sync) exceeded the maximum for the network interface.
# TYPE ecs_network_linklocal_allowance_exceeded_total counter
ecs_network_linklocal_allowance_exceeded_total{interface="eth1"} 5
# HELP ecs_network_pps_allowance_exceeded_total Number of packets queued or dropped because the bidirectional packets per second exceeded the maximum for the instance.
# TYPE ecs_network_pps_allowance_exceeded_total counter
ecs_network_pps_allowance_exceeded_total{interface="eth1"} 3
`

func TestCollect(t *testing.T) {
	tests := []struct {
		name      string
		reader    *fakeReader
		readerErr error
		want      string
	}{
		{
			name: "ena interface",
			reader: &fakeReader{
				order:  []string{"eth1"},
				ifaces: map[string]fakeIface{"eth1": {driver: "ena", stats: enaStats}},
			},
			want: allMetricsEth1,
		},
		{
			name: "skips interfaces that do not use the ena driver",
			reader: &fakeReader{
				order: []string{"eth0", "eth1", "ecs-eth0"},
				ifaces: map[string]fakeIface{
					"eth0":     {driver: "veth", stats: enaStats},
					"eth1":     {driver: "ena", stats: enaStats},
					"ecs-eth0": {driverErr: errors.New("operation not supported")},
				},
			},
			want: allMetricsEth1,
		},
		{
			name: "skips statistics that the driver does not report",
			reader: &fakeReader{
				order: []string{"eth1"},
				ifaces: map[string]fakeIface{"eth1": {driver: "ena", stats: map[string]uint64{
					"bw_in_allowance_exceeded": 7,
				}}},
			},
			want: `
# HELP ecs_network_bw_in_allowance_exceeded_total Number of packets queued or dropped because the inbound aggregate bandwidth exceeded the maximum for the instance.
# TYPE ecs_network_bw_in_allowance_exceeded_total counter
ecs_network_bw_in_allowance_exceeded_total{interface="eth1"} 7
`,
		},
		{
			name: "keeps other interfaces when stats fail",
			reader: &fakeReader{
				order: []string{"eth0", "eth1"},
				ifaces: map[string]fakeIface{
					"eth0": {driver: "ena", statsErr: errors.New("no such device")},
					"eth1": {driver: "ena", stats: enaStats},
				},
			},
			want: allMetricsEth1,
		},
		{
			name:   "no interfaces",
			reader: &fakeReader{},
		},
		{
			name:   "interface list fails",
			reader: &fakeReader{ifacesErr: errors.New("boom")},
		},
		{
			name:      "ethtool fails",
			readerErr: errors.New("socket: permission denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestCollector(tt.reader, tt.readerErr)
			if err := testutil.CollectAndCompare(c, strings.NewReader(tt.want)); err != nil {
				t.Fatal(err)
			}
			if tt.reader != nil && tt.readerErr == nil && !tt.reader.closed {
				t.Error("reader was not closed")
			}
		})
	}
}

func TestCollectorLint(t *testing.T) {
	c := newTestCollector(&fakeReader{
		order:  []string{"eth1"},
		ifaces: map[string]fakeIface{"eth1": {driver: "ena", stats: enaStats}},
	}, nil)
	problems, err := testutil.CollectAndLint(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range problems {
		t.Errorf("%s: %s", p.Metric, p.Text)
	}
}

// TestNewCollector checks that the real collector does not panic on the host
// that runs the tests. Hosts outside of AWS have no ENA interfaces, so the
// metric count is not checked.
func TestNewCollector(t *testing.T) {
	c := NewCollector(slog.New(slog.DiscardHandler))
	t.Logf("collected %d metrics", testutil.CollectAndCount(c))
}
