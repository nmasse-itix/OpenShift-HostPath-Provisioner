# An Hostpath Provisioner for OpenShift

## If you just want to use it

### Setup

```
$ oc project default
$ oc process -f setup/hostpath-provisioner-template.yaml | oc create -f -
```

### Test

```
$ oc project my-project
$ oc create -f setup/sample-claim.yaml
$ ls -l /tmp/openshift/
```

### Cleanup

```
$ oc project default
$ oc delete all -l template=hostpath-provisioner-template
```

## If you want to hack it

### Build

```
$ export GOPATH="$PWD"
$ cd src
$ glide install -v
$ CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ../hostpath-provisioner hostpath-provisioner/hostpath-provisioner.go
```

### Package

```
$ docker build -t openshift-hostpath-provisioner:latest .
```

### Setup

```
$ oc project default
$ oc process -f setup/hostpath-provisioner-template.yaml -p HOSTPATH_PROVISIONER_IMAGE=openshift-hostpath-provisioner:latest | oc create -f -
```

### Test

```
$ oc project my-project
$ oc create -f setup/sample-claim.yaml
$ ls -l /tmp/openshift/
```

### Cleanup

```
$ oc project default
$ oc delete all -l template=hostpath-provisioner-template
```

### Pushing your image to DockerHub (Optional)

```
$ docker login https://index.docker.io/v1/
$ docker images openshift-hostpath-provisioner:latest --format '{{ .ID }}'
$ docker tag $(docker images openshift-hostpath-provisioner:latest --format '{{ .ID }}') index.docker.io/<your-username>/openshift-hostpath-provisioner
$ docker push index.docker.io/<your-username>/openshift-hostpath-provisioner
```
