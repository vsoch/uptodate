# Up to Date

![docs/assets/img/uptodate.png](docs/assets/img/uptodate.png)

This is a library and GitHub action to automatically update different assets in your
repository. While the tool can be extended to "other kinds of things" to update, it is
currently very Docker-centric. The tool includes the following commands:
 
  - [dockerfile](#dockerfile): update the `FROM` images in your Dockerfile to the latest hash
  - [dockerhierarchy](#docker-hierarchy): maintain a set of base images and check for new tags. When a new tag is found, create a new Dockerfile to build it.
  
For each of the above, when paired with the GitHub action (not written yet) and
functionality to make updates and then open a pull request, it's possible to
have a repository that will run automated updates of container bases (`dockerfile`)
and builds (`dockerhierarchy`).

## Usage

### Install

To build the library:

```bash
$ make
```

This will create a binary executable, `uptodate` that you can use directly or
copy into a directory on your path.

### dockerfile

This command will read one or more Dockerfiles, and tell us if the digest is
up to date. When run by default, it will automatically update digests.

For example, to update a single Dockerfile, you would do:

```bash
$ ./uptodate dockerfile /path/to/Dockerfile
              _            _       _       
  _   _ _ __ | |_ ___   __| | __ _| |_ ___ 
 | | | | '_ \| __/ _ \ / _  |/ _  | __/ _ \
 | |_| | |_) | || (_) | (_| | (_| | ||  __/
  \__,_| .__/ \__\___/ \__,_|\__,_|\__\___|
       |_|                     dockerfile


  ⭐️ Updated ⭐️
     Checked: 2
    Modified: 0
```

To update an entire directory of Dockerfile:

```bash
$ ./uptodate dockerfile /path/to/directory
```

Or just run without any arguments to parse the present working directory (recursively):

```bash
$ ./uptodate dockerfile
```

If you don't want to write changes but just preview, add `--dry-run`:

```bash
$ ./uptodate dockerhierarchy --dry-run
```

To update your `Dockerfile`s we use [lookout](https://github.com/alecbcs/lookout) for updated versions 

### dockerhierarchy

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

For the above, running the dockerhierarchy updater will see there is an uptodate.yaml in
the folder, parse it, and then understand that the directory is ubuntu specific. It will
then derive all the existing tags for ubuntu (and by default use semver or semantic versioning
to decide whether to include tags) and create new Dockerfile folders for those that
are missing. The reason we need the `uptodate.yaml` is to store preference about
tags to skip, or more generally, a pattern to match. For containers with more complex
names that don't map nicely to a folder, we can just write it there. Here is an example,
with a few examples of filters you might use:

```yaml
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
```

Not including a filter defaults to looking for a numerical (something that has
a minor and major) version and something else. See the [version regex](#version-regular-expressions)
sections for more examples for your recipes.


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

### Development

You can easily use the Makefile to also just build or run:

```bash
$ make

# This won't include formatting to change the files
$ make build
```

or you can use go directly!

```bash
$ go run main.go dockerfile
```

## Previous Art

 - [binoc](https://github.com/autumus/binoc): is another updating tool that uses lookout, and the main difference will be in the design.
