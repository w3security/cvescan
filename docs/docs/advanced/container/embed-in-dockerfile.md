# Embed in Dockerfile

Scan your image as part of the build process by embedding cvescan in the
Dockerfile. This approach can be used to update Dockerfiles currently using
Aqua’s [Microscanner][microscanner].

```bash
$ cat Dockerfile
FROM alpine:3.7

RUN apk add curl \
    && curl -sfL https://raw.githubusercontent.com/w3security/cvescan/main/contrib/install.sh | sh -s -- -b /usr/local/bin \
    && cvescan rootfs --exit-code 1 --no-progress /

$ docker build -t vulnerable-image .
```
Alternatively you can use cvescan in a multistage build. Thus avoiding the
insecure `curl | sh`. Also the image is not changed.
```bash
[...]
# Run vulnerability scan on build image
FROM build AS vulnscan
COPY --from=w3security/cvescan:latest /usr/local/bin/cvescan /usr/local/bin/trivy
RUN cvescan rootfs --exit-code 1 --no-progress /
[...]
```

[microscanner]: https://github.com/aquasecurity/microscanner
