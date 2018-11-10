package vml

import(

	"os/exec"
	"os"
	"log"
	"strings"
	"syscall"
)


const (
	truncate = "truncate"
	cryptsetup = "cryptsetup"
	losetup = "losetup"
	decrypt = "decrypt"
)

type runCmd struct {
	output string
	err error
}

func IsVolumeCreated(sparseFilePath string, deviceMapperLocation string, keyFile string, diskSize string) bool {

	// input validation
	if (sparseFilePath == "") || (len(sparseFilePath) <= 0) ||
		(deviceMapperLocation == "") || (len(deviceMapperLocation) <= 0) ||
        (keyFile == "") || (len(keyFile) <=0) {
			log.Fatal("Invalid input parameters")
			return false
	}

	// validate the disksize in KB
	// check if the key file exists in the location
	

	var formatDevice bool = false
	var cmdOutput runCmd
	var args []string
	var deviceLoop string

	// 1. check if the sparse file exists
	log.Println("Checking if the sparse file exists in :", sparseFilePath)
	_, err := os.Stat(sparseFilePath)
	// sprase file does not exist, creating a new sparsefile
	if(os.IsNotExist(err)) {
		log.Println("Sparse file does not exist, creating a new file")
		//2. Create a sparse file
		args = []string{"-s", diskSize, sparseFilePath}
		cmdOutput = runCommand(truncate, args)
		if cmdOutput.err != nil {
			log.Fatal("Error creating a sparse file")
			return false
		}
		formatDevice = true
	}
	log.Println("Sparse file exists in location", sparseFilePath)

	//3. find the loop device associated to the sparse file
	log.Println("Finding a loop device that is associated to the sparse file")
	args = []string{"-j", sparseFilePath}
	cmdOutput = runCommand(losetup, args)
		if cmdOutput.err != nil {
			log.Fatal("Error trying to find a loop device associated with the sparse file")
			return false
		}
		// find the loop device and associate it with the sparse file
		if (cmdOutput.output == "") || (len(cmdOutput.output) <= 0) {
			log.Println("No loop device found that is assciated to the sparse file")
			log.Println("finding a loop device and associating it to the sparse file")
			//4. find the loop device
			args = []string{"-f", sparseFilePath}
			cmdOutput = runCommand(losetup, args)
			if cmdOutput.err != nil {
				log.Fatal("Error trying to attach the sparse file to the loop device")
				return false
			}
		}

	//check if the loop device is associated to the sparse file
	args = []string{"-j", sparseFilePath}
	cmdOutput = runCommand(losetup, args)
		if (cmdOutput.output == "") || (len(cmdOutput.output) <= 0) {
			log.Fatal("Sparse file is not associated to the loop device correctly")
			return false
		} else {
			var modifiedOutput = strings.Split(cmdOutput.output, ":")
			deviceLoop = modifiedOutput[0]
			log.Println("The sparse file is associated to the loop device : ", deviceLoop)
		}

	// 6. format loop device
	log.Println("Formatting the loop device : ", deviceLoop)
	if (formatDevice) {
		log.Println("Format device value : ", formatDevice)
		args = []string{"-v", "--batch-mode", "luksFormat", deviceLoop, "-d", keyFile}
		cmdOutput = runCommand(cryptsetup, args)
		if cmdOutput.err != nil {
			log.Fatal("Error trying to format the loop device")
			return false
		}
	}

	var deviceMapperString = strings.Split(deviceMapperLocation, "/")
    var deviceMapperName string = deviceMapperString[len(deviceMapperString)-1]
	

	// 7. check the status of the device device mapper
	log.Println("Checking the status of the device mapper ", deviceMapperLocation)
	log.Println("Checking the status of the device mapper name", deviceMapperName)
	args = []string{"status", deviceMapperLocation}
	cmdOutput = runCommand(cryptsetup, args)
		if (strings.Contains(cmdOutput.output, "inactive")) {
			log.Println("The device mapper is inactive, opening the luks volume")
			// 8. open the luks volume
			args = []string{"-v", "luksOpen", deviceLoop, deviceMapperName, "-d", keyFile}
			cmdOutput = runCommand(cryptsetup, args)
			if cmdOutput.err != nil {
				log.Fatal("Error trying to open the luks volume")
				return false
			} else {
				formatDevice = true
			}

			//checking the status of the volume again
			log.Println("Checking the status of the device mapper ", deviceMapperLocation)
			args = []string{"status", deviceMapperLocation}
			cmdOutput = runCommand(cryptsetup, args)
			if (!strings.Contains(cmdOutput.output, "active")) {
				log.Fatal("Volume is not active for use")
				return false
			}
		}
	// 9. format the volume
	log.Println("Formatting the dm-crypt volume")
	if (formatDevice) {
		log.Println("The format device value is : ", formatDevice)
		args = []string{"-v", deviceMapperLocation}
		cmdOutput = runCommand("mkfs.ext4", args)
		if cmdOutput.err != nil {
			log.Fatal("Error trying to format the luks volume")
			return false
		}
	}

	return true
}


func IsVolumeDeleted(deviceMapperLocation string) bool {
	if (deviceMapperLocation == "") || (len(deviceMapperLocation) <= 0) {
		return false
	}
	log.Println("Deleteing teh dm-crypt volume ", deviceMapperLocation)
	deleteVolumeCmd := "cryptsetup"
	args := []string{"luksClose", deviceMapperLocation}
	var cmdOutput runCmd = runCommand(deleteVolumeCmd, args)
	if cmdOutput.err != nil {
		log.Fatal("Error trying to close the device mapper")
		return false
	}	
	return true
}

func IsMount(deviceMapper string, mountLocation string) bool{
	if (deviceMapper == "") || (len(deviceMapper) <= 0) ||
	(mountLocation == "") || (len(mountLocation) <= 0) {
		log.Fatal("Invalid input parameters")
			return false
	}
	err := syscall.Mount(deviceMapper, mountLocation, "ext4", 0, "")
	if  err != nil {
		log.Fatal("Error while trying to mount %v ", err)
		return false
	}
 return true
}

func IsUnmount(mountLocation string) bool{
	if (mountLocation == "") || (len(mountLocation) <= 0) {
		log.Fatal("Invalid input parameters")
		return false
	}
	err := syscall.Unmount(mountLocation, 0)
	if  err != nil {
		log.Fatal("Error while trying to unmount %v ", err)
		return false
	}
	
	return true
}

func runCommand(cmd string, args []string) runCmd {

	out, err := exec.Command(cmd, args...).Output()
	cmdOutput := runCmd{output: string(out), err : err}
	return cmdOutput
}