# UpToDate

Welcome to UpToDate 👋  This is a small library and GitHub action to help you with keeping your repository
assets up to date. The following is a quick getting started guide, and you can explore sections in the sidebar for 
more detailed examples. After [installation](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=install), the following parsers are available:

- [Dockerfile](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=dockerfile): update the `FROM` images in your Dockerfile to the latest hash
- [Docker Hierarchy](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=docker-hierarchy): maintain a set of base images and check for new tags. When a new tag is found, create a new Dockerfile to build it.
- [Dockerfile List](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=dockerfile-list): If you simply want a listing of Dockerfiles found in a root in the terminal (and output as a matrix for a GitHub action) you can use this commmand.
- [Docker Build](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=docker-build): Will take an uptodate.yaml (a config file for this tool) and parse it into a GitHub actions matrix (meaning json) that you can use for multistage builds.
- [Git](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=git): Get a build matrix of Git changes for the current commit.

You might also be interested in using the tool as a [GitHub Action](https://vsoch.github.io/uptodate/docs/#/user-guide/github-action)
or directly in a [container](https://github.com/vsoch/uptodate/pkgs/container/uptodate).
