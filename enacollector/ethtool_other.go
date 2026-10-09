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
