// Copyright 2021 yvoilee.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type PodSandbox struct {
	PodName   string
	Namespace string
	Pid       int
}

type ClientWithCache struct {
	isPodSandbox map[string]bool
	podSandboxes map[string]PodSandbox
	*client.Client
}

func New() (ClientWithCache, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return ClientWithCache{}, err
	}

	return ClientWithCache{
		isPodSandbox: map[string]bool{},
		podSandboxes: map[string]PodSandbox{},
		Client:       cli,
	}, nil
}

func (cli *ClientWithCache) GetPodSandbox(ctx context.Context, containerId string) (sandbox PodSandbox, isSandbox bool, err error) {
	if isSandbox, found := cli.isPodSandbox[containerId]; found {
		if isSandbox {
			return cli.podSandboxes[containerId], true, nil
		} else {
			return PodSandbox{}, false, nil
		}
	}

	sandbox, isSandbox, err = cli.getPodSandbox(ctx, containerId)
	if err != nil {
		return PodSandbox{}, false, err
	}
	cli.isPodSandbox[containerId] = isSandbox
	if isSandbox {
		cli.podSandboxes[containerId] = sandbox
	}

	return sandbox, isSandbox, nil
}

func (cli *ClientWithCache) getPodSandbox(ctx context.Context, id string) (sandbox PodSandbox, isSandbox bool, err error) {
	info, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return PodSandbox{}, false, err
	}

	var found bool
	var dockerType string
	if dockerType, found = info.Config.Labels["io.kubernetes.docker.type"]; !found {
		return PodSandbox{}, false, nil
	}
	if dockerType != "podsandbox" {
		return PodSandbox{}, false, nil
	}

	var podName string
	if podName, found = info.Config.Labels["io.kubernetes.pod.name"]; !found {
		return PodSandbox{}, false, nil
	}
	var namespace string
	if namespace, found = info.Config.Labels["io.kubernetes.pod.namespace"]; !found {
		return PodSandbox{}, false, nil
	}

	sandbox = PodSandbox{
		PodName:   podName,
		Namespace: namespace,
		Pid:       info.State.Pid,
	}

	return sandbox, true, nil
}

func (cli *ClientWithCache) ListPodSandboxes(ctx context.Context, namespaces map[string]struct{}) ([]PodSandbox, error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	sandboxes := make([]PodSandbox, 0)
	for _, c := range containers {
		sandbox, isSandbox, err := cli.GetPodSandbox(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		_, ok := namespaces[sandbox.Namespace]
		_, all := namespaces["all"]
		if isSandbox && (all || ok) {
			sandboxes = append(sandboxes, sandbox)
		}
	}

	return sandboxes, nil
}
