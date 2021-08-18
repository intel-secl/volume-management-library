# volume-management library (VML)
The Volume Management Library is used to perform some of the tasks during the VM launch as a part of VM integrity and confidentiality use case. Some of the tasks performed by VML are create the image and VM volumes, mount the image, decrypt the image, unmount an image, delete the dm-crypt volumes created and creation of VM manifest. 

pkg - contains the library source code
cmd - contains the main method which calls into the library

## Key features
- Create dm-crypt volume
- Delete dm-crypt volume
- Mount a device
- Unmount a device
- Decrypt a file
- Create VM manifest
- Create container manifest


## System Requirements
- RHEL 8.1
- Epel 8 Repo
- Proxy settings if applicable

## Software requirements
- git
- `go` version 1.16.7

# Step By Step Build Instructions

### Install `go` version 1.16.7
The `VML` requires Go version 1.16.7 that has support for `go modules`. The build was validated with the latest version 1.16.7 of `go`. It is recommended that you use 1.16.7 version of `go`. More recent versions may introduce compatibility issues. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.16.7.linux-amd64.tar.gz
tar -xzf go1.16.7.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

# Third Party Dependencies

## VML

### Direct dependencies

| Name                  | Repo URL           | Minimum Version Required           |
| ----------------------| -------------------| :--------------------------------: |
| system commands       | golang.org/x/sys   | v0.0.0-20181107165924-66b7b1311ac8 |


*Note: All dependencies are listed in go.mod*

# Links
https://01.org/intel-secl/
