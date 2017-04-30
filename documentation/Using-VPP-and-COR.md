# Setup for Running VPP

The Data Plane Development Kit (DPDK) is a set of libraries and drivers for fast packet processing.  Vector Packet Processing (VPP) is a platform extensible framework that provides out-of-the-box production quality switch/router fucntionailty.  It is a high performance packet-processing stack that can run on commodity CPUs.  Enabling VPP with DPDK support can yield significant performance improvements over linux-bridge, providing a switch with DPDK vhost-user ports.

For more information about VPP https://wiki.fd.io/view/VPP

## Installation of VPP

Installation directions for VPP can be found at: https://wiki.fd.io/view/VPP/Installing_VPP_binaries_from_packages

After successful installation, your host system is ready to start connecting Clear Containers with VPP bridges

## Grab and install VPP  Docker plugin
We need to install a VPP Docker plugin in order to facilitate creating a network and connecting Clear Containers to this network via Docker.

Details on pulling and installing the plugin can be found at https://github.com/clearcontainers/vpp

This VPP plugin will allow for creating a VPP network. Every container added to this network will be connected via an L2 bridge-domain provided by VPP.

## Example: Launching two Clear Containers using VPP

To make use of VPP, the next step is to use Docker to create a network using the OVS-DPDK switch.  An example:
```
$ sudo docker network create -d=vpp --ipam-driver=vpp --subnet=192.168.1.0/24 --gateway=192.168.1.1  vpp_net
```

You can now test connectivity by launching two containers as follows (this assumes that you have Docker setup to use this branch's runtime):
```
$ sudo docker run --net=vpp_net --ip=192.168.1.2 --mac-address=CA:FE:CA:FE:01:02 -it debian bash -c "ip a; ip route; sleep 300"
$ sudo docker run --net=vpp_net --ip=192.168.1.3 --mac-address=CA:FE:CA:FE:01:03 -it debian bash -c "ip a; ip route; ping 192.168.1.2"
```
This will setup two Clear Containers connected via a VPP L2 bridge domain.  The first of the the VMs will display networking details and then sleep, providing a period of time during which it can be pinged.  The second VM will display its networking details and then ping the first VM, verifying connectivity between the VMs.

After verifying, you can cleanup:
```
$ sudo docker kill $(sudo docker ps --no-trunc -aq)
$ sudo docker rm $(sudo docker ps --no-trunc -aq)
$ sudo docker network rm vpp_net
$ sudo service vpp stop
$ sudo service vpp start
```
