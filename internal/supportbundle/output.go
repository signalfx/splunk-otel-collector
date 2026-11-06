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

package supportbundle

import (
	"errors"
	"fmt"
	"os"
)

func createOutputFile(dest string) (*os.File, error) {
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", dest, err)
	}
	return f, nil
}

// finishOutputFile closes f and removes dest if any write or close failed.
// Callers must close any wrapping writers before calling it.
func finishOutputFile(f *os.File, dest string, errs ...error) error {
	if err := errors.Join(append(errs, f.Close())...); err != nil {
		_ = os.Remove(dest)
		return fmt.Errorf("close %s: %w", dest, err)
	}
	return nil
}
