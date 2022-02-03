# FPC Vagrant Dev Environment

We provide an alternative development environment for FPC using Vagrant. 

## Setup

On a Mac you can install virtualbox and vagrant using brew.

```bash
brew install --cask virtualbox
brew install vagrant
```

## Starting the VM

We provide the scripts to create and provision a virtual machine with all tools needed to build and test FPC.
Once you cloned the FPC repository, just run `vagrant up` in this directory to create and provision the dev VM.
With `vagrant ssh` you can access the development shell inside the VM.

```bash
cd $FPC_PATH/utils/vagrant/
vagrant up
vagrant ssh
```

## Building FPC in Vagrant

Once you have logged into the virtual machine using `vagrant ssh`, you can build and test FPC.

Note that the FPC source code from the host is mounted into the virtual machine. 

```bash
vagrant@vagrant:~/projects/src/github.com/hyperledger/fabric-private-chaincode$ make clean && make
```
