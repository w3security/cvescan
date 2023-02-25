Docker Hub needs `TRIVY_USERNAME` and `TRIVY_PASSWORD`.
You don't need to set ENV vars when download from public repository.

```bash
export cvescan_USERNAME={DOCKERHUB_USERNAME}
export cvescan_PASSWORD={DOCKERHUB_PASSWORD}
```
