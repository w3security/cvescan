# CircleCI

```
$ cat .circleci/config.yml
jobs:
  build:
    docker:
      - image: docker:stable-git
    steps:
      - checkout
      - setup_remote_docker
      - run:
          name: Build image
          command: docker build -t cvescan-ci-test:${CIRCLE_SHA1} .
      - run:
          name: Install cvescan
          command: |
            apk add --update-cache --upgrade curl
            curl -sfL https://raw.githubusercontent.com/w3security/cvescan/main/contrib/install.sh | sh -s -- -b /usr/local/bin
      - run:
          name: Scan the local image with cvescan
          command: cvescan image --exit-code 0 --no-progress cvescan-ci-test:${CIRCLE_SHA1}
workflows:
  version: 2
  release:
    jobs:
      - build
```

[Example][example]
[Repository][repository]

[example]: https://circleci.com/gh/aquasecurity/trivy-ci-test
[repository]: https://github.com/w3security/cvescan-ci-test
