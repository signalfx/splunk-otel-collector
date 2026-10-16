// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !linux

package supportbundle

func writeOrWarn(logger *logger, dest, content string) {
	f, err := createOutputFile(dest)
	if err != nil {
		logger.Warnf("failed to create %s: %v", dest, err)
		return
	}

	var writeErr error
	defer func() {
		if err := finishOutputFile(f, dest, writeErr); err != nil {
			logger.Warnf("failed to write %s: %v", dest, err)
		}
	}()

	_, writeErr = f.WriteString(content)
}
