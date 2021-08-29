# Up to Date

This is a library and GitHub action to automatically update different assets in your
repository. To start, we will be updating the `FROM` images of Dockerfiles.

**under development! not ready for use!**

## Usage

### Install

To build the library:

```bash
$ make
```

This will create a binary executable, `uptodate` that you can use directly or
copy into a directory on your path.

### dockerfile

The uptodate dockerfile command will update one or more Dockerfile.

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

To update your `Dockerfile`s we will use:

 - [containerspec](https://github.com/vsoch/containerspec): for LABELS
 - [lookout](https://github.com/alecbcs/lookout) for updated versions
 

Specifically, I'd like to have commands that can read a Dockerfile, or
a directory / repository of `Dockerfile`s, and be able to tell us:

1. Is the digest up to date?
2. Are there new tags we might want to build?

For base images and updating them this means:

1. User can target a Dockerfile directly for one off update, or a folder with tags for scaled
2. Read in Dockerfile, keep track of labels and FROMS (add dockerfile parser)
3. Look at labels to see if a tag is there for the hash
4. For each FROM, look up list of tags, update hash (use lookout)
5. For each FROM, if a label exists after it for opencontainers, delete it
6. Update label to use new tag

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
