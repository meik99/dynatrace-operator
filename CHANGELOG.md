### Future
#### Features
* Updated CRD version from `v1alpha1` to `v1beta1`
  * Improved structure of the Custom Resource to simplify usage
  * Added Conversion Webhook to migrate existing Dynakubes to the new version
  * Added `ValidatingAdmissionWebhook` to verify correctness of newly applied Dynakubes
* Added new monitoring modes:
  *  (PREVIEW with `useCSIDriver: true`) `applicationMonitoring`: webhook based injection mechanism for automatic-app-only injection
    * Added auto-detection of c runtime flavor
  * `hostMonitoring`: only monitoring the host in the cluster without app-only injection
  * (PREVIEW) `cloudNativeFullStack`: combination of `hostMonitoring` and `applicationMonitoring`
  * existing monitoring mode, i.e. `classicFullStack`, will stay in place
* Improved ActiveGate support:
  * Allow multiple capabilities within one ActiveGate pod
  * Added option to define custom certificate (#293)
  * Added support for self-signed K8S API certificates (#102)
* Added new way of monitoring applications in different namespaces
  * Per default, all namespaces (excluding system namespaces) are monitored
  * Added configuration via `namespaceSelector` to limit to namespaces with specific labels
* (PREVIEW) Added CSI Driver for `applicationMonitoring` and `cloudNativeFullStack`
  * Minimizes binary downloads (once per node, instead of once per pod)
  * Can be enabled for `applicationMonitoring` via `useCSIDriver: true`
  * Mandatory for `cloudNativeFullStack`
* Added script to troubleshoot common installation issues
​
#### Bug fixes
* Removed Beta affinity for Kubernetes >= 1.14
​
#### Other changes
* Changed `imagePullPolicy` for `initContainer` to `IfNotPresent` (#299)
* Changed deployment strategy for OneAgent Daemonset to `RollingUpdate`
​
#### Upgrading
The Operator can be upgraded from `v0.2.2` with,

```sh
# Kubernetes
$ kubectl apply -f https://github.com/Dynatrace/dynatrace-operator/releases/download/v0.3.0/kubernetes.yaml
​
# Openshift
$ oc apply -f https://github.com/Dynatrace/dynatrace-operator/releases/download/v0.3.0/openshift.yaml
```

If you want to use the PREVIEW features described above, install/upgrade via the following commands:
```sh
# Kubernetes
$ kubectl apply -f https://github.com/Dynatrace/dynatrace-operator/releases/download/v0.3.0/kubernetes-csi.yaml
​
# Openshift
$ oc apply -f https://github.com/Dynatrace/dynatrace-operator/releases/download/v0.3.0/openshift-csi.yaml
```