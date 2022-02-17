# User's Guide
  
## How it works

This is a Go library that will look for `Dockerfile`s and (eventually) other assets
in your repository and help you keep this up to date (hence the name!) This means for:

 - **Dockerfile**: is an updated that will look for all Dockerfiles, find `FROM` lines, and make sure we are using the most up to date hash. We do not advance the tag itself (e.g., 18.04 will not be updated to 20.04) but just the sha256 sum, in the case that there are security updates, etc.
 - **Docker Hierarchy**: is a structure that has some top level folder identified by an `uptodate.yaml` file, which is described [here](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=uptodate-yaml). Within this folder are subfolders that correspond to tags, and the tool looks for new tags, and generates you a template Dockerfile to build if there are. 
 - **Dockerfile List**: updater will simply find all your Dockerfiles, and list them or provide them in a json output for a GitHub action for further parsing.
 - **Docker Build**: provides a matrix of builds to pipe into a GitHub Action.
 - **Git**: is an updater that will list changed files (and generate the same information in a matrix).
 
For all of the above, you can run the tool manually on the command line, or as a GitHub action.
For the last (Docker Build) the builds aren't currently performed for you (and you can [request this](https://github.com/vsoch/uptodate/issues) if you'd like it)
as it is expected to be run in a GitHub Workflow. With the GitHub action you can make workflows to check for updates at some frequency, and open a pull request with updates to test if/when a new version is found and a file is created or updated.


## Install

To install the library, first clone the repository:

```bash
$ git clone https://github.com/vsoch/uptodate
$ cd uptodate
```

You will need a recent version of Go on your path. And then to build the library:

```bash
$ make
```

This will create a binary executable, `uptodate` that you can use directly or
copy into a directory on your path. If you don't want to install, you can use
the set of [containers](https://github.com/vsoch/uptodate/pkgs/container/uptodate) available on GitHub packages, or the [GitHub Action](https://vsoch.github.io/uptodate/docs/#/user-guide/github-action).


## Commands

The following commands are available.

### Dockerfile

?> $ uptodate dockerfile

This command will read one or more Dockerfiles, and tell us if the digest is
up to date. For this parser, an update means the digest of a `FROM` statement, or a build argument
that follows a particular convention. 

#### Digest Update

For digests, you might see that:

```dockerfile
FROM ubuntu:20.04
```

is updated to

```dockerfile
FROM ubuntu:18.04@sha256:9bc830af2bef73276515a29aa896eedfa7bdf4bdbc5c1063b4c457a4bbb8cd79
```

And then subsequent updates will compare the digest to any known, newer one.

#### Build Arguments

For build arguments, it can only work given that you name them according to
a known uptodate updater, and you provide a default. Here is a generate format:

```dockerfile
ARG uptodate_<build-arg-type>_<build-arg-value>=<default>
```

##### Spack Build Argument

For example, here is how we might specify a build arg for a version of the spack package, "ace"

```dockerfile
ARG uptodate_spack_ace=6.5.6
```
After the updater runs, if it finds a new version 6.5.12, the line will read:

```dockerfile
ARG uptodate_spack_ace=6.5.12
```

##### GitHub Release Build Argument

Let's say we are installing spack itself, and we want our updater to find new releases
on GitHub. The line might look like this:

```dockerfile
ARG uptodate_github_release_spack__spack=v0.16.1
```

And the double underscore indicates indicates the separation between the organization
and repository names. So the above will look for new releases on [https://github.com/spack/spack](https://github.com/spack/spack)
and update as follows:

```dockerfile
ARG uptodate_github_release_spack__spack=v0.16.2
```

Since we don't see any use cases for a container identifier as an ARG, we don't currently support this.
But if you do, please [open an issue](https://github.com/vsoch/uptodate/issues).

#### GitHub Commit Build Argument

If you want a more bleeding edge update (e.g., re-build the image every time there
is a new commit) then you'll likely want a GitHub commit argument, which takes the following format:

```dockerfile
ARG uptodate_github_commit_<org>__<name>__<branch>=<release-tag>
```

For example, to update spack based on commits from develop, we might do:

```dockerfile
ARG uptodate_github_commit_spack__spack__develop=NA
```

We are using NA for the first check because we are lazy and don't want to look up
the actual commit - uptodate will find it for us!

#### Example

As an example example, to update a single Dockerfile, you would do:

```bash
$ ./uptodate dockerfile /path/to/Dockerfile
              _            _       _       
  _   _ _ __ | |_ ___   __| | __ _| |_ ___ 
 | | | | '_ \| __/ _ \ / _  |/ _  | __/ _ \
 | |_| | |_) | || (_) | (_| | (_| | ||  __/
  \__,_| .__/ \__\___/ \__,_|\__,_|\__\___|
       |_|                     dockerfile


  ⭐️ Updated ⭐️
     Checked: 1
    Modified: 0
```

To update an entire directory of `Dockerfile`s:

```bash
$ ./uptodate dockerfile /path/to/directory
```

Or just run without any arguments to parse the present working directory (recursively):

```bash
$ ./uptodate dockerfile
```

If you don't want to write changes but just preview, add `--dry-run`:

```bash
$ ./uptodate dockerfile --dry-run
```

or if you just want to update changed files (for the current git commit)

```bash
$ uptodate dockerfile --changes
```

To update your `Dockerfile`s we use [lookout](https://github.com/alecbcs/lookout) for updated versions 


### Docker Hierarchy

?> $ uptodate dockerhierarchy


Have you ever noticed that when people maintain a repository of Dockerfile, they
tend to give it structure like:

```
container/
   tag1/
      Dockerfile
   tag2/
      Dockerfile
```

I have! And I think it makes sense to create a command alongside this tool that
can easily read that structure and then determine if there are new tags that we might want
to build. By default, we want the tool to automatically discover these organized hierarchies,
so we suggest adding an `uptodate.yaml` file to indicate a directory:

```
ubuntu/
   uptodate.yaml
   latest/
     Dockerfile
   20.04/
     Dockerfile
   18.04/
     Dockerfile
```

and then we would run the command to generate new tag subfolders and files!

```bash
$ ./uptodate dockerhierarchy
              _            _       _       
  _   _ _ __ | |_ ___   __| | __ _| |_ ___ 
 | | | | '_ \| __/ _ \ / _  |/ _  | __/ _ \
 | |_| | |_) | || (_) | (_| | (_| | ||  __/
  \__,_| .__/ \__\___/ \__,_|\__,_|\__\___|
       |_|               dockerhierarchy

Copying /home/vanessa/go/src/github.com/vsoch/uptodate/ubuntu/20.04/Dockerfile to /home/vanessa/go/src/github.com/vsoch/uptodate/ubuntu/21.04/Dockerfile

  ⭐️ Updated ⭐️
     Updated versions for ubuntu: [21.04]
     Present versions for ubuntu: [16.04 18.04 20.04]
```

The tool will look for the `uptodate.yaml` and proceed to parse the directory when it finds one.
If you don't want to write new files, but rather preview what will be added, add `--dry-run`:

```bash
$ ./uptodate dockerhierarchy --dry-run
```

#### Uptodate Yaml

The `uptodate.yaml` file's presence in a root (such as the relative path `./ubuntu` above)
indicates that the directory is structured like  a container URI, with tags as the subfolders. 
However, you are free to put other subfolders within the directory that don't correspond to tags, such
as matrix builds (described in a [following section](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=docker-build)) that might use the image. For the above, running the dockerhierarchy updater will see there is an uptodate.yaml in
the folder, parse it, and then read the `dockerhierarchy` section and understand that the directory is ubuntu specific. It will
then derive all the existing tags for ubuntu (and by default use semver or semantic versioning
to decide whether to include tags) and create new Dockerfile folders for those that
are missing. The reason we need the `uptodate.yaml` is to store preference about
tags to skip, or more generally, a pattern to match. For containers with more complex
names that don't map nicely to a folder, we can just write it there. Here is an example for the `dockerhierarchy` updater,
with a few examples of filters you might use:

```yaml
dockerhierarchy:
  container:
    name: ubuntu
    filter: 
      # Don't include anything that starts with arm
      - "^((?!arm).)*$"
      # include anything that starts with 3.9 (e.g., 3.9.1)
      - "3.9*"

    # The earliest version that should be used
    startat: 16.04

    # Skip these versions (e.g., not long term releases or LTS)
    skips:
      - "17.04"
      - "19.04"

    # Include these versions no matter what
    includes:
     - "21.10"
```

Not including a filter defaults to looking for a numerical (something that has
a minor and major) version and something else. See the [version regex](/user-guide/user-guide?id=version-regular-expressions)
sections for more examples for your recipes. 

### Dockerfile List

?> $ uptodate dockerfilelist

If you want to just list Dockerfiles discovered, use:

```bash
$ ./uptodate dockerfilelist
/home/vanessa/go/src/github.com/vsoch/uptodate/Dockerfile
/home/vanessa/go/src/github.com/vsoch/uptodate/shpc/Dockerfile
/home/vanessa/go/src/github.com/vsoch/uptodate/shpc/Dockerfile.tcl
/home/vanessa/go/src/github.com/vsoch/uptodate/tests/ubuntu/21.04/Dockerfile
```

if you just want to list updated (modified or added) you can do:

```bash
$ uptodate dockerfilelist --changes
```

If you want to exclude Dockerfiles with any build args:

```bash
$ uptodate dockerfilelist --no-build-args
```

or just empty build args:

```bash
$ uptodate dockerfilelist --no-empty-build-args
```


You can also explicitly ask for more than one path:


```bash
$ uptodate dockerfilelist --no-empty-build-args ubuntu/20.04/Dockerfile ubuntu/18.04/Dockerfile
```


### Docker Build

Docker build is similar to the Docker Hierarchy updater in that it reads an `uptodate.yaml`
and then generates one or more build matrices for it. You have two options for this dockerbuild section:

1. Include a predefined build matrix, one with pairs of values to build
2. Only include build args and constraints, in which case all combinations are built.

For the first example, the matrices will include all versions or other variables
that you've specified, along with the Dockerfile that are discovered under the root where
the `uptodate.yaml` is. The matrices can be parsed into
a GitHub action to drive further container builds using one or more base images.
For example, let's say that we start with this configuration file:

```yaml
dockerbuild:
  
  # switch to off to stop building, defaults to true if not present
  active: true
  
  build_args:
  
    # This is an example of a manual build arg, versions are required
    llvm_version:

      # The key is a shorthand used for naming (required)
      key: llvm
      versions:
       - "4.0.0"
       - "5.0.1"
       - "6.0.0"

    # This is an example of a spack build arg, the name is the package
    abyss_version:
      key: abyss
      name: abyss
      type: spack

    # This will be parsed by the Dockerfile parser, name is the container name
    ubuntu_version:

      key: ubuntu
      name: ubuntu
      type: container
      startat: "16.04"
      endat: "20.04"
      filter: 
        - "^[0-9]+[.]04$" 
      skips:
      - "17.04"
      - "19.04"
```

You'll see the primary section of interest is under `dockerbuild`, and under this
we have three build args. There are three `type` of build args:

 - *manual*: meaning you define a name and a list of versions or values, no extra parsing or updating done!
 - *spack*: derive a list of versions from spack, with the same options to start at, filter, skip, etc. The data is parsed from [the spack packages interface](https://spack.github.io/packages/) that is updated nightly from spack develop.
 - *container*: meaning you define similar fields to if you were asking to update Dockerfile froms - a container name, startat (version), filter, and versions to skip. If you include a tag with your container, we will simply update digests (and keep the same tag) so you'll get a much smaller matrix.

For each, the key is used to derive a suggested container name.  If you provide a container build arg variable,
it's assumed to be for the `FROM` and will be added to the container name. Other variables will be represented in the tag.

For the second example, here we see a configuration file that has a predefined "matrix":

```yaml
dockerbuild:

  # This says "only build these combinations of build args"
  # variables left out won't be included (e.g., abyss)
  matrix:
    llvm_version: [4.0.0, 5.0.1, "6.0.0"]
    ubuntu_version: ["16.04", "18.04", "20.04"]
    
  # With a matrix, we will derive possible versions from build args here
  # without the matrix above, we build all possible versions
  build_args:

    # This is an example of a manual build arg, versions are required
    llvm_version:

      # The key is a shorthand used for naming (required)
      key: llvm

    # This is an example of a spack build arg, the name is the package
    abyss_version:
      key: abyss
      type: spack

    # This will be parsed by the Dockerfile parser, name is the container name
    ubuntu_version:

      key: ubuntu
      name: ubuntu
      type: container
```

Since we are going to be honoring the list verbatim, you don't need to worry about
extra metadata for the build args beyond the key, name, and type. Finally, for either
type of matrix generation, you can exclude particular combinations. The following:

```yaml
# Specific combinations of build args to exclude (optional)
exclude:
  abyss_version: ["2.3.1", "2.3.1"]
  ubuntu_version: ["21.04", "21.04"]
  llvm_version: ["6.0.0", "5.0.1"]
```

Would exclude the two builds (you read in pairs top to bottom, and the list lengths
must be the same:

 - abyss 2.3.1, ubuntu 21.04, llvm 6.0.0
 - abyss 2.3.1, ubuntu 21.04, llvm 5.0.1

To run this for a repository root:

```bash
$ uptodate dockerbuild 
```

And to only include changed uptodate.yaml files:

```bash
$ uptodate dockerbuild --changes
```
And finally, if you want the most functionality (not just to look for changes to files but
also check previous images and rebuild those that have detailed changes in build args) then you
need to provide uptodate with the registry that the images are intended for (and this also
assumes you use the build names that are generated).

```bash
$ uptodate dockerbuild --registry ghcr.io/rse-radiuss
```

Note that the registry flag is enough to indicate that you want to look for changes of
this type. Finally, since we parse Dockerfiles to see if changes are needed, if you
want to over-ride this check and just build the entire matrix:

```bash
$ uptodate dockerbuild --all
```

But sometimes the nesting of folders isn't what you want! If you want to define your
own container basename:

```yaml
dockerbuild:
  container_basename: ubuntu-hip-cuda
```

### Git

The git uptodate parser will simply generate a json structure of changed files.
For example, here we are running it locally to see that the changes in the most
recent commit include an update to a GitHub workflow:

```bash
$ ./uptodate git /path/to/repo
              _            _       _       
  _   _ _ __ | |_ ___   __| | __ _| |_ ___ 
 | | | | '_ \| __/ _ \ / _  |/ _  | __/ _ \
 | |_| | |_) | || (_) | (_| | (_| | ||  __/
  \__,_| .__/ \__\___/ \__,_|\__,_|\__\___|
       |_|                          git


  ⭐️ Changed Files ⭐️
    .github/workflows/build-matrices.yaml: Modify
```

By default, the name of the branch we compare against is "main." To change this:

```bash
$ ./uptodate git /path/to/repo --branch master
```

You can also use this functionality with the `dockerbuild` or `dockerfile` or `dockerfilelist` parsers by adding `--changes`
to only include changed files. The `--branch` flag works here as well, and can be added with `flags: --branch master` in the GitHub action. If it's found that you are on the same branch you are comparing against, instead
we look one commit back, the assumption being that you merged and the changes are found in the
last commit.

### GitHub Action

For all of the commands above, if you run them in a GitHub action, a matrix of results
will be produced that you can pipe into a build matrix, or parse as a string for your
own usage. See the [GitHub Action](https://vsoch.github.io/uptodate/docs/#/user-guide/github-action)
for more details.

### Version Regular Expressions

When selecting versions, you will need to customize your choice based on the container
or package of interest. Here is a long list of experiments for you to experiment with.
For all of the below, remove the `^` or `$` if you want to allow an additional prefix
or suffix, respectively. It's helpful to run the updates with `--dry-run` if you
just want to preview changes first.

```yaml
filter:

  # This is the default if you specify nothing
  - "[0-9]+[.][0-9]+(?:[.][0-9]+)?"

  # we only want XX.04 versions (e.g., ubuntu)
  - "^[0-9]+[.]04$"

  # we only want XX.XX versions
  - "^[0-9]+[.][0-9]+$"    

  # Don't include anything that starts with arm
  - "^((?!arm).)*$"

  # include anything that starts with 3.9 (e.g., 3.9.1)
  - "3.9*"
```
