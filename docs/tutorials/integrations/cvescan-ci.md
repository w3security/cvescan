# Travis CI

```
$ cat .travis.yml
services:
  - docker

env:
  global:
    - COMMIT=${TRAVIS_COMMIT::8}

before_install:
  - docker build -t cvescan-ci-test:${COMMIT} .
  - export VERSION=$(curl --silent "https://api.github.com/repos/w3security/cvescan/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
  - wget https://github.com/w3security/cvescan/releases/download/v${VERSION}/trivy_${VERSION}_Linux-64bit.tar.gz
  - tar zxvf cvescan_${VERSION}_Linux-64bit.tar.gz
script:
  - ./cvescan image --exit-code 0 --severity HIGH --no-progress cvescan-ci-test:${COMMIT}
  - ./cvescan image --exit-code 1 --severity CRITICAL --no-progress cvescan-ci-test:${COMMIT}
cache:
  directories:
    - $HOME/.cache/trivy
```

[Example][example]
[Repository][repository]

[example]: https://travis-ci.org/aquasecurity/trivy-ci-test
[repository]: https://github.com/w3security/cvescan-ci-test
