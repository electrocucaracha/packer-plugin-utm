// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	"github.com/electrocucaracha/packer-plugin-utm/builder/utm/cloud"
	"github.com/electrocucaracha/packer-plugin-utm/builder/utm/iso"
	"github.com/electrocucaracha/packer-plugin-utm/builder/utm/utm"
	utmPPvagrant "github.com/electrocucaracha/packer-plugin-utm/post-processor/vagrant"
	utmPPzip "github.com/electrocucaracha/packer-plugin-utm/post-processor/zip"
	"github.com/electrocucaracha/packer-plugin-utm/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("iso", new(iso.Builder))
	pps.RegisterBuilder("utm", new(utm.Builder))
	pps.RegisterBuilder("cloud", new(cloud.Builder))
	pps.RegisterPostProcessor("zip", new(utmPPzip.PostProcessor))
	pps.RegisterPostProcessor("vagrant", new(utmPPvagrant.PostProcessor))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
