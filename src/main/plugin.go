// Docker Image Authorization Plugin.
// Allows docker images to be fetched from a list of authorized registries only.
// AUTHOR: Chaitanya Prakash N <cpdevws@gmail.com>
package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	dockerapi "github.com/docker/docker/api"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/authorization"
)

var imgProcessing = make(map[string]bool)

// Image Authorization Plugin struct definition
type ImgAuthZPlugin struct {
	// Docker client
	client *dockerclient.Client
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
func newPlugin(dockerHost string, notary string) (*ImgAuthZPlugin, error) {
	client, err := dockerclient.NewClient(dockerHost, dockerapi.DefaultVersion, nil, nil)

	if err != nil {
		return nil, err
	}

	return &ImgAuthZPlugin{
		client:           client,
		authorizedNotary: notary}, nil
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

// Authorizes the docker client command.
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

	// There are no authorized registries.
	if len(plugin.authorizedNotary) == 0 {
		// So, deny the request by default!
		log.Println("[DENIED] No authorized notaries configured", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: false, Msg: "No authorized notaries configured"}
	}

	// Enforce DCT
	if _, found := imgProcessing[requestedImage]; found {
		log.Println("Image is already being processed: ", requestedImage, ". Request:", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: true}
	}
	log.Println("Enforcing DCT on", requestedImage, ". Request:", req.RequestMethod, reqURL.String())
	cmd := exec.Command("docker", "image", "pull", requestedImage)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DOCKER_CONTENT_TRUST=1")
	cmd.Env = append(cmd.Env, "DOCKER_CONTENT_TRUST_SERVER=" + plugin.authorizedNotary)
	imgProcessing[requestedImage] = true
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("[DENIED]", requestedImage, ". Reason:", string(out))
		delete(imgProcessing, requestedImage)
		return authorization.Response{Allow: false, Msg: string(out)}
	}
	log.Println("[ALLOWED]", requestedImage)
	delete(imgProcessing, requestedImage)
	return authorization.Response{Allow: true}
}

// AuthZRes authorizes the docker client response.
// All responses are allowed by default.
func (plugin *ImgAuthZPlugin) AuthZRes(req authorization.Request) authorization.Response {
	// Allowed by default.
	return authorization.Response{Allow: true}
}
