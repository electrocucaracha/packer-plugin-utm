package common

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepAttachFloppy attaches the floppy disk image created by StepCreateFloppy to the VM.
// The floppy path is read from state key "floppy_path" (set by commonsteps.StepCreateFloppy).
// If no floppy was created, this step is a no-op.
type StepAttachFloppy struct {
	floppyDriveUUID string
}

func (s *StepAttachFloppy) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	driver := state.Get("driver").(Driver)
	vmId := state.Get("vmId").(string)

	// StepCreateFloppy stores the image path here; absent means no floppy files were specified.
	floppyPathRaw, ok := state.GetOk("floppy_path")
	if !ok {
		log.Println("No floppy_path in state; skipping floppy attachment")
		return multistep.ActionContinue
	}

	floppyPath := floppyPathRaw.(string)

	// UTM requires a real absolute POSIX path for AppleScript sandbox file access.
	// filepath.Abs handles relative paths; filepath.EvalSymlinks resolves /tmp -> /private/tmp
	// on macOS (StepCreateFloppy writes to /tmp which is a symlink on macOS).
	if !filepath.IsAbs(floppyPath) {
		absPath, err := filepath.Abs(floppyPath)
		if err != nil {
			err := fmt.Errorf("error resolving absolute floppy path: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		floppyPath = absPath
	}
	resolvedPath, err := filepath.EvalSymlinks(floppyPath)
	if err != nil {
		err := fmt.Errorf("error resolving symlink for floppy path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	floppyPath = resolvedPath

	ui.Say(fmt.Sprintf("Attaching floppy image: %s", floppyPath))

	output, err := driver.ExecuteOsaScript(
		"add_floppy.applescript", vmId,
		"--source", floppyPath,
	)
	if err != nil {
		err := fmt.Errorf("error attaching floppy: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Track the drive UUID for cleanup (remove_drive.applescript needs it).
	re := regexp.MustCompile(`[0-9a-fA-F-]{36}`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 0 {
		s.floppyDriveUUID = matches[0]
		ui.Message(fmt.Sprintf("Floppy attached (drive UUID: %s)", s.floppyDriveUUID))
	}

	return multistep.ActionContinue
}

func (s *StepAttachFloppy) Cleanup(state multistep.StateBag) {
	if s.floppyDriveUUID == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmId := state.Get("vmId").(string)

	ui.Say("Detaching floppy image...")
	if _, err := driver.ExecuteOsaScript("remove_drive.applescript", vmId, s.floppyDriveUUID); err != nil {
		log.Printf("error detaching floppy (UUID: %s): %s", s.floppyDriveUUID, err)
	}
}
