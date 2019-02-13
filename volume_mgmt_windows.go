// +build windows

package vml

import "fmt"

// WARNING : Product does not work on windows  - stub implementation only

// Windows specific implementation. We will stub out these functions for now so that we can build on
// windows.

// Mount method is used to attach the filesystem on the device mapper of a specific type at the mount path.
//
// Input Parameters:
//
// 	deviceMapperLocation – Absolute path of the dm-crypt volume.
//
// 	mountLocation – Mount point location where the device will be mounted
func Mount(deviceMapperLocation string, mountLocation string) error {

	return fmt.Errorf("Function not implemented on Windows")

}

// Unmount method is used to detach the filesystem from the mount path.
//
// Input Parameter:
//
// mountLocation – Mount point location  where we want to unmount the device.
func Unmount(mountLocation string) error {

	return fmt.Errorf("Function not implemented on Windows")

}
