// Package docker provides a client for the Docker Engine API.
package docker

import (
	"bytes"
	"encoding/json"
	"net"
	"os/exec"
	"testing"
)

// Container tracks information about a docker container started for tests.
type Container struct {
	ID   string
	Host string //IP:Port
}

// StartContainer starts a container for testing.
func StartContainer(t *testing.T, image string, port string, args ...string) *Container {
	arg := []string{"run", "-P", "-d"}
	arg = append(arg, args...)
	arg = append(arg, image)

	cmd := exec.Command("docker", arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container %s: %v", image, err)
	}

	id := out.String()[:12]

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
	}
	var doc []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("could not unmarshal JSON: %v", err)
	}

	ip, randPort := extractIPAndPort(t, doc, port)

	c := Container{
		ID:   id,
		Host: net.JoinHostPort(ip, randPort),
	}

	t.Logf("image:		 %s", image)
	t.Logf("containerID 	 %s", c.ID)
	t.Logf("host:		 %s", c.Host)

	return &c
}

// StopContainer stops and removes a container.
func StopContainer(t *testing.T, ID string) {
	cmd := exec.Command("docker", "stop", ID)
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not stop container %s: %v", ID, err)
	}
	t.Log("stopped", ID)

	if err := exec.Command("docker", "rm", ID, "-v").Run(); err != nil {
		t.Fatalf("could not remove container %s: %v", ID, err)
	}

	t.Log("removed", ID)
}

// DumpContainerLogs outputs the logs for the running container.
func DumpContainerLogs(t *testing.T, ID string) {
	out, err := exec.Command("docker", "logs", ID).CombinedOutput()
	if err != nil {
		t.Fatalf("could not get container logs: %v", err)
	}
	t.Logf("container %s logs:\n%s", ID, out)
}
func extractIPAndPort(t *testing.T, doc []map[string]interface{}, port string) (string, string) {
	nw, exists := doc[0]["NetworkSettings"]
	if !exists {
		t.Fatal("could not find network settings")
	}
	ports, exists := nw.(map[string]interface{})["Ports"]
	if !exists {
		t.Fatal("could not find ports settings")
	}
	tcp, exists := ports.(map[string]interface{})[port+"/tcp"]
	if !exists {
		t.Fatal("could not find port/tcp settings")
	}
	list, exists := tcp.([]interface{})
	if !exists {
		t.Fatal("could not find list of ports")
	}

	var hostIP string
	var hostPort string
	for _, i := range list {
		data, exists := i.(map[string]interface{})
		if !exists {
			t.Fatal("could not get network ports\tcp list data")
		}
		hostIP = data["HostIp"].(string)

		if hostIP != "::" {
			hostPort = data["HostPort"].(string)
		}
	}

	return hostIP, hostPort
}
