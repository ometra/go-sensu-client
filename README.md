Sensu Client
============
sensu-client is an implementation of the [sensu](https://github.com/sensu/sensu) 
client written in [Golang](https://golang.org).

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

Gotchas
=======
* Currently only linux is supported.
* Gathering current CPU Frequency requires root access to:

	/sys/devices/system/cpu/<cpu>/cpufreq/cpuinfo_cur_freq