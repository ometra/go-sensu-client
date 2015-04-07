Sensu Client
============
sensu-client is an implementation of the [sensu](https://github.com/sensu/sensu) 
client written in [Golang](https://golang.org).

Forked from https://github.com/chrishoffman/go-sensu-client

It has a bunch of checks that are compiled into the main binary, this binary can 
then be built and deployed to a variety of platforms as single binary and config
file.

Because of it's native compiled nature it is a lightweight and fast alternative
to the full sensu-client.

Instructions
============

For this project we use a unique GOPATH. All commands must be run from root of
the checked out repository.

Setup
-----
Once you have cloned the repository, change into the checkout folder and run:

	./setup.sh
	
This will "go get" all the supporting libraries.

Configuration
-------------
You will need to setup a config.json file of your own, feel free to copy one of
the .dist files in "src/config/" and modify it for your own needs.

Running
-------
There is a handy shell script that you can use to run the code during 
development, it takes care of setting up the GOPATH, so you will need to run it
from the base of your checked out repository.

	sudo ./run.sh

The gathering of some of the stats requires more permissions than a normal user
has. In production this can be handled by assigning the user the appropriate
permissions, but for development we can simply run as root.

Building
--------
Since we are targetting more than 1 platform here (we are wanting to run this
client under android after all) the build script is a shortcut for making all of
our targets.

### Building For ARM
Golang makes it really easy to cross compile for ARM.

	$ GOARCH=arm go build sensu-client.go
	
The binaries have been tested against the RPi and Rpi2 under Raspbian and they work fine as well! 

### Building under Golang >= 1.4 for Android/ARM
To build for the android platform (GOOS=android) you need to do the following:

I placed all the downloads into ~/dev

	1. Get a copy of Golang >= version 1.4 from https://golang.org/dl/
		* Unpack it into ~/dev
	2. Grab the latest copy of the Android NDK from https://developer.android.com/tools/sdk/ndk/index.html
		* Make it executable and run it (it unpacks into the current folder, so ~/dev)
	3. Time to get a copy of our platform NDK.
		* export NDK_ROOT=~/dev/ndk-toolchain
		* ./android-ndk-r10c/build/tools/make-standalone-toolchain.sh --platform=android-16 --install-dir=$NDK_ROOT
	4. Now we need to build the Golang toolchain, cd into ~/dev/go/src
		* export NDK_CC=~/dev/ndk-toolchain/bin/arm-linux-androideabi-gcc
		* CC_FOR_TARGET=$NDK_CC GOOS=android GOARCH=arm GOARM=7 ./make.bash
	5. Now we can run our sensu build script, change to the go sensu client checked out folder
		* ./build.sh

### Building for MIPS
Stay tuned....

Running
-------
You can get some extra debug information by setting the environment variable DEBUG.
e.g.

	$ DEBUG=1 ./sensu-client-armv7-linux

Gotchas
=======
* Currently only linux/android is supported.
* TCP checks require root access
* Gathering current CPU Frequency requires root access to:

	/sys/devices/system/cpu/<cpu>/cpufreq/cpuinfo_cur_freq

