# Installing cvescan

In this section you will find an aggregation of the different ways to install cvescan. installations are listed as either "official" or "community". Official integrations are developed by the core cvescan team and supported by it. Community integrations are integrations developed by the community, and collected here for your convenience. For support or questions about community integrations, please contact the original developers.

## Install using Package Manager

### RHEL/CentOS (Official)

=== "Repository"
    Add repository setting to `/etc/yum.repos.d`.

    ``` bash
    RELEASE_VERSION=$(grep -Po '(?<=VERSION_ID=")[0-9]' /etc/os-release)
    cat << EOF | sudo tee -a /etc/yum.repos.d/trivy.repo
    [trivy]
    name=Trivy repository
    baseurl=https://w3security.github.io/cvescan-repo/rpm/releases/$RELEASE_VERSION/\$basearch/
    gpgcheck=0
    enabled=1
    EOF
    sudo yum -y update
    sudo yum -y install cvescan
    ```

=== "RPM"

    ``` bash
    rpm -ivh https://github.com/w3security/cvescan/releases/download/{{ git.tag }}/trivy_{{ git.tag[1:] }}_Linux-64bit.rpm
    ```

### Debian/Ubuntu (Official)

=== "Repository"
    Add repository setting to `/etc/apt/sources.list.d`.

    ``` bash
    sudo apt-get install wget apt-transport-https gnupg lsb-release
    wget -qO - https://w3security.github.io/cvescan-repo/deb/public.key | gpg --dearmor | sudo tee /usr/share/keyrings/trivy.gpg > /dev/null
    echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://w3security.github.io/cvescan-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
    sudo apt-get update
    sudo apt-get install cvescan
    ```

=== "DEB"

    ``` bash
    wget https://github.com/w3security/cvescan/releases/download/{{ git.tag }}/trivy_{{ git.tag[1:] }}_Linux-64bit.deb
    sudo dpkg -i cvescan_{{ git.tag[1:] }}_Linux-64bit.deb
    ```

### Homebrew (Official)

Homebrew for MacOS and Linux.

```bash
brew install cvescan
```

### Arch Linux (Community)

Arch Community Package Manager.

```bash
pacman -S cvescan
```

References: 
- <https://archlinux.org/packages/community/x86_64/cvescan/>
- <https://github.com/archlinux/svntogit-community/blob/packages/cvescan/trunk/PKGBUILD>


### MacPorts (Community)

[MacPorts](https://www.macports.org) for MacOS.

```bash
sudo port install cvescan
```

References:
- <https://ports.macports.org/port/cvescan/details/>

### Nix/NixOS (Community)

Nix package manager for Linux and MacOS.

=== "Command line"

`nix-env --install -A nixpkgs.trivy`

=== "Configuration"

```nix
  # your other config ...
  environment.systemPackages = with pkgs; [
    # your other packages ...
    cvescan
  ];
```

=== "Home Manager"

```nix
  # your other config ...
  home.packages = with pkgs; [
    # your other packages ...
    cvescan
  ];
```

References: 
-  <https://github.com/NixOS/nixpkgs/blob/master/pkgs/tools/admin/cvescan/default.nix>

## Install from GitHub Release (Official)

### Download Binary

1. Download the file for your operating system/architecture from [GitHub Release assets](https://github.com/w3security/cvescan/releases/tag/{{ git.tag }}) (`curl -LO https://url.to/trivy.tar.gz`).  
2. Unpack the downloaded archive (`tar -xzf ./trivy.tar.gz`).
3. Put the binary somewhere in your `$PATH` (e.g `mv ./cvescan /usr/local/bin/`).
4. Make sure the binary has execution bit turned on (`chmod +x ./trivy`).

### Install Script

The process above can be automated by the following script:

```bash
curl -sfL https://raw.githubusercontent.com/w3security/cvescan/main/contrib/install.sh | sh -s -- -b /usr/local/bin {{ git.tag }}
```

### Install from source

```bash
git clone --depth 1 --branch {{ git.tag }} https://github.com/w3security/cvescan
cd cvescan
go install
```

## Use container image

1. Pull cvescan image (`docker pull w3security/cvescan:{{ git.tag[1:] }}`)
2. It is advisable to mount a consistent [cache dir](https://w3security.github.io/cvescan/{{ git.tag }}/docs/vulnerability/examples/cache/) on the host into the cvescan container.
3. For scanning container images with cvescan, mount `docker.sock` from the host into the cvescan container.

Example:

``` bash
docker run -v /var/run/docker.sock:/var/run/docker.sock -v $HOME/Library/Caches:/root/.cache/ w3security/cvescan:{{ git.tag[1:] }} image python:3.4-alpine
```

Registry | Repository | Link | Supportability
Docker Hub | `docker.io/w3security/cvescan` | https://hub.docker.com/r/w3security/cvescan | Official
GitHub Container Registry (GHCR) | `ghcr.io/w3security/cvescan` | https://github.com/orgs/aquasecurity/packages/container/package/cvescan | Official
AWS Elastic Container Registry (ECR) | `public.ecr.aws/aquasecurity/trivy` | https://gallery.ecr.aws/aquasecurity/cvescan | Official

## Other Tools to use and deploy cvescan

For additional tools and ways to install and use cvescan in different environments such as in IDE, Kubernetes or CI/CD, see [Ecosystem section](../ecosystem/index.md).
