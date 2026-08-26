/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package device

// DPU represents a single DPU card and its interfaces.
// Per the class diagram, DPU carries its own metrics map and accessor methods.
type DPU struct {
	// CardName e.g. "hinic0"
	CardName string
	// CardType e.g. "CAL_2X400G_UB_EXP"
	CardType string
	// HcaName HCA name associated with this card
	HcaName string
	// Interfaces list of network interfaces under this DPU
	Interfaces []Interface
	// metrics holds global metrics for this DPU card (from hinicadm5)
	metrics map[string]float64
}

// NewDPU creates a DPU with initialized metrics map
func NewDPU(cardName, cardType, hcaName string) DPU {
	return DPU{
		CardName: cardName,
		CardType: cardType,
		HcaName:  hcaName,
		metrics:  make(map[string]float64),
	}
}

// GetMetric returns the metric value for the given key, or 0 if not found
func (d *DPU) GetMetric(key string) float64 {
	return d.metrics[key]
}

// SetMetrics replaces all metrics for this DPU
func (d *DPU) SetMetrics(metrics map[string]float64) {
	d.metrics = metrics
}

// GetMetrics returns a copy of all metrics
func (d *DPU) GetMetrics() map[string]float64 {
	cp := make(map[string]float64, len(d.metrics))
	for k, v := range d.metrics {
		cp[k] = v
	}
	return cp
}

// GetEthList returns the list of ethernet interface names under this DPU
func (d *DPU) GetEthList() []string {
	ethList := make([]string, 0, len(d.Interfaces))
	for i := range d.Interfaces {
		ethList = append(ethList, d.Interfaces[i].EthName)
	}
	return ethList
}
