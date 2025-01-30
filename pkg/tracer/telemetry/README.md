# install jaeger in kubernetes

```bash
helm repo add jaeger-all-in-one https://raw.githubusercontent.com/hansehe/jaeger-all-in-one/master/helm/charts
helm repo update
helm install jaeger-all-in-one jaeger-all-in-one/jaeger-all-in-one
```

# access jaeger

```bash
export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/name=jaeger-all-in-one,app.kubernetes.io/instance=jaeger-all-in-one" -o jsonpath="{.items[0].metadata.name}")
echo "Visit http://127.0.0.1:16686 to use your application"
kubectl --namespace default port-forward $POD_NAME 16686:16686
```