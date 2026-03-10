package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ctfer-io/chall-manager/sdk"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type dockerConfigFile struct {
	Auths map[string]dockerConfigAuth `json:"auths"`
}

type dockerConfigAuth struct {
	Auth string `json:"auth"`
}

func loadRegistryCredentials(configPath string, registry string) (string, string, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", err
	}

	var config dockerConfigFile
	if err := json.Unmarshal(content, &config); err != nil {
		return "", "", err
	}

	registryKeys := []string{
		registry,
		strings.TrimSuffix(registry, "/"),
		"https://" + strings.TrimSuffix(registry, "/"),
		"http://" + strings.TrimSuffix(registry, "/"),
	}

	for _, key := range registryKeys {
		authEntry, ok := config.Auths[key]
		if !ok || authEntry.Auth == "" {
			continue
		}

		decoded, err := base64.StdEncoding.DecodeString(authEntry.Auth)
		if err != nil {
			return "", "", err
		}

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			return "", "", fmt.Errorf("invalid auth entry for registry %s", registry)
		}

		return credentials[0], credentials[1], nil
	}

	return "", "", fmt.Errorf("registry credentials not found for %s", registry)
}

func main() {
	sdk.Run(func(req *sdk.Request, resp *sdk.Response, opts ...pulumi.ResourceOption) error {

		// Image configuration
		image, ok := req.Config.Additional["image"]
		if !ok {
			return fmt.Errorf("image not specified in config")
		}

		portStr, ok := req.Config.Additional["port"]
		if !ok {
			return fmt.Errorf("port not specified in config")
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return err
		}

		hostname, ok := req.Config.Additional["hostname"]
		if !ok {
			return fmt.Errorf("hostname not specified in config")
		}

		protocol_port, ok := req.Config.Additional["protocol_port"]
		if !ok {
			protocol_port = "tcp"
		}

		protocol_url, ok := req.Config.Additional["protocol_url"]
		if !ok {
			return fmt.Errorf("protocol_url not specified in config")
		}

		// Docker configuration
		docker_host, ok := req.Config.Additional["docker_host"]
		if !ok {
			return fmt.Errorf("docker_host not specified in config")
		}

		var provArgs docker.ProviderArgs
		provArgs = docker.ProviderArgs{
			Host: pulumi.String(docker_host),
		}

		if strings.HasPrefix(docker_host, "ssh") {
			provArgs.SshOpts = pulumi.StringArray{
				pulumi.String("-o"), pulumi.String("StrictHostKeyChecking=no"),
				pulumi.String("-o"), pulumi.String("UserKnownHostsFile=/dev/null"),
			}
		}

		registry, ok := req.Config.Additional["registry"]
		if ok {
			provArgs.RegistryAuth = docker.ProviderRegistryAuthArray{
				&docker.ProviderRegistryAuthArgs{
					Address:    pulumi.String(registry),
					ConfigFile: pulumi.String("/root/.docker/config.json"),
				},
			}
		}
		prov, err := docker.NewProvider(req.Ctx, "provider", &provArgs)
		if err != nil {
			return err
		}
		opts = append(opts, pulumi.Provider(prov))

		// pull image
		img, err := docker.NewRemoteImage(req.Ctx, "challenge-image", &docker.RemoteImageArgs{
			Name:        pulumi.String(image),
			Platform:    pulumi.String("linux/amd64"),
			KeepLocally: pulumi.Bool(true), // do not remove image (same image for several instances)

		}, opts...)
		if err != nil {
			return err
		}

		// create a swarm service constrained to nodes tagged for challenges
		serviceArgs := &docker.ServiceArgs{
			Name: pulumi.StringPtr(fmt.Sprintf("challenge-%s", req.Config.Identity)),
			Mode: docker.ServiceModeArgs{
				Replicated: docker.ServiceModeReplicatedArgs{
					Replicas: pulumi.IntPtr(1),
				},
			},
			TaskSpec: docker.ServiceTaskSpecArgs{
				ContainerSpec: docker.ServiceTaskSpecContainerSpecArgs{
					Image: img.Name,
				},
				Placement: docker.ServiceTaskSpecPlacementArgs{
					Constraints: pulumi.StringArray{
						pulumi.String("node.labels.type==challs"),
					},
				},
			},
			EndpointSpec: docker.ServiceEndpointSpecArgs{
				Ports: docker.ServiceEndpointSpecPortArray{
					docker.ServiceEndpointSpecPortArgs{
						Protocol:   pulumi.StringPtr(protocol_port),
						TargetPort: pulumi.Int(port),
						// PublishedPort intentionally omitted to let Swarm allocate an available port.
					},
				},
			},
			ConvergeConfig: docker.ServiceConvergeConfigArgs{
				Delay:   pulumi.StringPtr("1s"),
				Timeout: pulumi.StringPtr("15s"),
			},
		}

		if registry != "" {
			username, password, err := loadRegistryCredentials("/root/.docker/config.json", registry)
			if err != nil {
				return err
			}

			serviceArgs.Auth = docker.ServiceAuthArgs{
				ServerAddress: pulumi.String(registry),
				Username:      pulumi.StringPtr(username),
				Password:      pulumi.StringPtr(password),
			}
		}

		service, err := docker.NewService(req.Ctx, "challenge-service", serviceArgs, opts...)
		if err != nil {
			return err
		}

		resp.ConnectionInfo = service.EndpointSpec.Ports().ApplyT(func(ports []docker.ServiceEndpointSpecPort) string {
			external := port
			if len(ports) > 0 && ports[0].PublishedPort != nil {
				external = *ports[0].PublishedPort
			}

			url := fmt.Sprintf("%s%s%s:%d", protocol_url, func() string {
				if strings.HasPrefix(protocol_url, "http") {
					return "://"
				}
				return " "
			}(), hostname, external)
			return url
		}).(pulumi.StringOutput)

		return nil
	})
}
