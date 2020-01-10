# kube-image-bouncer

A simple webhook endpoint server that can be used to validate the images being
created inside of the kubernetes cluster.

It works with two different types of [Kubernetes admission controller](https://kubernetes.io/docs/admin/admission-controllers/):

  * [ImagePolicyWebhook](https://kubernetes.io/docs/admin/admission-controllers/#imagepolicywebhook)
  * [GenericAdmissionWebhook](https://v1-8.docs.kubernetes.io/docs/admin/admission-controllers/#genericadmissionwebhook-alpha) (which starting from Kubernetes 1.9 has been renamed
[ValidatingAdmissionWebhook](https://kubernetes.io/docs/admin/admission-controllers/#validatingadmissionwebhook-alpha-in-18-beta-in-19).

This admission controller will reject all the pods that are using images with
the `latest` tag.

## Disclaimer

I personally find the documentation of these admission controllers vague,
confusing and missing some details.

I found the documentation about `GenericAdmissionWebhook` and
`ValidatingAdmissionWebhook` more troublesome to understand compared to the
one of `ImagePolicyWebhook`. You have to combine the 1.9 documentation of
`ValidatingAdmissionWebhook` together with the 1.8 documentation of
`GenericAdmissionWebhook`. The latter one references the
[Dynamic Admission Control](https://v1-8.docs.kubernetes.io/docs/admin/extensible-admission-controllers/)
which has more details.

This document will try to shed some light about these admission controllers,
especially about how to deploy them.

I promise I'll try to improve upstream documentation as well :)

**Note well:** during the 1.8 -> 1.9 transition, in addition to being renamed,
the `GenericAdmissionWebhook` has also been promoted from being an `alpha1`
resource to be a `beta1` one. That caused some changes in terms of the request
format sent from the API server to the webhook endpoint and in terms of
expected response.

Right now this document (and the code) focuses on version 1.9.

# Comparison

The [ImagePolicyWebhook](https://kubernetes.io/docs/admin/admission-controllers/#imagepolicywebhook)
is an admission controller that evaluates only images.

Good things about `ImagePolicyWebhook`:

  * The API server can be instructed to reject the images if the webhook
    endpoint is not reachable.

Bad things about `ImagePolicyWebhook`:

  * More configuration files are expected on the API server node(s) compared to
    `GenericAdmissionWebhook`.
  * It's a bit tricky to deploy the service providing the webhook endpoint on the
    kubernetes cluster (more on that later).


The [GenericAdmissionWebhook](https://v1-8.docs.kubernetes.io/docs/admin/admission-controllers/#genericadmissionwebhook-alpha) (which starting from Kubernetes 1.9 has been renamed
[ValidatingAdmissionWebhook](https://kubernetes.io/docs/admin/admission-controllers/#validatingadmissionwebhook-alpha-in-18-beta-in-19)
can evaluate all kind of resources.

Good things about `ValidatingAdmissionWebhook`:

  * Only two changes are required on the kubernetes master node(s).
  * Part of the configuration is defined by a Kubernetes object.
  * It's incredibly easy to deploy the service providing the webhook endpoint
    on the kubernetes cluster.

Bad things about `ValidatingAdmissionWebhook`:

  * Starting from the 1.8 release it's no longer possible to instruct the API
    server to reject the images if the webhook endpoint is
    not reachable. Hence, when the endpoint is not reachable, all the resources
    are going to be automatically accepted.

# Building

To build the project just do:

```
$ go get github.com/flavio/kube-image-bouncer
```

The project dependencies are tracked inside of this repository and are managed
using [dep](https://github.com/golang/dep).

This application is distributed also as a [Docker image](https://hub.docker.com/r/flavio/kube-image-bouncer/):

```
$ docker pull flavio/kube-image-bouncer
```

# Deployment of `ImagePolicyWebhook`

The webhook endpoint must be secured by tls to be used by kubernetes. This
certificate can also be a self-signed one.

Create a server key and certificate with the following command:

```
$ openssl req  -nodes -new -x509 -keyout webhook-key.pem -out webhook.pem
```

The API server uses a certificate to prove its identity. This
certificate can also be a self-signed one.

Create a server key and certificate with the following command:

```
$ openssl req  -nodes -new -x509 -keyout apiserver-client-key.pem -out apiserver-client.pem
```

## Kubernetes master node(s)

Ensure the `ImagePolicyWebhook` admission controller is enabled. Refer to
the [official](https://kubernetes.io/docs/admin/admission-controllers/#imagepolicywebhook)
documentation.

Create an admission control configuration file named
`/etc/kubernetes/admission_configuration.json` file with the following
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

**Note well:** this configuration file will automatically reject all the images
if the server referenced by the webhook configuration is not reachable
(see the `defaultAllow: false` directive).

Create a kubeconfig file `/etc/kubernetes/kube-image-bouncer.yml` with the
following contents:

```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority: /etc/kubernetes/kube-image-bouncer/webhook.pem
    server: https://bouncer.local.lan:1323/image_policy
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
    client-certificate: /etc/kubernetes/kube-image-bouncer/apiserver-client.pem
    client-key:  /etc/kubernetes/kube-image-bouncer/apiserver-client-key.pem
```

This configuration file instructs the API server to reach the webhook server
at `https://bouncer.local.lan:1323` and use its `/image_policy` endpoint.

Note that the certificates and keys we previously generated have been copied
under `/etc/kubernetes/kube-image-bouncer`.

## Node running the webhook endpoint

Copy the `webhook.pem` and `webhook-key.pem` files to node running the webhook
service.

Start the webhook by doing:

```
$ kube-image-bouncer --cert webhook.pem --key webhook-key.pem
```

When using the [Docker image](https://hub.docker.com/r/flavio/kube-image-bouncer/):

```
$ docker run --rm -v `pwd`/webhook-key.pem:/certs/webhook-key.pem:ro -v `pwd`/webhook.pem:/certs/webhook.pem:ro -p 1323:1323 flavio/kube-image-bouncer -k /certs/webhook-key.pem -c /certs/webhook.pem
```

This will start a container with the server key and certificate mounted read-only
inside of it.

If you want to perform tls termination outside of this application, just start
it without providing a key and a certificate.

# Deployment of `ValidatingAdmissionWebhook`

The webhook endpoint must be secured by tls to be used by kubernetes. This
certificate can also be a self-signed one.

Create a server key and certificate with the following command:

```
$ openssl req  -nodes -new -x509 -keyout webhook-key.pem -out webhook.pem
```

The API server uses a certificate to prove its identity. This
certificate can also be a self-signed one.

Create a server key and certificate with the following command:

```
$ openssl req  -nodes -new -x509 -keyout apiserver-client-key.pem -out apiserver-client.pem
```

## Kubernetes master node(s)

Ensure the `GenericAdmissionWebhook` admission controller is enabled:
the `--admission-control` flag must mention it.

Ensure the API server is started with the following flags:
```
--proxy-client-cert-file=/etc/kubernetes/apiserver-client.pem \
--proxy-client-key-file=/etc/kubernetes/apiserver-client-key.pem"
```

## Define Kubernetes objects

First of all you have to create a tls secret holding the webhook certificate
and key:

```
kubectl create secret tls tls-image-bouncer-webhook \
  --key webhook-key.pem \
  --cert webhook.pem
```

Then create a kubernetes deployment for the `image-bouncer-webhook`:

```
kubectl apply -f kubernetes/image-bouncer-webhook.yaml
```

Finally create `ValidatingWebhookConfiguration` that makes use of
our webhook endpoint:

```
kubectl apply -f kubernetes/validating-webhook-configuration.yaml
```

**Note well:** the `ExternalAdmissionHookConfiguration` resource defined inside of
`validating-webhook-configuration.yaml` includes a CA certificate. This
is the `apiserver-client.pem` converted to base64.

As reported by the upstream docs:

> After you create the validating webhook configuration, the system will take a few seconds to honor the new configuration.

## Profit!

It doesn't matter which kind of admission controller you created, the
behaviour will be the same.

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
$ kubectl create -f nginx-versioned.yml
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
$ kubectl create -f nginx-latest.yml
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

# Caveats of `ImagePolicyWebhook`

The admission controller is used to vet **all** the containers scheduled to run
inside of the cluster. That includes containers providing core services like
kube-dns, dex, kubedash,... If the image bouncer service is unreachable these
services won't be accepted inside of the cluster (because we set `defaultAllow` to
`false` inside of `/etc/kubernetes/admission_configuration.json`).

We could run the image bouncer on top of the kubernetes cluster, but that
would require its container to be accepted into the cluster, which leads to
a *"chicken-egg"* problem.
