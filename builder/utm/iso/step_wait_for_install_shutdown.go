package iso

import (
	"context"
	"errors"
	"log"
	"time"

	utmcommon "github.com/electrocucaracha/packer-plugin-utm/builder/utm/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// stepWaitForInstallShutdown waits for the installer to power off the VM after
// the unattended install completes.
type stepWaitForInstallShutdown struct {
	Timeout time.Duration
}

func (s *stepWaitForInstallShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(utmcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmId := state.Get("vmId").(string)

	ui.Say("Waiting for install to complete and power off...")
	log.Printf("Waiting max %s for installer shutdown", s.Timeout)

	shutdownTimer := time.After(s.Timeout)
	for {
		running, _ := driver.IsRunning(vmId)
		if !running {
			log.Println("Installer shut down the VM.")
			return multistep.ActionContinue
		}

		select {
		case <-shutdownTimer:
			err := errors.New("timeout while waiting for installer shutdown")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (s *stepWaitForInstallShutdown) Cleanup(state multistep.StateBag) {}
