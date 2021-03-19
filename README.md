# pod_tcpstate_exporter
Exports tcp stats metric for k8s pod

# deploy
Specify namespaces that collect pod tcp  metrics under
```yaml
          env:
            - name: NAMESPACES
              # namespaces, split ,support all
              # value: all
              value: daily,tag
```

```bash
kubectl apply -f deploy/deployment.yaml
```