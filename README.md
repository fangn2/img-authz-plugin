# SixSq's Docker Image Authorization Plugin

This image authorization plugin is a fork and adaptation from the original upstream project [crosslibs/img-authz-plugin](https://github.com/crosslibs/img-authz-plugin). 

The plugin makes sure that all Docker Registry-related requests are limited to a user-specified Docker Registry endpoint and for trusted Docker images only. 

Docker and Notary tools are used to enforce and verify the Docker Registry and Image verification workflows.

For additional information, please refer to [docker
documentation](https://docs.docker.com/engine/extend/) on plugins, or the original plugin at [https://github.com/crosslibs/img-authz-plugin](https://github.com/crosslibs/img-authz-plugin).

![arch.png](./docs/arch.png)

```

██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗██╗███╗   ██╗ ██████╗     ████████╗███████╗███████╗████████╗███████╗
██╔══██╗██║   ██║████╗  ██║████╗  ██║██║████╗  ██║██╔════╝     ╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔════╝
██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║██║██╔██╗ ██║██║  ███╗       ██║   █████╗  ███████╗   ██║   ███████╗
██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██║██║╚██╗██║██║   ██║       ██║   ██╔══╝  ╚════██║   ██║   ╚════██║
██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║██║██║ ╚████║╚██████╔╝       ██║   ███████╗███████║   ██║   ███████║██╗██╗██╗
╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚═╝╚═╝  ╚═══╝ ╚═════╝        ╚═╝   ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝╚═╝╚═╝

```

## Build and package the plugin

 1. Build the plugin filesystem:

    ```bash
    make build
    ```
    
    This will build a base Docker image with the compiled plugin code and necessary dependencies. This step will also create a staging container from this Docker image and extract its filesystem to the plugin's build directory, `rootfs` (also configured in this step), as mandated by the [Docker Plugin documentation](https://docs.docker.com/engine/extend/#developing-a-plugin)

 2. Create the actual Docker plugin:

    ```bash
    make create
    ```
    
    This will **cleanup any previous installation of the plugin**, if the names match. So please **make sure** your Docker Daemon is not running with this plugin's name enabled (meaning that `--authorization-plugin` is not set).
    
    At the end of this step, you'll have a new Docker plugin, but disabled.

 3. Publish the plugin to the registry:

    ```bash
    make push
    ```
    
    This step will upload the said Docker Plugin into the specified Docker registry. Please **make sure** you have given Docker your registry credentials before executing this step (`docker login ...`).

---

**Alternatively** you can also execute the three steps from above with `make all`


## Install the plugin

The plugin needs the configuration variables: `REGISTRY`, `NOTARY` and `NOTARY_ROOT_CA`. If you leave those empty, the plugin defaults to `docker.io` and `notary.docker.io`.

 - `REGISTRY`: is the _host:port_ of the registry to be authorized
 - `NOTARY`: is the _https://fqdn:port_ of the Notary server used for signing the images
 - `NOTARY_ROOT_CA`: is the raw public Certificate Authority certificate used for the Notary server TLS certificates 

### From source

To install the plugin from this repo:

`make local_install REGISTRY=<registry> NOTARY=<notary-server> NOTARY_ROOT_CA='''<raw-ca-cert>'''`

(if you've already run `make build` and `make create`, then you can simply run
`make enable REGISTRY=<registry> ...`).

Add the following JSON key value to `/etc/docker/daemon.json`:

```json
"authorization-plugins": ["sixsq/img-authz-plugin:latest"]
```

and run `kill -SIGHUP $(pidof dockerd)`

### From a Docker registry

**(RECOMMENDED)**

To get and install the plugin, simply run:

`docker plugin install sixsq/img-authz-plugin:latest REGISTRY=<registry> NOTARY=<notary-server> NOTARY_ROOT_CA='''<raw-ca-cert>'''`

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

`docker plugin set sixsq/img-authz-plugin:latest REGISTRY=<registry> NOTARY=<notary-server> NOTARY_ROOT_CA='''<raw-ca-cert>'''`

Re-enable the plugin, and reload the Docker daemon:

`docker plugin enable sixsq/img-authz-plugin:latest && kill -SIGHUP $(pidof dockerd)`

## Using a self-signed private registry

If you're using a private Docker registry with self-signed TLS certificates, **please remember** to add the client certificates to your Docker trust directory. 

To do this, please follow the steps at: https://docs.docker.com/registry/insecure/#use-self-signed-certificates

**NOTE:** using insecure Docker registries (without TLS) is not recommended and thus not covered by this plugin. Use it at your own risk.


## Test the plugin

You can test the full plugin build, creation and execution workflow by running:

```bash
make --makefile=Makefile.test 
```

Please **note** that this process, due to its broad coverage, can take around 10 minutes to complete.

At the end of a successful test run, you should get an output like the following:

```

██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗██╗███╗   ██╗ ██████╗     ████████╗███████╗███████╗████████╗███████╗
██╔══██╗██║   ██║████╗  ██║████╗  ██║██║████╗  ██║██╔════╝     ╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔════╝
██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║██║██╔██╗ ██║██║  ███╗       ██║   █████╗  ███████╗   ██║   ███████╗
██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██║██║╚██╗██║██║   ██║       ██║   ██╔══╝  ╚════██║   ██║   ╚════██║
██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║██║██║ ╚████║╚██████╔╝       ██║   ███████╗███████║   ██║   ███████║██╗██╗██╗
╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚═╝╚═╝  ╚═══╝ ╚═════╝        ╚═╝   ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝╚═╝╚═╝


test_docker_pull_is_allowed
test_docker_run_is_allowed
test_pull_is_not_allowed_when_registry_is_not_authorized
test_run_is_not_allowed_when_registry_is_not_authorized
test_pull_is_allowed_when_registry_is_authorized
test_run_is_allowed_when_registry_is_authorized
test_pull_is_allowed_when_tag_not_specified
test_run_is_allowed_when_tag_not_specified

Ran 8 tests.

OK
```

---

**If** you already have the plugin installed and you only want to test the REGISTRY/NOTARY workflows, then you can just quickly run the unit test by doing:

```bash
 docker run --rm -v $(pwd)/test:/tmp -v /var/run/docker.sock:/var/run/docker.sock docker:dind sh -c 'apk update && apk add shunit2 && SHUNIT_COLOR="always" shunit2 /tmp/tests.sh && docker ps'
```

## Plugin logs

The plugin logs are appended to the Docker daemon logs, and thus you can find them in your respective Docker logs' directory (for example, for Ubuntu you can do `journal -u docker`)


## Stop and uninstall the plugin

_(assuming the plugin name is sixsq/img-authz-plugin:latest)_

Stop the plugin:
 1. `docker plugin disable sixsq/img-authz-plugin:latest`
 2. Remove the `authorization-plugins` attribute from /etc/docker/daemon.json
 3. `kill -SIGHUP $(pidof dockerd)`
 
Uninstall the plugin:
 1. `docker plugin rm -f sixsq/img-authz-plugin:latest`
