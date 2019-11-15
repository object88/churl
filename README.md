# churl

`churl` (**ch**art museum c**url**) is a utility for interacting with a [Helm](https://helm.sh) [chart museum]().

The default chart museum deployment does not include a kubernetes ingress, meaning that it's only accessible from within the kubernetes cluster, perhaps using it's service address (i.e., http://cm-chartmuseum.chartmuseum:8080/index.yaml).  `churl` will create a kubernetes port-forwarding tunnel from a local machine and issue HTTP requests to retrieve charts and chart archives.

## Limits

`churl` is _not_ meant to add, remove, or alter charts or chart archives.  It's recommended to use helm commands for that purpose.

## Goals

This project has several goals.  First and foremost is to create a tool that helps me access a protected chart museum.  There are several secondary goals:

### Branching strategy

In order to better understand the OneFlow branching stragey, it will be applied here.  Reading about it does not impart the same level of undestanding as using it.

### Github tooling

Github has made a number of CI/CD tools available to open source developers.

### Over-engineering

`churl` started as a bash script.  The following is a much simplified version of the script.

``` sh
#!/usr/bin/env bash

QUERY=$1
shift 1

# Set up the port forward to run in the background, but grab the PID so that we
# can terminate the forwarding when we're done.
kubectl port-forward svc/cm-chartmuseum 9000:8080 --address localhost --context "minikube" --namespace chartmuseum > /dev/null 2>&1 &
KUBEPID=$!

# Ensure that the port-forwarding process dies with this script.
closePortForward() {
  kill $KUBEPID
}

trap closePortForward EXIT

sleep 10

curl -sL "http://localhost:9000/$QUERY"
```

## Future

The port forward currently lives as long as the `churl` executable, however it is probably common to perform multiple requests.  An improvement might be to spawn a long-lived background process that manages the port-forward, and closes after a period of inactivity.  In this arrangement, the user's request is routed to the background process, which performs the actual request.

## Testing

`churl` is tested with a local kubernetes & helm installation:

``` sh
helm repo add stable https://kubernetes-charts.storage.googleapis.com
helm install stable/chartmuseum --name cm
```