oc create -f service-account.yaml
oc create -f roles.yaml
# oc adm policy add-scc-to-user hostmount-anyuid system:serviceaccount:test-provisioner:hostpath-provisioner
# oc adm policy add-cluster-role-to-user hostpath-provisioner-runner system:serviceaccount:test-provisioner:hostpath-provisioner

oc create -f pod.yaml
oc create -f class.yaml
