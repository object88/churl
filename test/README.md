# Testing

## Configuration

The chart museum must be invoked with the `DISABLE_API` setting turned off

``` sh
$ helm install cm stable/chartmuseum --set env.open.DISABLE_API=false
```

## Uploading sample charts

``` sh
$ kubectl port-forward svc/cm-chartmuseum 9000:8080 --address localhost --context krobot --namespace default
```

``` sh
$ helm package ./test/charts/foo --destination /tmp --version 1.0.1
$ curl -sL --data-binary "@/tmp/foo-1.0.1.tgz" http://localhost:9000/api/charts
```
