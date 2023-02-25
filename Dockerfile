FROM alpine:3.17.2
RUN apk --no-cache add ca-certificates git
COPY cvescan /usr/local/bin/trivy
COPY contrib/*.tpl contrib/
ENTRYPOINT ["trivy"]
