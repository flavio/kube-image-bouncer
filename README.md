# kube-image-bouncer

A simple [Kubernetes admission controller](https://kubernetes.io/docs/admin/admission-controllers/)
implementing the
[ImagePolicyWebhook](https://kubernetes.io/docs/admin/admission-controllers/#imagepolicywebhook)
model.

This admission control will reject all the pods that are using images with the
`latest` tag.

## Building

To build the project just do:

```
go get github.com/flavio/kube-image-bouncer
```

The project dependencies are tracked inside of this repository and are managed
using [dep](https://github.com/golang/dep).

## Deployment

### Kubernetes master node(s)

Ensure the `ImagePolicyWebhook` admission controller is enabled. Refer to
the [official](https://kubernetes.io/docs/admin/admission-controllers/#imagepolicywebhook)
documentation.

Create a admission control configuration file. For example create
a `/etc/kubernetes/admission_configuration.json` file with the following
contents:

```json
{
  "imagePolicy": {
     "kubeConfigFile": "/etc/kubernetes/kube-image-bouncer.yml",
     "allowTTL": 50,
     "denyTTL": 50,
     "retryBackoff": 500,
     "defaultAllow": false
  }
}
```

*Note well:* this configuration file will automatically reject all the images if the server
referenced by the webhook configuration is not reachable.


Create a kubeconfig file `/etc/kubernetes/kube-image-bouncer.yml` with the
following contents:

```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority: /etc/kubernetes/kube-image-bouncer/server.cert
    server: https://bouncer.local.lan:1323/policy
  name: bouncer_webhook
contexts:
- context:
    cluster: bouncer_webhook
    user: api-server
  name: bouncer_validator
current-context: bouncer_validator
preferences: {}
users:
- name: api-server
  user:
    client-certificate: /etc/kubernetes/kube-image-bouncer/user.cert
    client-key:  /etc/kubernetes/kube-image-bouncer/user.key
```

This configuration file instructs the API server to reach the webhook server
at `https://bouncer.local.lan:1323` and use its `/policy` endpoint.

The remote server is using the certificate found inside of
`/etc/kubernetes/kube-image--bouncer/server.cert`. The certificate can be
a self-signed one, it doesn't matter.

The kubeconfig references also a client certificate and key. These can be generated
by doing:

```
openssl req  -nodes -new -x509  -keyout user.key -out user.cert
```

### Node running the actual web application

Create a server key and certificate with the following command:

```
openssl req  -nodes -new -x509  -keyout server.key -out server.cert
```

Copy the `server.cert` file to the kubernetes manager node(s) as
`/etc/kubernetes/k8s-bouncer/server.cert`. Ensure you restart the
`kube-apiserver` on all the master node(s).

Start the web application by doing:
```
kube-image-bouncer -cert server.cert -key server.key
```

## Profit!

Create a `nginx-versioned.yml` file:

```yml
apiVersion: v1
kind: ReplicationController
metadata:
  name: nginx-versioned
spec:
  replicas: 1
  selector:
    app: nginx-versioned
  template:
    metadata:
      name: nginx-versioned
      labels:
        app: nginx-versioned
    spec:
      containers:
      - name: nginx-versioned
        image: nginx:1.13.8
        ports:
        - containerPort: 80
```

Then create the resource:

```
kubectl create -f nginx-versioned.yml
```
Ensure the replication controller is actually running:

```
$ kubectl get rc
NAME              DESIRED   CURRENT   READY     AGE
nginx-versioned   1         1         0         2h
```


Now create a `nginx-latest.yml` file:

```yml
apiVersion: v1
kind: ReplicationController
metadata:
  name: nginx-latest
spec:
  replicas: 1
  selector:
    app: nginx-latest
  template:
    metadata:
      name: nginx-latest
      labels:
        app: nginx-latest
    spec:
      containers:
      - name: nginx-latest
        image: nginx
        ports:
        - containerPort: 80
```

Then create the resource:

```
kubectl create -f nginx-latest.yml
```

This time the replication controller won't have all the desired pods running:

```
$ kubectl get rc
NAME              DESIRED   CURRENT   READY     AGE
nginx-latest      1         0         0         4s
nginx-versioned   1         1         0         2h
```

Get more details about the `nginx-versioned` replication controller:

```
$ kubectl describe rc nginx-latest
Name:         nginx-latest
Namespace:    default
Selector:     app=nginx-latest
Labels:       app=nginx-latest
Annotations:  <none>
Replicas:     0 current / 1 desired
Pods Status:  0 Running / 0 Waiting / 0 Succeeded / 0 Failed
Pod Template:
  Labels:  app=nginx-latest
  Containers:
   nginx-latest:
    Image:        nginx
    Port:         80/TCP
    Environment:  <none>
    Mounts:       <none>
  Volumes:        <none>
Conditions:
  Type             Status  Reason
  ----             ------  ------
  ReplicaFailure   True    FailedCreate
Events:
  Type     Reason        Age                From                    Message
  ----     ------        ----               ----                    -------
  Warning  FailedCreate  22s (x14 over 1m)  replication-controller  Error creating: pods "nginx-latest-" is forbidden: image policy webhook backend denied one or more images: Images using latest tag are not allowed

```

The culprit is inside of the latest line of the output, the pod creation has
been forbidden by our admission controller with the following message:

> Images using latest tag are not allowed

# Caveats

The admission controller is used to vet **all** the containers scheduled to run
inside of the cluster. That includes containers providing core services like
kube-dns, dex, kubedash,... If the image bouncer service is unreachable these
services won't be accepted inside of the cluster (because we set `defaultAllow` to
`false` inside of `/etc/kubernetes/admission_configuration.json`).

We could run the image bouncer on top of the kubernetes cluster, but that
would require its container to be accepted into the cluster, which leads to
a *"chicken-egg"* problem.
