Kubernetes Admission Controller
=======================================================

Write a simple [Admission
Controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)
for Kubernetes that admits or rejects Pods based on their labels. The
Controller should admit and reject Pods based on the following rules:

* If a Pod has the label **team** with value **ops** it should be admitted to the
  cluster.
* If a Pod's **team** label has any other value than **ops**, it should be
  rejected.
* If the **team** label is absent from a Pod's metadata, the Pod should
  be admitted but the **team** label should be set to **ops**.

Also make sure that resources managing Pods (like Deployments) are checked for
the labels in the Pod template and admitted or rejected based on the rules
above.

Implementation
--------------

You should implement the Admission Controller in Golang. Your implementation
will be tested against [Minikube](https://github.com/kubernetes/minikube) and
has to at least correctly handle the manifests that are bundled in the
**admission/** subfolder.

Be sure to include a **Dockerfile** which can be used to build a container image
and the Kubernetes manifests necessary to deploy the admission controller into
Minikube. Also do not forget to provide instructions on how to the deploy the
Admission Controller. You can also provide a deployment script if necessary.

Minikube Parameters
-------------------

You should use Kubernetes v1.14.0 for this task to ensure that Kubernetes
admission plugins are enabled automatically.

```sh
minikube start --kubernetes-version=v1.14.0
```

TLS Certificate Generation
--------------------------

You will need to generate a TLS certificate that is trusted by the cluster as
HTTPS is enforced for Admission Controllers.

You can use the the bundled script (**scripts/webhook-create-signed-cert.sh**)
to generate the certificate and key using the cluster CA. The script will
automatically save **cert.pem** and **key.pem** in a Kubernetes secret which
you can mount into your Admission Controller Pod.

Be sure to run the script with the name of the Service for the Admission
Controller and the name of the Secret that should be used for storing the key
and certificate:

```sh
scripts/webhook-create-signed-cert.sh \
  --service <name-of-the-service> \
  --secret <name-of-the-secret>
```

You can also choose a different approach to generate and deploy the
certificates if you like.

References
----------

* [Kubernetes Admission Controller Reference](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)
* [Minikube](https://github.com/kubernetes/minikube)
