# An Hostpath Provisioner for OpenShift

This project solves the PersistentVolume provisioning in OpenShift. It is mainly
a packaged and polished version of the Kubernetes Incubator's Hostpath Provisioner
(https://github.com/kubernetes-incubator/external-storage/tree/master/docs/demo/hostpath-provisioner).

The original license is Apache 2.0. This project remains under the same license.

__Main differences from the original project :__
 - OpenShift Template that works Out-of-the-Box
 - Naming convention (namespace/pvc-name instead of GUID)
 - Configurable root directory (HOSTPATH_TO_USE parameter)
 - Docker Image available on DockerHub (https://hub.docker.com/r/nmasse/openshift-hostpath-provisioner/)

__Current limitations :__
 - **DO NOT USE IT IN PRODUCTION**
 - Only works on OpenShift clusters that have a single node

## If you just want to use it

If you just want to use this project, without having to compile or debug,
here is the way to go.

### Setup

First, you will have to create a directory that will hold all PersistentVolumes :
```
mkdir /var/openshift
chmod 777 /var/openshift
chcon -Rt svirt_sandbox_file_t /var/openshift
```

If you chose a different path, mind that you will have to pass the HOSTPATH_TO_USE
parameter to the OpenShift template (-p HOSTPATH_TO_USE=/path/to/other/directory).

In order to setup the hostpath provisioner, you have to be cluster admin on your
OpenShift instance. There are multiple ways to login with cluster admin rights,
one way is to use your kube.config file on the master :
```
oc login -u system:admin --config ~/.kube/config
```

Then, you will have to process the template and create the generated objects :
```
oc process -f setup/hostpath-provisioner-template.yaml > objects.json
oc create -f objects.json
```

If you chose a different path, just pass the HOSTPATH_TO_USE parameter to the
"oc process" command :
```
oc process -f setup/hostpath-provisioner-template.yaml -p HOSTPATH_TO_USE=/path/to/other/directory > objects.json
oc create -f objects.json
```

The template by default creates objects in the "default" namespace, no matter
the "-n" option you used or the "oc project" command you issued before. This is
in fact a limitation (see https://github.com/openshift/origin/issues/8971).

But, if you want to generate objects in another namespace (project), just pass the
TARGET_NAMESPACE parameter to the "oc process" command :
```
oc process -f setup/hostpath-provisioner-template.yaml -p TARGET_NAMESPACE=my-project > objects.json
oc create -f objects.json
```

By default, the template uses the "nmasse/openshift-hostpath-provisioner:latest"
image on DockerHub (see https://hub.docker.com/r/nmasse/openshift-hostpath-provisioner/).
You can change this behavior by passing the
HOSTPATH_PROVISIONER_IMAGE to "oc process" command :

```
oc process -f setup/hostpath-provisioner-template.yaml -p HOSTPATH_PROVISIONER_IMAGE=john/openshift-hostpath-provisioner:latest > objects.json
oc create -f objects.json
```

If you need to pass multiple parameters, use multiple "-p" options :
```
oc process -f setup/hostpath-provisioner-template.yaml -p TARGET_NAMESPACE=my-project -p HOSTPATH_TO_USE=/path/to/other/directory > objects.json
oc create -f objects.json
```

__Note about the file 'objects.json' :__

It's a good idea to keep that file safe since it will be use to clean up your OpenShift
instance in case you change your mind. See the cleanup section below.

### Test

A sample PersistentVolumeClaim object is provided to test the provisioner :

```
oc project my-project
oc create -f setup/sample-claim.yaml
find /var/openshift/
```

### Cleanup

You can cleanup your OpenShift instance by running the "oc delete" command :

```
oc delete -f objects.json
```

## If you want to hack it

### Build

```
export GOPATH="$PWD"
cd src
glide install -v
CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ../hostpath-provisioner hostpath-provisioner/hostpath-provisioner.go
```

### Package

```
docker build -t openshift-hostpath-provisioner:latest .
```

### Manual Cleanup

```
oc project default
oc delete all -l template=hostpath-provisioner-template
oc delete sa hostpath-provisioner
oc delete clusterrolebinding hostpath-provisioner
oc delete clusterrole hostpath-provisioner
oc delete scc hostpath-provisioner
oc delete storageclass hostpath-provisioner
```

### Pushing your image to DockerHub (Optional)

```
docker login https://index.docker.io/v1/
docker images openshift-hostpath-provisioner:latest --format '{{ .ID }}'
docker tag $(docker images openshift-hostpath-provisioner:latest --format '{{ .ID }}') index.docker.io/<your-username>/openshift-hostpath-provisioner
docker push index.docker.io/<your-username>/openshift-hostpath-provisioner
```
