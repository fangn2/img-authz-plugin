# Docker Image Authorization Plugin

The image authorization plugin allows docker images only from a list of
authorized registries and notaries to be used by the docker engine. For
additional information, please refer to [docker
documentation](https://docs.docker.com/engine/extend/) on plugins.


## Build and package the plugin

Build the plugin filesystem:

`make build`

Create the actual Docker plugin:

`make create`

Publish the plugin to the registry:

`make push`

---

**Alternatively** you can also execute the three steps from above with `make all`


## Install the plugin


### From source

To install the plugin from this repo:

`make local_install REGISTRY=<registry>`

(if you've already run `make build` and `make create`, then you can simply run
`make enable REGISTRY=<registry>`).

Add the following JSON key value to `/etc/docker/daemon.json`:

```json
"authorization-plugins": ["sixsq/img-authz-plugin:latest"]
```

and run `kill -SIGHUP $(pidof dockerd)`

### From a Docker registry

**(recommended)**

To get and install the plugin, simply run:

`docker plugin install sixsq/img-authz-plugin:latest REGISTRY=<registry>`

Where `REGISTRY` is host[:port] of the registry to be authorized.

Add the following JSON key value to `/etc/docker/daemon.json`:

```json
"authorization-plugins": ["sixsq/img-authz-plugin:latest"]
```

and run `kill -SIGHUP $(pidof dockerd)`


## Update the registries in a running plugin

First disable the plugin:

`docker plugin disable sixsq/img-authz-plugin:latest`

Then set the new registries value:

`docker plugin set sixsq/img-authz-plugin:latest REGISTRY=<registry>`

Re-enable the plugin, and reload the Docker daemon:

`docker plugin enable sixsq/img-authz-plugin:latest && kill -SIGHUP $(pidof dockerd)`


## Test the plugin

To test the plugin locally before publishing:


```
# Create the plugin tests base image
docker build -t plugin-tests -f Dockerfile.test --build-arg DOCKER_VERSION=1.12.6 --rm .

# Start the plugin tests container
docker run --privileged -d --name plugin-tests-1.12.6 plugin-tests-1.12.6

# Run the plugin tests
docker exec plugin-tests-1.12.6 python tests.py

# Remove the plugin tests container
docker rm -f plugin-tests-1.12.6

# Remove the plugin tests image
docker rmi -f plugin-tests-1.12.6
```


#### Stop and uninstall the plugin
NOTE: Before doing below, remove the authorization-plugin configuration created above and restart the docker daemon.
```
# Stop the plugin service
systemctl stop img-authz-plugin
systemctl disable img-authz-plugin

# Uninstall the plugin service units
docker run --rm -v `pwd`:`pwd` -w `pwd` -e GOPATH=`pwd` -v /usr/libexec:/usr/libexec \
  -v /usr/lib/systemd/system:/usr/lib/systemd/system plugin-build-tools:latest \
  make uninstall

```

#### To remove the generated artifacts
```
docker run --rm -v `pwd`:`pwd` -w `pwd` -e GOPATH=`pwd` plugin-build-tools:latest make clean
```

#### Access plugin logs
```
journalctl -xe -u img-authz-plugin -f
```
