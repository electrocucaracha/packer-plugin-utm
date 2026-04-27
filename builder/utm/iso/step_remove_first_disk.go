package iso

import (
	"context"
	"fmt"

	utmcommon "github.com/electrocucaracha/packer-plugin-utm/builder/utm/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step removes the first virtual disk from the virtual machine.
type stepRemoveFirstDisk struct{}

func (s *stepRemoveFirstDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(utmcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmId := state.Get("vmId").(string)

	ui.Say("Removing first drive")

	command := []string{
		"remove_first_drive.applescript", vmId,
	}
	_, err := driver.ExecuteOsaScript(command...)
	if err != nil {
		err := fmt.Errorf("error removing first drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepRemoveFirstDisk) Cleanup(state multistep.StateBag) {}
