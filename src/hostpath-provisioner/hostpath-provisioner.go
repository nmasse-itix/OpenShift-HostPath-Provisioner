/*
Copyright 2016 The Kubernetes Authors.

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

package main

import (
	"errors"
	"flag"
	"os"
	"path"
	"time"
	"fmt"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"syscall"
)

const (
	resyncPeriod              = 15 * time.Second
	//
	// The provisionerName has to match the "provisioner" field of the
	// Kubernetes StorageClass object
	//
	provisionerName           = "itix.fr/hostpath"
	exponentialBackOffOnError = false
	failedRetryThreshold      = 5
	leasePeriod               = controller.DefaultLeaseDuration
	retryPeriod               = controller.DefaultRetryPeriod
	renewDeadline             = controller.DefaultRenewDeadline
	termLimit                 = controller.DefaultTermLimit
)

type hostPathProvisioner struct {
	// The directory to create PV-backing directories in
	pvDir string

	// Identity of this hostPathProvisioner, set to node's name. Used to identify
	// "this" provisioner's PVs.
	identity string
}

func NewHostPathProvisioner() controller.Provisioner {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		glog.Fatal("env variable NODE_NAME must be set so that this provisioner can identify itself")
	}
	hostPath := os.Getenv("HOSTPATH_TO_USE")
	if hostPath == "" {
		glog.Fatal("env variable HOSTPATH_TO_USE must be set")
	}
	return &hostPathProvisioner{
		pvDir:    hostPath,
		identity: nodeName,
	}
}

var _ controller.Provisioner = &hostPathProvisioner{}

func (p *hostPathProvisioner) generatePVPath(options controller.VolumeOptions) (string, error) {
	// Default value for Name Generation
	namespace := "_"
	name := options.PVName

	// Try to get information from PVC
	if pvc := options.PVC; pvc != nil {
		// Get PVC namespace if it exists
		ns := pvc.Namespace
		if ns != "" {
			namespace = ns
		}

		// Get PVC name if it exists
		n := pvc.Name
		if n != "" {
			name = n
		}
	}

	// Try to create namespace dir if it does not exist
	nspath := path.Join(p.pvDir, namespace)
	if _, err := os.Stat(nspath); os.IsNotExist(err) {
		if err := os.MkdirAll(nspath, 0777); err != nil {
			return "", err
		}
	}

	// Check if pvc name already exists
	pvpath := path.Join(nspath, name)
	if _, err := os.Stat(pvpath); err == nil {
		// If yes, try to generate a new name
		for i := 1; i < 100; i++ {
			 new_name := fmt.Sprintf("%s-%02d", name, i)
			 new_pvpath := path.Join(nspath, new_name)
			 if _, err := os.Stat(new_pvpath); os.IsNotExist(err) {
				 // Found a free name
				 name = new_name
				 pvpath = new_pvpath
				 return pvpath, nil
			 }
		}
	}

	return pvpath, nil
}

// Provision creates a storage asset and returns a PV object representing it.
func (p *hostPathProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	path, err := p.generatePVPath(options)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				"hostPathProvisionerIdentity": p.identity,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: path,
				},
			},
		},
	}

	return pv, nil
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *hostPathProvisioner) Delete(volume *v1.PersistentVolume) error {
	ann, ok := volume.Annotations["hostPathProvisionerIdentity"]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}
	if ann != p.identity {
		return &controller.IgnoredError{"identity annotation on PV does not match ours"}
	}


	// Not for the moment, please !
	//path := path.Join(p.pvDir, volume.Name)
	//if err := os.RemoveAll(path); err != nil {
	//	return err
	//}

	return nil
}

func main() {
	syscall.Umask(0)

	flag.Parse()
	flag.Set("logtostderr", "true")

	// Create an InClusterConfig and use it to create a client for the controller
	// to use to communicate with Kubernetes
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create client: %v", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Fatalf("Error getting server version: %v", err)
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	hostPathProvisioner := NewHostPathProvisioner()

	// Start the provision controller which will dynamically provision hostPath
	// PVs
	pc := controller.NewProvisionController(clientset, resyncPeriod, provisionerName, hostPathProvisioner, serverVersion.GitVersion, exponentialBackOffOnError, failedRetryThreshold, leasePeriod, renewDeadline, retryPeriod, termLimit)
	pc.Run(wait.NeverStop)
}
