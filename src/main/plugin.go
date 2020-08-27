// Docker Image Authorization Plugin.
// Allows docker images to be fetched from a list of authorized registries only.
// AUTHOR: Chaitanya Prakash N <cpdevws@gmail.com>
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"

	dockerapi "github.com/docker/docker/api"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/authorization"
)

// ImgAuthZPlugin Image Authorization Plugin struct definition
type ImgAuthZPlugin struct {
	// Docker client
	client *dockerclient.Client
	// Map of authorized registries
	authorizedRegistries map[string]bool
	// Number of authorized registries
	numAuthorizedRegistries int
	// List of authorized registries as string
	authRegistriesAsString string
	// Authorized notary
	authorizedNotary string
}

// Returns the list of authorized registries as string
func authRegistries(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// Create a new image authorization plugin
func newPlugin(dockerHost string, registries map[string]bool, notary string) (*ImgAuthZPlugin, error) {
	client, err := dockerclient.NewClient(dockerHost, dockerapi.DefaultVersion, nil, nil)

	if err != nil {
		return nil, err
	}

	return &ImgAuthZPlugin{
		client:                  client,
		authorizedRegistries:    registries,
		authorizedNotary:        notary,
		numAuthorizedRegistries: len(registries),
		authRegistriesAsString:  authRegistries(registries)}, nil
}

// Parses the docker client command to determine the requested registry used in the command.
// If a registry is used in the command (i.e. docker pull or docker run commands), then the registry url and true is returned.
// Otherwise, returns empty string and false.
func (plugin *ImgAuthZPlugin) isImageCommand(req authorization.Request, reqURL *url.URL) (string, bool) {

	image := ""

	// docker run
	if strings.HasSuffix(reqURL.Path, "/containers/create") {
		var config dockercontainer.Config
		json.Unmarshal(req.RequestBody, &config)
		image = config.Image
	}

	// docker pull
	if strings.HasSuffix(reqURL.Path, "/images/create") {
		image = reqURL.Query().Get("fromImage")
	}

	if len(image) > 0 {
		return image, true
	}

	return "", false
}

// Returns true if there are any authorized registries configured.
// Otherwise, returns false
func (plugin *ImgAuthZPlugin) hasAuthorizedRegistries() bool {
	return (plugin.numAuthorizedRegistries > 0)
}

// Parses the docker client command to determine the requested registry used in the command.
// If a registry is used in the command (i.e. docker pull or docker run commands), then the registry url and true is returned.
// Otherwise, returns empty string and false.
func (plugin *ImgAuthZPlugin) getRequestedRegistry(req authorization.Request, reqURL *url.URL) (string, bool) {

	image := ""
	registry := ""

	// docker run
	if strings.HasSuffix(reqURL.Path, "/containers/create") {
		var config dockercontainer.Config
		json.Unmarshal(req.RequestBody, &config)
		image = config.Image
	}

	// docker pull
	if strings.HasSuffix(reqURL.Path, "/images/create") {
		image = reqURL.Query().Get("fromImage")
	}

	if len(image) > 0 {
		// If no registry is specfied, assume it is the dockerhub!
		registry = "library"
		idx := strings.Index(image, "/")
		if idx != -1 {
			registry = image[0:idx]
		}
		return registry, true
	}

	return registry, false
}

// AuthZReq Authorizes the docker client command.
// Non registry related commands are allowed by default.
// If the command uses a registry, the command is allowed only if the registry is authorized.
// Otherwise, the request is denied!
func (plugin *ImgAuthZPlugin) AuthZReq(req authorization.Request) authorization.Response {
	// Parse request and the request body
	reqURI, _ := url.QueryUnescape(req.RequestURI)
	reqURL, _ := url.ParseRequestURI(reqURI)

	// Find out the requested registry and whether or not a registry is present in the client command
	requestedImage, isImageCommand := plugin.isImageCommand(req, reqURL)

	// Docker command do not involve registries
	if isImageCommand == false {
		// Allowed by default!
		log.Println("[ALLOWED] Not image or container creation command:", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: true}
	}

	// Find out the requested registry and whether or not a registry is present in the client command
	requestedRegistry, isRegistryCommand := plugin.getRequestedRegistry(req, reqURL)

	// Docker command do not involve registries
	if isRegistryCommand == false {
		// Allowed by default!
		log.Println("[ALLOWED] Not a registry command:", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: true}
	}

	// There are no authorized registries.
	if plugin.hasAuthorizedRegistries() == false {
		// So, deny the request by default!
		log.Println("[DENIED] No authorized registries", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: false, Msg: "No authorized registries configured"}
	}

	// Verify that registry requested is authorized
	if plugin.authorizedRegistries[requestedRegistry] == true {
		// Is an authorized registry: Allow!
		log.Println("[ALLOWED] Registry:", requestedRegistry, req.RequestMethod, reqURL.String())
		// return authorization.Response{Allow: true}
	} else {
		// Oops.. The requested registry is not authorized. Deny the request!
		log.Println("[DENIED] Registry:", requestedRegistry, req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: false, Msg: "You can only use docker images from the following authorized registries: " + plugin.authRegistriesAsString}
	}

	// There are no authorized notaries.
	if len(plugin.authorizedNotary) == 0 {
		// So, deny the request by default!
		log.Println("[DENIED] No authorized notaries configured", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: false, Msg: "No authorized notaries configured"}
	}

	// Enforce DCT
	log.Println("Enforcing DCT on", requestedImage, ". Request:", req.RequestMethod, reqURL.String())
	imageTag := strings.Split(requestedImage, ":")
	image := imageTag[0]
	var tag string
	if len(imageTag) > 1 {
		tag = imageTag[1]
	} else {
		tag = "latest"
	}
	notaryURL, _ := url.ParseRequestURI(plugin.authorizedNotary)
	cmd := exec.Command("notary",
		"-s", plugin.authorizedNotary, "-d", "/root/.docker/trust", "--tlscacert",
		fmt.Sprintf("/root/.docker/tls/%s/root-ca.crt", notaryURL.Host),
		"lookup", image, tag)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("[DENIED]", image+":"+tag, ". Reason:", string(out))
		log.Println(err)
		return authorization.Response{Allow: false, Msg: string(out)}
	}
	log.Println("[ALLOWED]", image+":"+tag)
	return authorization.Response{Allow: true}
}

// AuthZRes Authorizes the docker client response.
// All responses are allowed by default.
func (plugin *ImgAuthZPlugin) AuthZRes(req authorization.Request) authorization.Response {
	// Allowed by default.
	return authorization.Response{Allow: true}
}
