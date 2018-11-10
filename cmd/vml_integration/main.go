package main

import (
	"os"
	"log"
	"io/ioutil"
	"strings"
	"lib-go-volume-management/pkg/vml"
)

func main() {

	if (len(os.Args[1:]) < 2) {
		log.Fatal("Error while executing integration test. Usage :  ./pgmname wnlMethoName inputFile")
	}

	var methodName string = os.Args[1]
	var inputFileName string = os.Args[2]

	_, err := os.Stat(inputFileName)
	if(os.IsNotExist(err)) {
		log.Fatal("File with input parameters does not exist in ", inputFileName)
	}

	switch methodName {

	case "IsVolumeCreated" :
		log.Printf("Volume creation method called")
		var input = readFile(inputFileName)
		if(len(input) < 3) {
			log.Fatal("IsVolumeCreated() requires sparseFilePath , deviceMapperLocation, keyFile , diskSize")
		}		
		
		if (vml.IsVolumeCreated(input[0], input[1], input[2], input[3])) {
			log.Printf("dm-crypt volume created successfully")
		} else {
			log.Printf("dm-crypt volume was not created successfully")
		}
	
	case "IsVolumeDeleted" :
		log.Printf("Volume deletion method called")
		var input = readFile(inputFileName)
		if(len(input) < 1) {
			log.Fatal("IsVolumeDeleted() requires deviceMapperLocation")
		}

		if (vml.IsVolumeDeleted(input[0])) {
			log.Printf("dm-crypt volume deleted successfully")
		} else {
			log.Printf("dm-crypt volume could not be deleted")
		}

	case "IsMount" :
		log.Printf("Mount method called")
		var input = readFile(inputFileName)
		if(len(input) < 2) {
			log.Fatal("IsMount() requires deviceMapper, mountLocation")
		}

		if (vml.IsMount(input[0], input[1])) {
			log.Printf("Device mounted successfully")
		} else {
			log.Printf("Device could not be mounted")
		}

	case "IsUnmount" :
		log.Printf("Unmount method called")
		var input = readFile(inputFileName)
		if(len(input) < 2) {
			log.Fatal("IsUnMount() requires unmount device location")
		}

		if (vml.IsMount(input[0]) {
			log.Printf("Device unmounted successfully")
		} else {
			log.Printf("Device could not be unmounted")
		}

	case "IsDecrypt" :
		log.Printf("Decrypt method called")
		var input = readFile(inputFileName)
		if(len(input) < 2) {
			log.Fatal("IsMount() requires deviceMapper, mountLocation")
		}

		if (vml.IsMount(input[0], input[1])) {
			log.Printf("Device mounted successfully")
		} else {
			log.Printf("Device could not be mounted")
		}

	case "IsManifestCreated" :
		log.Printf("Manifest creation method called")
		var input = readFile(inputFileName)
		if(len(input) < 2) {
			log.Fatal("IsMount() requires deviceMapper, mountLocation")
		}

		if (vml.IsMount(input[0], input[1])) {
			log.Printf("Device mounted successfully")
		} else {
			log.Printf("Device could not be mounted")
		}

	default :
		log.Printf("Invalid method name mentioned.\nExpected values: IsVolumeCreated, IsVolumeDeleted, IsMount, IsUnmount, IsManifestCreated, IsDecrypt")	
	}
}

func readFile(inputFileName string) []string {
	fileRead, err := ioutil.ReadFile(inputFileName)
    if err != nil {
        log.Fatal("Error trying to read the input file ", inputFileName)
	}

	var inputParmString = string(fileRead)
	var inputParms = strings.Split(inputParmString, "\n")
	
	return inputParms
}