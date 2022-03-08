// Copyright 2021-2022 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package tracing

import (
	"context"
	"fmt"
)

func configureJaeger(ctx context.Context) error {
	return fmt.Errorf("Jaeger tracing is not supported in the WebAssembly build")
}
