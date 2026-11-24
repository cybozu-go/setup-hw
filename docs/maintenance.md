# Maintenance

## Prerequisites

CLI tools are managed by [aqua](https://aquaproj.github.io/) in `aqua.yaml`.
Their checksums are recorded in `aqua-checksums.json` and verified on installation.
The Makefile executes the tools via `aqua exec`, which installs them automatically.

1. Install aqua. See [Install aqua](https://aquaproj.github.io/docs/install) for details.
2. Put the `aqua` command in a directory included in `PATH`.

## Cooldown period

Use versions released at least 14 days ago for external dependencies.
This avoids adopting broken or compromised releases.

Dependencies maintained by Cybozu, such as the `ghcr.io/cybozu/ubuntu` image, are not subject to the cooldown period.

## How to update dependencies

1. Update `go.mod`.
    - Update the Go version.
    - Update the modules.
    - Run `go mod tidy`.
2. Update `aqua.yaml`.
    1. Update the standard registry.
    2. Update the packages.
    3. Run `aqua update-checksum -prune` to update `aqua-checksums.json`.
3. Update GitHub Actions in `.github/workflows/*.yaml`.
    - Update the actions. They are pinned by commit SHA, so update both the SHA and the version comment.
    - Update the tools installed by the actions (e.g. `aqua_version` of `aquaproj/aqua-installer`).
    - **Do not update `runs-on` here.** See [How to update iDRAC Tools](#how-to-update-idrac-tools).
4. Update the base image in `docker/Dockerfile`.
    - Choose the latest dated tag of the same Ubuntu version (e.g. `24.04.20261117`) from [ghcr.io/cybozu/ubuntu](https://github.com/cybozu/ubuntu-base/pkgs/container/ubuntu).
    - The image is pinned by digest, so update both the tag and the digest.
    - **Do not update the Ubuntu version here.** See [How to update iDRAC Tools](#how-to-update-idrac-tools).
5. Check that the build and the tests pass.
    - Run the checks and the unit tests.
      ```console
      $ make check-generated lint test
      ```
    - Build the container image.
      ```console
      $ make build-image
      ```

## How to update iDRAC Tools

`racadm` is installed from `Dell iDRAC Tools for Linux`.

The Ubuntu version is limited to the versions supported by iDRAC Tools.
The following tree shows the items that depend on the Ubuntu version, so update them together.

```
iDRAC Tools
 └─ Ubuntu version supported by iDRAC Tools
     ├─ ghcr.io/cybozu/ubuntu image in docker/Dockerfile
     ├─ iDRACTools/racadm/UBUNTUxx directory in docker/Dockerfile
     └─ runs-on in .github/workflows/*.yaml
```

1. Find the latest version of iDRAC Tools on the Dell support site.
    - The driver ID in the URL changes for each version, so the page in the comment of `Makefile` may not be the latest one.
    - Open [Drivers & Downloads](https://www.dell.com/support/home/?app=drivers) and select a PowerEdge server model in use.
    - Filter the list with the following conditions, and find "Dell iDRAC Tools for Linux".
    - Open the details page of the latest version, and check the following.
        - The version and the release date. Compare them with the version in `Makefile`.
        - The supported Ubuntu versions in the release notes.
        - The download URL (`https://dl.dell.com/...`) and the SHA-256 checksum.
2. Update `Makefile`.
    - Update the version and the URL of the driver page in the comment.
    - Update `IDRAC_TOOLS_URL` to the download URL.
    - Update `IDRAC_TOOLS_SHA256` to the SHA-256 checksum shown on the download page.
3. If a newer Ubuntu version is supported, update the following together.
    - The `ghcr.io/cybozu/ubuntu` image in `docker/Dockerfile`.
    - The `iDRACTools/racadm/UBUNTUxx` directory in `docker/Dockerfile`.
    - `runs-on` in `.github/workflows/*.yaml`.
4. Download iDRAC Tools and extract it to `docker/iDRACTools`.
   ```console
   $ make download-idractools
   ```
5. Check that the container image can be built and `racadm` works.
   ```console
   $ make build-image
   $ docker run --rm ghcr.io/cybozu-go/setup-hw:dev racadm version
   ```
