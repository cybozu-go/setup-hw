Release procedure
=================

This document describes how to release a new version of setup-hw.

Versioning
----------

Follow [semantic versioning 2.0.0][semver] to choose the new version number.

Change log
----------

Notable changes since the last release should be listed in [CHANGELOG.md](CHANGELOG.md).

The file should respect [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format.

Bump version
------------

1. Determine a new version number.  Export it as `VERSION` environment variable:

    ```console
    $ VERSION=x.y.z
    $ export VERSION
    ```

2. Make a branch to release, for example by `git neco dev "prepare-$VERSION"`
3. Edit `CHANGELOG.md` for the new version ([example][]).
4. Commit the change and create a new pull request

    ```console
    $ git commit -a -m "Bump version to $VERSION"
    $ git neco review
    ```

5. Merge the pull request.
6. Pull `main` branch, add a git tag, then push it.

    ```console
    $ git checkout main
    $ git pull
    $ git tag "v$VERSION"
    $ git push origin "v$VERSION"
    ```

Now the version is bumped up and the latest container image is uploaded to [quay.io](https://quay.io/cybozu/setup-hw).

(Option) Edit GitHub release page
------------
If you would like to give further descriptions of the new release, edit [the GitHub release page](https://github.com/cybozu-go/setup-hw/releases/latest).

[semver]: https://semver.org/spec/v2.0.0.html
[example]: https://github.com/cybozu-go/etcdpasswd/commit/77d95384ac6c97e7f48281eaf23cb94f68867f79
