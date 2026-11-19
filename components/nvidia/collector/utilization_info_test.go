/*
Copyright 2024 The Scitix Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package collector

import (
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/stretchr/testify/assert"
)

// fakeUtilDevice overrides only GetUtilizationRates; any other call panics
// on the nil embedded interface, which keeps the test honest.
type fakeUtilDevice struct {
	nvml.Device
	util nvml.Utilization
	ret  nvml.Return
}

func (d fakeUtilDevice) GetUtilizationRates() (nvml.Utilization, nvml.Return) {
	return d.util, d.ret
}

func TestUtilizationInfo_GetNotSupported(t *testing.T) {
	tests := []struct {
		name    string
		dev     fakeUtilDevice
		want    UtilizationInfo
		wantErr bool
	}{
		{
			name: "success",
			dev:  fakeUtilDevice{util: nvml.Utilization{Gpu: 42, Memory: 7}, ret: nvml.SUCCESS},
			want: UtilizationInfo{GPUUsagePercent: 42, MemoryUsagePercent: 7},
		},
		{
			// MIG-enabled parent GPU: must not be treated as a lost GPU.
			name: "mig not supported",
			dev:  fakeUtilDevice{ret: nvml.ERROR_NOT_SUPPORTED},
			want: UtilizationInfo{},
		},
		{
			name:    "gpu lost",
			dev:     fakeUtilDevice{ret: nvml.ERROR_GPU_IS_LOST},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got UtilizationInfo
			err := got.Get(tt.dev, "fake-uuid")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
