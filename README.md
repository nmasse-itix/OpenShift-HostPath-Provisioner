# An Hostpath Provisioner for OpenShift

## Build

```
$ export GOPATH="$PWD"
$ cd src
$ glide install -v
$ CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ../hostpath-provisioner hostpath-provisioner/hostpath-provisioner.go
```

## Package

```
$ docker build -t hostpath-provisioner .
```

## Setup

```
$ oc project default
$ oc process -f setup/hostpath-provisioner-template.yaml
```

## Test

```
$ oc project my-project
$ oc create -f setup/sample-claim.yaml
$ ls -l /tmp/openshift/
```

## Cleanup

```
$ oc project default
$ oc delete all -l template=hostpath-provisioner-template
```
