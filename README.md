# Up to Date

![docs/assets/img/uptodate.png](docs/assets/img/uptodate.png)

This is a library and GitHub action to automatically update different assets in your
repository. While the tool can be extended to "other kinds of things" to update, it is
currently very Docker-centric. The tool includes the following commands:
 
  - [dockerfile](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=dockerfile): update the `FROM` images in your Dockerfile to the latest hash
  - [dockerhierarchy](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=docker-hierarchy): maintain a set of base images and check for new tags. When a new tag is found, create a new Dockerfile to build it.
  - [dockerfilelist](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=dockerfile-list): List all Dockerfiles discovered at some root.
  - [dockerbuild](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=docker-build): Build a matrix of Dockerfiles from an uptodate config file (with build args)
  - [git](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=git): Look for updated files as a matrix.
  
For each of the above, when paired with the [GitHub action](https://vsoch.github.io/uptodate/docs/#/user-guide/github-action) and
functionality to make updates and then open a pull request, it's possible to
have a repository that will run automated updates of container bases (`dockerfile`)
and builds (`dockerhierarchy`), and then (optionally) builds additional images with (`dockerbuild`) or
lists all available Docker files for other interactions (`dockerlist`). 
You can read more at:

⭐️ [The UpToDate Documentation](https://vsoch.github.io/uptodate) ⭐️


## Previous Art

 - [binoc](https://github.com/autumus/binoc): is another updating tool that uses lookout, and the main difference will be in the design.
