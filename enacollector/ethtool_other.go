//go:build !linux

package enacollector

// noopReader finds no interfaces. ENA interfaces only exist on Linux.
type noopReader struct{}

func newEthtoolReader() (statsReader, error) {
	return noopReader{}, nil
}

func (noopReader) Interfaces() ([]string, error)           { return nil, nil }
func (noopReader) DriverName(string) (string, error)       { return "", nil }
func (noopReader) Stats(string) (map[string]uint64, error) { return nil, nil }
func (noopReader) Close()                                  {}
