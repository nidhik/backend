/*

Modified from code found here: "github.com/niilo/golib/test/dockertest"

Copyright 2014 The Camlistore Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package dockertest contains helper functions for setting up and tearing down docker containers to aid in testing.
*/
package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var port = 27017
var mongo_container_port = "27017:27017"
var MONGO_URI = "127.0.0.1:27017"

/// runLongTest checks all the conditions for running a docker container
// based on image.
func runLongTest(t *testing.T, image string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	if !haveDocker() {
		t.Error("'docker' command not found")
	}
	if ok, err := haveImage(image); !ok || err != nil {
		if err != nil {
			t.Errorf("Error running docker to check for %s: %v", image, err)
		}
		log.Printf("Pulling docker image %s ...", image)
		if err := Pull(image); err != nil {
			t.Errorf("Error pulling %s: %v", image, err)
		}
	}
}

// haveDocker returns whether the "docker" command was found.
func haveDocker() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func haveImage(name string) (ok bool, err error) {
	out, err := exec.Command("docker", "images", "--no-trunc").Output()
	if err != nil {
		return
	}
	return bytes.Contains(out, []byte(name)), nil
}

func run(args ...string) (containerID string, err error) {
	cmd := exec.Command("docker", append([]string{"run"}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("%v%v", stderr.String(), err)
		return
	}
	containerID = strings.TrimSpace(stdout.String())
	if containerID == "" {
		return "", errors.New("unexpected empty output from `docker run`")
	}
	return
}

func KillContainer(container string) error {
	return exec.Command("docker", "kill", container).Run()
}

// Pull retrieves the docker image with 'docker pull'.
func Pull(image string) error {
	out, err := exec.Command("docker", "pull", image).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%v: %s", err, out)
	}
	return err
}

// IP returns the IP address of the container.
func IP(containerID string) (string, error) {
	out, err := exec.Command("docker", "inspect", containerID).Output()
	if err != nil {
		return "", err
	}
	type networkSettings struct {
		IPAddress string
	}
	type container struct {
		NetworkSettings networkSettings
	}
	var c []container
	if err := json.NewDecoder(bytes.NewReader(out)).Decode(&c); err != nil {
		return "", err
	}
	if len(c) == 0 {
		return "", errors.New("no output from docker inspect")
	}
	if ip := c[0].NetworkSettings.IPAddress; ip != "" {
		return ip, nil
	}
	return "", errors.New("could not find an IP. Not running?")
}

type ContainerID string

func (c ContainerID) IP() (string, error) {
	return IP(string(c))
}

func (c ContainerID) Kill() error {
	return KillContainer(string(c))
}

// Remove runs "docker rm" on the container
func (c ContainerID) Remove() error {
	return exec.Command("docker", "rm", "-v", string(c)).Run()
}

// KillRemove calls Kill on the container, and then Remove if there was
// no error. It logs any error to t.
func (c ContainerID) KillRemove(t *testing.T) {
	if err := c.Kill(); err != nil {
		t.Log(err)
		return
	}
	if err := c.Remove(); err != nil {
		t.Log(err)
	}
}

// lookup retrieves the ip address of the container, and tries to reach
// before timeout the tcp address at this ip and given port.
func (c ContainerID) lookup(port int, timeout time.Duration) (ip string, err error) {
	ip, err = c.IP()
	if err != nil {
		err = fmt.Errorf("error getting IP: %v", err)
		return
	}
	return
}

// setupContainer sets up a container, using the start function to run the given image.
// It also looks up the IP address of the container, and tests this address with the given
// port and timeout. It returns the container ID and its IP address, or makes the test
// fail on error.
func setupContainer(t *testing.T, image string, port int, timeout time.Duration,
	start func() (string, error)) (c ContainerID, ip string) {
	runLongTest(t, image)

	containerID, err := start()
	if err != nil {
		t.Fatalf("docker run: %v", err)
	}
	c = ContainerID(containerID)
	ip, err = c.lookup(port, timeout)
	if err != nil {
		c.KillRemove(t)
		t.Errorf("Container %v setup failed: %v", c, err)
	}
	return
}

const (
	mongoImage = "mongo"
)

// SetupMongoContainer sets up a real MongoDB instance for testing purposes,
// using a Docker container. It returns the container ID and its IP address,
// or makes the test fail on error.
// Currently using https://hub.docker.com/_/mongo/
func SetupMongoContainer(t *testing.T) (c ContainerID, ip string) {
	return setupContainer(t, mongoImage, port, 10*time.Second, func() (string, error) {
		return run("-dt", "-p", mongo_container_port, "--name", "dockertestmongodb", mongoImage)
	})
}
