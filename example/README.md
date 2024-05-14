# Simple Migration Example

```console
$ kubectl apply -f service.yaml
$ migrate --migrations-dir ./migrations --direction up
$ kubectl get service/test-service | grep targetPort
    targetPort: 9371
$ migrate --migrations-dir ./migrations --direction down
$ kubectl get service/test-service | grep targetPort
    targetPort: 9376
```
