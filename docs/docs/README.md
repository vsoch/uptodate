# UpToDate

Welcome to UpToDate 👋  This is a small library and GitHub action to help you with keeping your repository
assets up to date. The following is a quick getting started guide, and you can explore sections in the sidebar for 
more detailed examples. After [installation](#install), the following parsers are available:

- [Dockerfile](#dockerfile): update the `FROM` images in your Dockerfile to the latest hash
- [Docker Hierarchy](#docker-hierarchy): maintain a set of base images and check for new tags. When a new tag is found, create a new Dockerfile to build it.

And for additional help, see examples of 

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
a minor and major) version and something else. See the [version regex]()
sections for more examples for your recipes.
