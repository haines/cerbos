// Copyright 2021-2022 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/alecthomas/kong"

	"github.com/cerbos/cerbos/cmd/cerbos/compile"
	"github.com/cerbos/cerbos/internal/util"
)

func main() {
	var cli struct {
		Compile compile.Cmd `cmd:"" help:"Compile and test policies"`
	}

	ctx := kong.Parse(&cli,
		kong.Name(util.AppName),
		kong.Description("Painless access controls for cloud-native applications"),
		kong.UsageOnError(),
	)

	ctx.FatalIfErrorf(ctx.Run())
}
