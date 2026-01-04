/*
 * Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 		http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	"fmt"
)

// Validate ensures that NpuSharingStrategy has a valid set of values.
func (s NpuSharingStrategy) Validate() error {
	switch s {
	case TimeSlicingStrategy, SpacePartitioningStrategy:
		return nil
	}
	return fmt.Errorf("unknown GPU sharing strategy: %v", s)
}

// Validate ensures that TimeSliceInterval has a valid set of values.
func (d TimeSliceInterval) Validate() error {
	switch d {
	case DefaultTimeSlice, ShortTimeSlice, MediumTimeSlice, LongTimeSlice:
		return nil
	}
	return fmt.Errorf("unknown time-slice interval: %v", d)
}

// Validate ensures that TimeSlicingConfig has a valid set of values.
func (c *TimeSlicingConfig) Validate() error {
	return c.Interval.Validate()
}

// Validate ensures that SpacePartitioningConfig has a valid set of values.
func (c *SpacePartitioningConfig) Validate() error {
	if c.PartitionCount < 0 {
		return fmt.Errorf("invalid partition count: %v", c.PartitionCount)
	}
	return nil
}

// Validate ensures that NpuSharing has a valid set of values.
func (s *NpuSharing) Validate() error {
	if err := s.Strategy.Validate(); err != nil {
		return err
	}
	switch {
	case s.IsTimeSlicing():
		return s.TimeSlicingConfig.Validate()
	case s.IsSpacePartitioning():
		return s.SpacePartitioningConfig.Validate()
	}
	return fmt.Errorf("invalid GPU sharing settings: %v", s)
}

// Validate ensures that NpuConfig has a valid set of values.
func (c *NpuConfig) Validate() error {
	if c.Sharing == nil {
		return fmt.Errorf("no sharing strategy set")
	}
	return c.Sharing.Validate()
}
