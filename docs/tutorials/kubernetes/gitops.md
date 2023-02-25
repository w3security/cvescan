# Installing the cvescan-Operator through GitOps

This tutorial shows you how to install the cvescan Operator through GitOps platforms, namely ArgoCD and FluxCD.

## ArgoCD

Make sure to have [ArgoCD installed](https://argo-cd.readthedocs.io/en/stable/getting_started/) and running in your Kubernetes cluster.

You can either deploy the cvescan Operator through the argocd CLI or by applying a Kubernetes manifest.

ArgoCD command:
```
> kubectl create ns cvescan-system
> argocd app create cvescan-operator --repo https://github.com/w3security/cvescan-operator --path deploy/helm --dest-server https://kubernetes.default.svc --dest-namespace cvescan-system
```
Note that this installation is directly related to our official Helm Chart. If you want to change any of the value, we'd suggest you to create a separate values.yaml file.

Kubernetes manifest `trivy-operator.yaml`:
```
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: cvescan-operator
  namespace: argocd
spec:
  project: default
  source:
    chart: cvescan-operator
    repoURL: https://w3security.github.io/helm-charts/
    targetRevision: 0.0.3
    helm:
      values: |
        cvescan:
          ignoreUnfixed: true
  destination:
    server: https://kubernetes.default.svc
    namespace: cvescan-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

The apply the Kubernetes manifest. If you have the manifest locally, you can use the following command through kubectl:
```
> kubectl apply -f cvescan-operator.yaml

application.argoproj.io/trivy-operator created
```

If you have the manifest in a Git repository, you can apply it to your cluster through the following command:
```
> kubectl apply -n argocd -f https://raw.githubusercontent.com/AnaisUrlichs/argocd-starboard/main/starboard/argocd-starboard.yaml
```
The latter command would allow you to make changes to the YAML manifest that ArgoCD would register automatically.

Once deployed, you want to tell ArgoCD to sync the application from the actual state to the desired state:
```
argocd app sync cvescan-operator
```

Now you can see the deployment in the ArgoCD UI. Have a look at the ArgoCD documentation to know how to access the UI.

![ArgoCD UI after deploying the cvescan Operator](../../imgs/argocd-ui.png)

Note that ArgoCD is unable to show the cvescan CRDs as synced.


## FluxCD

Make sure to have [FluxCD installed](https://fluxcd.io/docs/installation/#install-the-flux-cli) and running in your Kubernetes cluster.

You can either deploy the cvescan Operator through the Flux CLI or by applying a Kubernetes manifest.

Flux command:
```
> kubectl create ns cvescan-system
> flux create source helm cvescan-operator --url https://w3security.github.io/helm-charts --namespace cvescan-system
> flux create helmrelease cvescan-operator --chart cvescan-operator
  --source HelmRepository/trivy-operator
  --chart-version 0.0.3
  --namespace cvescan-system
```

Kubernetes manifest `trivy-operator.yaml`:
```
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: cvescan-operator
  namespace: flux-system
spec:
  interval: 60m
  url: https://w3security.github.io/helm-charts/

---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: cvescan-operator
  namespace: cvescan-system
spec:
  chart:
    spec:
      chart: cvescan-operator
      sourceRef:
        kind: HelmRepository
        name: cvescan-operator
        namespace: flux-system
      version: 0.10.1
  interval: 60m
  values:
    cvescan:
      ignoreUnfixed: true
  install:
    crds: CreateReplace
    createNamespace: true
```

You can then apply the file to your Kubernetes cluster:
```
kubectl apply -f cvescan-operator.yaml
```

## After the installation

After the install, you want to check that the cvescan operator is running in the cvescan-system namespace:
```
kubectl get deployment -n cvescan-system
```

