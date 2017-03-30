# Setup for Running OVS-DPDK

The Data Plane Development Kit (DPDK) is a set of libraries and drivers for fast packet processing.  Open vSwitch (OVS) is an open-source implementation of a distributed multilayer switch.  Enabling DPDK support within OVS can yield significant performance improvements over linux-bridge, providing a switch with DPDK vhost-user ports.

## Installation of DPDK and OVS on host system

There is tight coupling between OVS and DPDK versioning, and what is available via distributions is often very out of date.  The directions below will instruct you how to build and configure from source. You MUST update the packages listed regularly to ensure your Clear Containers benefit from performance improvements, new features and most importantly CVE security fixes.

As a prerequisite to building, you'll need a few packages on your host.  As an example, on an Ubuntu host system you'd need to install as follows:
```
$ sudo apt-get -y install autoconf automake kernel-common \
    libpcap-dev libtool python python-six
```

Next we will set up DPDK:
```
$ git clone http://dpdk.org/git/dpdk -b v16.11 $HOME/dpdk
$ cd $HOME/dpdk
$ export RTE_SDK=$(pwd)
$ export RTE_TARGET=x86_64-native-linuxapp-gcc
$ make -j install T=$RTE_TARGET DESTDIR=install EXTRA_CFLAGS='-g'
$ sudo cp -f install/lib/lib* /lib64/
$ export DPDK_BUILD_DIR=x86_64-native-linuxapp-gcc
```
Once this completes without issue, build and configure OVS:
```
$ git clone https://github.com/openvswitch/ovs -b branch-2.6 $HOME/ovs
$ cd $HOME/ovs
$ ./boot.sh
$ ./configure --prefix=/usr --sysconfdir=/etc --localstatedir=/var \
    --with-dpdk="$RTE_SDK/$DPDK_BUILD_DIR"  --disable-ssl --with-debug  CFLAGS='-g'
$ sudo -E make install -j
```

## Running OVS-DPDK on host system
Once installed without error, we'll need to start up the ovsdb-server and ovs-vswitchd. The Open_vSwitch configuration settings will vary depending on the hardware you are running on (ex: lcore-mask, socket-mem). An example configuration is shown below, which should be adjusted for your setup:

```
$ sudo mkdir -p /var/run/openvswitch
$ sudo killall ovsdb-server ovs-vswitchd
$ sudo rm -f /var/run/openvswitch/vhost-user*
$ sudo rm -f /etc/openvswitch/conf.db

$ export DB_SOCK=/var/run/openvswitch/db.sock
$ sudo -E ovsdb-tool create /etc/openvswitch/conf.db /usr/share/openvswitch/vswitch.ovsschema
$ sudo -E ovsdb-server --remote=punix:$DB_SOCK --remote=db:Open_vSwitch,Open_vSwitch,manager_options --pidfile --detach

$ sudo -E ovs-vsctl --no-wait init
$ sudo -E ovs-vsctl --no-wait set Open_vSwitch . other_config:dpdk-lcore-mask=0x3
$ sudo -E ovs-vsctl --no-wait set Open_vSwitch . other_config:dpdk-socket-mem=1024,0
$ sudo -E ovs-vsctl --no-wait set Open_vSwitch . other_config:dpdk-init=true

$ sudo sysctl -w vm.nr_hugepages=4096
$ sudo mount -t hugetlbfs -o pagesize=2048k none /dev/hugepages

$ sudo -E ovs-vswitchd unix:$DB_SOCK --pidfile --detach --log-file=/var/log/openvswitch/ovs-vswitchd.log
$ ps -ae | grep ovs
```
In the example above, there are a few configuration options which we set for DPDK:
*dpdk-lcore-mask*: Specifies the CPU cores on which DPDK threads should be spawned.  We assign 0x3, which is identifying cores 0 and 1.  This was an explicit decision based on the setup of the setup of the host system after monitoring the output of lspci to see which cores are tied to the same NUMA node.
*dpdk-socket-mem*: Comma seperated list of memory (in MB) to pre-allocate from hugepages on specific sockets.  In this case, I am requesting 1024 MB from socket 0, which is where cores 0 and 1 are also located.
*dpdk-init*:  Specifices whether OVS should initialize and support DPDK ports.

The number of hugepages mounted at /dev/hugepages was selected given the size allocated for each VM as well as the size allocated to OVS itself.  OVS-DPDK enabled Clear Containers preallocate 2GB per container from hugepages,  1024 MB was selected for OVS-DPDK, thus 5 GB was the minimum to reserve.

After running above and modifying as necessary, your host system is ready to start creating DPDK enabled OVS switches.

## Grab and install OVSDPDK  Docker plugin
We need to install an OVS-DPDK Docker plugin in order to facilitate creating a network and connecting Clear Containers to this network via Docker.

Details on pulling and installing the plugin can be found at https://github.com/clearcontainers/ovsdpdk

## Example: Launching two Clear Containers using OVS-DPDK

To make use of OVS-DPDK, the next step is to use Docker to create a network using the OVS-DPDK switch.  An example:
```
$ sudo docker network create -d=ovsdpdk --ipam-driver=ovsdpdk --subnet=192.168.1.0/24 --gateway=192.168.1.1  --opt "bridge"="ovsbr" ovsdpdk_net
```
You can verify the OVS switch was created by looking for the OVS bridge, ovsbr, in the output of:
```
$ sudo ovs-vsctl show
```

You can now test connectivity by launching two containers as follows (this assumes that you have Docker setup to use this branch's runtime):
```
$ sudo docker run --net=ovsdpdk_net --ip=192.168.1.2 --mac-address=CA:FE:CA:FE:01:02 -it debian bash -c "ip a; ip route; sleep 300"
$ sudo docker run --net=ovsdpdk_net --ip=192.168.1.3 --mac-address=CA:FE:CA:FE:01:03 -it debian bash -c "ip a; ip route; ping 192.168.1.2"
```
This will setup two VMs, the first of which will display networking details and then sleep, providing a period of time during which it can be pinged.  The second VM will display its networking details and then ping the first VM, verifying connectivity between the VMs.

After verifying, you can cleanup:
```
$ sudo docker kill $(sudo docker ps --no-trunc -aq)
$ sudo docker rm $(sudo docker ps --no-trunc -aq)
$ sudo docker network rm ovsdpdk_net
```

