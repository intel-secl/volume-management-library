// +build linux

package vml

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/sys/unix"
)

// Mount method is used to attach the filesystem on the device mapper of a specific type at the mount path.
//
// Input Parameters:
//
// 	deviceMapperLocation – Absolute path of the dm-crypt volume.
//
// 	mountLocation – Mount point location where the device will be mounted
func Mount(deviceMapperLocation string, mountLocation string) error {
	//input parameters validation
	if len(strings.TrimSpace(deviceMapperLocation)) <= 0 {
		return fmt.Errorf("device mapper location not given")
	}
	if len(strings.TrimSpace(mountLocation)) <= 0 {
		return fmt.Errorf("mount location not given")
	}
	// call syscall to mount the file system
	err := unix.Mount(deviceMapperLocation, mountLocation, "ext4", 0, "")
	if err != nil {
		log.Println("Error: ", err)
		if strings.Contains(string(err.Error()), "device or resource busy") {
			return fmt.Errorf("device is already mounted")
		} else {
			return fmt.Errorf("error while trying to mount")
		}
	}
	return nil
}

// Unmount method is used to detach the filesystem from the mount path.
//
// Input Parameter:
//
// mountLocation – Mount point location  where we want to unmount the device.
func Unmount(mountLocation string) error {
	//input parameters validation
	if len(strings.TrimSpace(mountLocation)) <= 0 {
		return fmt.Errorf("unmount location not given")
	}

	// call syscall to unmount the file system from the mount location
	err := unix.Unmount(mountLocation, 0)
	if err != nil {
		log.Println("Error: ", err)
		return fmt.Errorf("error while trying to unmount - function returned error : %s", err.Error())
	}
	return nil
}
