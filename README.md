# Kubernetes Custom Resource Definition (CRD) Example

This is an example how to use CRD in order to operate secrets with Sops inside Kubernetes.

## Setup

Building the example:

    $ go get github.com/harbur/sops-operator

Setting up a custom resource definition (CRD) with an example object:

    $ kubectl apply -f https://raw.githubusercontent.com/harbur/sops-operator/master/kubernetes/crd.yaml
    $ kubectl apply -f https://github.com/harbur/sops-operator/blob/master/kubernetes/sealedsecret.yaml