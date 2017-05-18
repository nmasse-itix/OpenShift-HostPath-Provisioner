oc project default
oc create -f service-account.yaml
oc create -f roles.yaml
oc adm policy add-scc-to-user hostmount-anyuid -z hostpath-provisioner
oc adm policy add-cluster-role-to-user hostpath-provisioner -z hostpath-provisioner

mkdir /tmp/openshift
chmod 777 /tmp/openshift
chcon -Rt svirt_sandbox_file_t /tmp/openshift

oc create -f pod.yaml
oc create -f class.yaml

oc create -f sample-claim.yaml
