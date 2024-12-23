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

1. Determine a new version number. Then set `VERSION` variable.

    ```console
    # Set VERSION and confirm it. It should not have "v" prefix.
    $ VERSION=x.y.z
    $ echo $VERSION
    ```

2. Make a branch to release

    ```console
    $ git checkout main
    $ git pull
    $ git checkout -b "bump-$VERSION"
    ```

3. Edit `CHANGELOG.md` for the new version ([example][]).
4. Commit the change and create a new pull request

    ```console
    $ git commit -a -m "Bump version to $VERSION"
    $ git push -u origin HEAD
    $ gh pr create -f
    ```

5. Merge the pull request.
6. Pull `main` branch, add a git tag, then push it.

    ```console
    # Set VERSION again.
    $ VERSION=x.y.z
    $ echo $VERSION

    $ git checkout main
    $ git pull
    $ git tag -a -m "Release v$VERSION" "v$VERSION"

    # Make sure the release tag exists.
    $ git tag -ln | grep $VERSION

    $ git push origin "v$VERSION"
    ```

GitHub actions will build and push artifacts such as container images and
create a new GitHub release.

[semver]: https://semver.org/spec/v2.0.0.html
[example]: https://github.com/cybozu-go/etcdpasswd/commit/77d95384ac6c97e7f48281eaf23cb94f68867f79
