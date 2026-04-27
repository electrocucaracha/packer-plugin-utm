package iso

import (
	"context"
	"fmt"
	"strconv"

	utmcommon "github.com/electrocucaracha/packer-plugin-utm/builder/utm/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(utmcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmId := state.Get("vmId").(string)

	// The main disk and additional disks
	// We do not give names to the disks, as UTM does not support it
	diskSizes := []uint{config.DiskSize}
	if len(config.AdditionalDiskSize) > 0 {
		diskSizes = append(diskSizes, config.AdditionalDiskSize...)
	}

	// Create all required disks
	for i := range diskSizes {
		ui.Say(fmt.Sprintf("Creating hard drive with size %d MiB...", diskSizes[i]))

		// Convert controllerName to the corresponding enum code
		controllerEnumCode, err := utmcommon.GetControllerEnumCode(config.HardDriveInterface)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		command := []string{
			"add_drive.applescript", vmId,
			"--interface", controllerEnumCode,
			"--size", strconv.FormatUint(uint64(diskSizes[i]), 10),
		}
		_, err = driver.ExecuteOsaScript(command...)
		if err != nil {
			err := fmt.Errorf("error creating hard drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// In UTM disk creation and attaching are done in the same step

	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
