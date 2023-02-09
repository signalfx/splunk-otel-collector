// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

// IsSignalChanClosed returns whether a channel is closed that is used only for
// the sake of sending a single singal.  The channel should never be sent any
// actual values, but should only be closed to tell other goroutines to stop.
func IsSignalChanClosed(ch <-chan struct{}) bool {
	if ch == nil {
		return true
	}

	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}
