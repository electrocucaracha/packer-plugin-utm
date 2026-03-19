---
-- add_floppy.applescript
-- Attaches a floppy disk image (.img) to a specified UTM virtual machine.
-- Usage: osascript add_floppy.applescript <VM_UUID> --source <FLOPPY_IMG_PATH>
-- Example: osascript add_floppy.applescript A1B2C3 --source "/tmp/floppy.img"
-- Returns the UUID of the new drive for tracking.

on run argv
  set vmId to item 1 of argv
  -- Parse the --source argument
  set floppyPath to item 3 of argv as string

  -- Gain sandbox access to the file for UTM (sandboxed app)
  set floppyFile to POSIX file floppyPath

  tell application "UTM"
    -- Get the VM and its configuration
    set vm to virtual machine id vmId
    set config to configuration of vm

    -- Existing drives
    set vmDrives to drives of config
    -- Create new floppy drive: interface "QdIf" is UTM's enum for floppy
    set newDrive to {removable: true, interface: "QdIf", source: floppyFile}
    -- Add to the end of the drive list
    copy newDrive to end of vmDrives
    -- Update drive list
    set drives of config to vmDrives

    -- Save the configuration (VM must be stopped)
    update configuration of vm with config

    -- Return the new drive UUID for cleanup tracking
    set updatedConfig to configuration of vm
    set updatedDrives to drives of updatedConfig
    set updatedDrive to item -1 of updatedDrives
    return id of updatedDrive
  end tell
end run
