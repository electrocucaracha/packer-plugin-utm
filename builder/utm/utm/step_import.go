// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utm

import (
	"context"
	"fmt"

	utmcommon "github.com/electrocucaracha/packer-plugin-utm/builder/utm/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step imports an UTM VM into UTM.
type StepImport struct {
	Name           string
	ImportFlags    []string
	KeepRegistered bool

	vmName string
	vmId   string
}

func (s *StepImport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(utmcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmPath := state.Get("vm_path").(string)

	var vmId string
	var err error

	ui.Say(fmt.Sprintf("Importing VM: %s", vmPath))
	if vmId, err = driver.Import(vmPath); err != nil {
		err := fmt.Errorf("error importing VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.vmId = vmId
	state.Put("vmId", s.vmId)

	// set VM name
	if _, err = driver.ExecuteOsaScript("customize_vm.applescript", vmId, "--name", s.Name); err != nil {
		err := fmt.Errorf("error setting VM name: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.vmName = s.Name
	state.Put("vmName", s.Name)

	return multistep.ActionContinue
}

func (s *StepImport) Cleanup(state multistep.StateBag) {
	if s.vmId == "" {
		return
	}

	driver := state.Get("driver").(utmcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if (s.KeepRegistered) && (!cancelled && !halted) {
		ui.Say("Keeping virtual machine registered with UTM host (keep_registered = true)")
		return
	}

	ui.Say("Deregistering and deleting imported VM...")
	if err := driver.Delete(s.vmId); err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM: %s", err))
	}
}
