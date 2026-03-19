package common

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepAttachFloppy_noFloppyInState(t *testing.T) {
	state := testState(t)
	state.Put("vmId", "test-vm-id")
	// No "floppy_path" in state — step should skip gracefully

	step := &StepAttachFloppy{}
	action := step.Run(context.Background(), state)

	if action != multistep.ActionContinue {
		t.Fatalf("expected ActionContinue when no floppy_path, got %v", action)
	}

	driver := state.Get("driver").(*DriverMock)
	if len(driver.ExecuteOsaCalls) != 0 {
		t.Fatalf("expected 0 osascript calls when no floppy, got %d", len(driver.ExecuteOsaCalls))
	}
}

func TestStepAttachFloppy_attachesFloppyFromState(t *testing.T) {
	// filepath.EvalSymlinks requires the file to actually exist.
	// Create a real temp file; the mock driver handles the osascript call.
	f, err := os.CreateTemp("", "test-floppy-*.img")
	if err != nil {
		t.Fatalf("failed to create temp floppy file: %s", err)
	}
	defer os.Remove(f.Name())
	f.Close()

	state := testState(t)
	state.Put("vmId", "AAAABBBB-CCCC-DDDD-EEEE-FFFFAAAABBBB")
	state.Put("floppy_path", f.Name())

	driver := state.Get("driver").(*DriverMock)
	// Return a UUID from the applescript (mimics add_floppy.applescript output)
	driver.ExecuteOsaResult = "11111111-2222-3333-4444-555555555555"

	step := &StepAttachFloppy{}
	action := step.Run(context.Background(), state)

	if action != multistep.ActionContinue {
		t.Fatalf("expected ActionContinue, got %v", action)
	}

	if len(driver.ExecuteOsaCalls) != 1 {
		t.Fatalf("expected 1 osascript call, got %d", len(driver.ExecuteOsaCalls))
	}

	call := driver.ExecuteOsaCalls[0]
	if call[0] != "add_floppy.applescript" {
		t.Fatalf("expected add_floppy.applescript, got %q", call[0])
	}
	if call[1] != "AAAABBBB-CCCC-DDDD-EEEE-FFFFAAAABBBB" {
		t.Fatalf("expected vmId as second arg, got %q", call[1])
	}
	if call[2] != "--source" {
		t.Fatalf("expected --source flag, got %q", call[2])
	}
	// Path must be absolute
	if len(call[3]) == 0 || call[3][0] != '/' {
		t.Fatalf("expected absolute path, got %q", call[3])
	}
	if step.floppyDriveUUID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("expected UUID to be tracked, got %q", step.floppyDriveUUID)
	}
}

func TestStepAttachFloppy_cleanupDetachesDrive(t *testing.T) {
	state := testState(t)
	state.Put("vmId", "AAAABBBB-CCCC-DDDD-EEEE-FFFFAAAABBBB")

	step := &StepAttachFloppy{
		floppyDriveUUID: "11111111-2222-3333-4444-555555555555",
	}
	step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)
	if len(driver.ExecuteOsaCalls) != 1 {
		t.Fatalf("expected 1 osascript call during cleanup, got %d", len(driver.ExecuteOsaCalls))
	}
	if driver.ExecuteOsaCalls[0][0] != "remove_drive.applescript" {
		t.Fatalf("expected remove_drive.applescript, got %q", driver.ExecuteOsaCalls[0][0])
	}
}
