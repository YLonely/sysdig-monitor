# sysdig-monitor

sysdig-monitor uses [sysdig](https://github.com/draios/sysdig) to track container events and exports information about running containers(file system, network, system call)

## Install

Before using this tool, you have to [install sysdig](https://github.com/draios/sysdig/wiki/How-to-Install-Sysdig-for-Linux) first.

To use sysdig-monitor, you can just build it from source:
```
    go build -mod vendor
```
Or you can download from release page.

## Running inside a Docker container

sysdig-monitor can also run inside a Docker container. first, the kernel headers must be installed in the host operating system.
Debian-like dist:
```
    apt-get -y install linux-headers-$(uname -r)
```
RHEL-like distributions:
```
    yum -y install kernel-devel-$(uname -r)
```
Pull the sysdig-monitor image:
```
    docker pull lwyan/sysdig-monitor
```
Run the sysdig-monitor
```
    docker run -it --privileged -v /var/run/docker.sock:/host/var/run/docker.sock -v /dev:/host/dev -v /proc:/host/proc:ro -v /boot:/host/boot:ro -v /lib/modules:/host/lib/modules:ro -v /usr:/host/usr:ro -v /var/lib/docker:/var/lib/docker lwyan/sysdig-monitor
```

## Usage

Use `./sysdig-monitor` to start the monitor, now sysdig-monitor is running on **http://localhost:port**, or use can use `--port` flag to set another port.

Now you can start some containers on your host as you wish.

To list all containers running on the host, you can visit **localhost:port/container/** (not **localhost:port/container**) in your web browser, or use `curl` to fetch the result. The result will be in json format as below
```json
    {
        "container-id": "container-name",
        "": ""
    }
```

To get all the information about a container, visit **localhost:port/container/:id** in your web browser or use `curl localhost:port/container/:id | jq` to get a more user-friendly result. The result will be in json format. Visit [here](/docs/format.md) to read about the format.

API **localhost:port/metrics** is also available, which exports some useful metrics in promethuse-style