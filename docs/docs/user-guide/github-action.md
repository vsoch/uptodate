# GitHub Action


To use the GitHub action, you can basically checkout your repository,
and then select a parser and root to use.

> **Important!** If you use any of the parsers with `--changes` (the GitHub action variables `changes:true` for a parser or just using the git parser, you _must_ perform a checkout that will checkout the ref in full. Without these steps, the respository will be in a detached head state, and the functionality will not work.

Here are examples of how to do a checkout, either for a pull request or direct push.

```yaml
# default will checkout detached, which won't work for using git in the github action container!
- uses: actions/checkout@v2
  if: github.event_name == 'pull_request'
  with:
    fetch-depth: 0
    ref: ${{ github.event.pull_request.head.ref }}

- uses: actions/checkout@v2
  if: github.event_name == 'push'
  with:
     fetch-depth: 0
```

If you forget to do this, you'll see this error:

```bash
              _            _       _       
  _   _ _ __ | |_ ___   __| | __ _| |_ ___ 
 | | | | '_ \| __/ _ \ / _  |/ _  | __/ _ \
 | |_| | |_) | || (_) | (_| | (_| | ||  __/
  \__,_| .__/ \__\___/ \__,_|\__,_|\__\___|
       |_|                          git

2021/09/10 01:05:08 Cannot get previous commit: object not found
```

If you don't use `--changes` / `changes: true` or the git parser, you need not worry.

## Dockerfile

The following is an example of using the uptodate action to update Dockerfiles,
and then outputting the result into a string:

```yaml
name: update-containers

on:  
  schedule:
    - cron:  '0 4 * * *'

jobs:
  test:
    name: Run Dockerfile Updater
    runs-on: ubuntu-latest
    outputs:
      dockerfile_matrix: ${{ steps.dockerfile_check.outputs.dockerfile_matrix }}
    steps:
    - name: Checkout Actions Repository
      uses: actions/checkout@v2
    - name: Find and Update Dockerfiles in root
      uses: vsoch/uptodate@main
      id: dockerfile_check
      with: 
        parser: dockerfile

  view:
    needs:
      - test
    runs-on: ubuntu-latest
    steps:
      - name: Check Dockerfile result
        env:
          result: ${{ needs.test.outputs.dockerfile_matrix }}
        run: echo ${result}
```

or you might choose to pipe the json into a matrix.

```yaml
jobs:
  update:
    runs-on: ubuntu-latest
    outputs:
      dockerfile_matrix: ${{ steps.dockerfile_check.outputs.dockerfile_matrix }}
    steps:
    - name: Checkout Actions Repository
      uses: actions/checkout@v2

    - name: Update existing Dockerfile hashes
      uses: vsoch/uptodate@main
      id: dockerfile_check
      with: 
        parser: dockerfile
         
  build:
    needs:
      - update
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        result: ${{ fromJson(needs.update.outputs.dockerfile_matrix) }}

    name: "Build ${{ matrix.result.name }}"
    steps:
      - name: Checkout Actions Repository
        uses: actions/checkout@v2
      - name: Build container
        run:
            # You might also want to change into the Dockerfile directory first!
            docker build -f ${{ matrix.result.name }} -t mycontainer/name:tag .
```

As the action is currently under development, we default to the main branch,
shown above. When the action has releases we will use a release that also
has a pre-built image to make the action run faster.


##  Dockerfile Hierarchy

A Dockerfile hierarchy requires an `uptodate.yaml` file as described [here](https://vsoch.github.io/uptodate/docs/#/user-guide/user-guide?id=uptodate-yaml). The relevant section is the `dockerhierarchy` key,
and it will specify a container base and regular expressions of tags to look for.
Here is a basic example that creates new `Dockerfile`s based on finding new tags
for an image, and then outputting the newly found files as a json string
for a GitHub action:

```yaml
name: update-containers

on:  
  schedule:
    - cron:  '0 4 * * *'

jobs:
  test:
    name: Run Dockerfile Updater
    runs-on: ubuntu-latest
    outputs:
      dockerhierarchy_matrix: ${{ steps.dh_check.outputs.dockerhierarchy_matrix }}
    steps:
    - name: Checkout Actions Repository
      uses: actions/checkout@v2

    - name: Find and Update Docker Hierarchy in root
      uses: vsoch/uptodate@main
      id: dh_check
      with: 
        parser: dockerhierarchy

  view:
    needs:
      - test
    runs-on: ubuntu-latest
    steps:
      - name: Check Dockerfile result
        env:
          result: ${{ needs.test.outputs.dockerhierachy_matrix }}
        run: echo ${result}
```

And equivalently, you might parse this into a matrix:

```yaml
jobs:
  update:
    runs-on: ubuntu-latest
    outputs:
      dockerhierarchy_matrix: ${{ steps.dh_check.outputs.dockerhierarchy_matrix }}
    steps:
    - name: Checkout Actions Repository
      uses: actions/checkout@v2

    - name: Add new Dockerfiles for build
      uses: vsoch/uptodate@main
      id: dh_check
      with: 
        parser: dockerhierarchy

  build:
    needs:
      - update
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        result: ${{ fromJson(needs.update.outputs.dockerhierarchy_matrix) }}

    name: "Build ${{ matrix.result.name }}"
    steps:
      - name: Checkout Actions Repository
        uses: actions/checkout@v2
      - name: Build container
        run:
            # You might also want to change into the Dockerfile directory first!
            docker build -f ${{ matrix.result.name }} -t mycontainer/name:tag .

```

The above is pretty cool when you combine with the Dockerfile updater, because you
can imagine generating new `Dockerfile` first, and then updating the files to have
the latest hashes, and outputting the final list of all of them.
  
## Dockerfile List

Finally, Dockerfile list is useful if you don't want any special parsing or updates,
but just a list of Dockerfiles. I use this when I want to do automated builds of all
my Dockerfile, no matter what!

```yaml
name: list-containers

on:  
  schedule:
    - cron:  '0 4 * * *'

jobs:
  test:
    name: Run Dockerfile List
    runs-on: ubuntu-latest
    outputs:
      dockerfilelist_matrix: ${{ steps.dl_list.outputs.dockerfilelist_matrix }}
    steps:
    - name: Checkout Actions Repository
      uses: actions/checkout@v2

    - name: List all Dockerfile
      uses: vsoch/uptodate@main
      id: df_list
      with: 
        parser: dockerfilelist

  view:
    needs:
      - test
    runs-on: ubuntu-latest
    steps:
      - name: Check Dockerfile result
        env:
          result: ${{ needs.test.outputs.dockerfilelist_matrix }}
        run: echo ${result}
```

Or used in a matrix:

```yaml
jobs:
  update:
    runs-on: ubuntu-latest
    outputs:
      dockerfilelist_matrix: ${{ steps.dl_list.outputs.dockerfilelist_matrix }}
    steps:
    - name: Checkout Actions Repository
      uses: actions/checkout@v2

    - name: List all Dockerfile
      uses: vsoch/uptodate@main
      id: df_list
      with: 
        parser: dockerfilelist
         
  build:
    needs:
      - update
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        result: ${{ fromJson(needs.update.outputs.dockerfilelist_matrix) }}

    name: "Build ${{ matrix.result.name }}"
    steps:
      - name: Checkout Actions Repository
        uses: actions/checkout@v2
      - name: Build container
        run:
            # You might also want to change into the Dockerfile directory first!
            docker build -f ${{ matrix.result.name }} -t mycontainer/name:tag .
```


## Docker Build

A Docker Build is similar to a Docker Hierarchy, except instead of putting instructions
in the `uptodate.yaml` for how to build base images, we have a `dockerbuild` section
that describes a matrix.


```yaml
name: docker-build-matrix

on:  
  schedule:
    - cron:  '0 2 * * *'

jobs:
  test:
    name: Run Docker Builds
    runs-on: ubuntu-latest
    outputs:
      dockerbuild_matrix: ${{ steps.dbuild.outputs.dockerbuild_matrix }}
    steps:
    - name: Checkout Repository
      uses: actions/checkout@v2

    - name: Generate Docker Build Matrices
      uses: vsoch/uptodate@main
      id: dbuild
      with: 
        parser: dockerbuild

  view:
    needs:
      - test
    runs-on: ubuntu-latest
    steps:
      - name: Check Docker Build Result
        env:
          result: ${{ needs.test.outputs.dockerbuild_matrix }}
        run: echo ${result}
```

**Matrix example coming shortly!**


## Git

Here is how to get a matrix of changed git files:

```yaml
name: git-changes-matrix

on:  
  schedule:
    - cron:  '0 2 * * *'

jobs:
  test:
    name: Get git changes
    runs-on: ubuntu-latest
    outputs:
      git_matrix: ${{ steps.git.outputs.git_matrix }}
    steps:
    - name: Checkout Repository
      uses: actions/checkout@v2

    - name: Get git changes
      uses: vsoch/uptodate@main
      id: git
      with: 
        parser: git

  view:
    needs:
      - test
    runs-on: ubuntu-latest
    steps:
      - name: Check Docker Build Result
        env:
          result: ${{ needs.test.outputs.git_matrix }}
        run: echo ${result}
```

Note that if you need to change the branch compared against, set `flags:` to include `--branch master`.
If it's found that you are on the same branch you are comparing against, instead
we look one commit back, the assumption being that you merged and the changes are found in the
last commit.

#### Inputs

The following inputs are provided by the action:


| Name | Description | Required | Default |
|------|-------------|----------|---------|
| root | Root path to provide to command. Can be a Dockerfile or directory | false | "" |
| parser | Parser to run, one of dockerfile, or dockerhierarchy | true | |
| dry_run | Do a dry run (don't write, but show changes) one of true or false, defaults to false | false | false |
| changes| Only include changed files (defaults to false) | false | false |

For the root, if you leave it undefined, the root of your repository will be used, and discovery
of relevant files (e.g., Dockerfile) will be done from there. If `dry_run` is added, no outputs
are produced for next steps as no files are updated or created.

#### Outputs

The following outputs are provided by the action:


| Name | Description |
|------|-------------|
| dockerfile_matrix | A matrix of Dockerfile changes with name and filename set to the Dockerfile name |
| dockerfile_matrix_empty | A boolean true/false if the matrix is empty or not |
| dockerhierarchy_matrix |A matrix of new Dockerfiles and the corresponding tag (Name) |
| dockerhierarchy_matrix_empty | A boolean true/false if the matrix is empty or not |
| dockerfilelist_matrix | The result of Dockerfile list, akin to docker_file matrix but including all files |
| dockerfilelist_matrix_empty | A boolean true/false if the matrix is empty or not |
| dockerbuild_matrix | The result of the Docker Build parser, a build matrix to pipe into next steps |
| dockerbuild_matrix_empty | A boolean true/false if the matrix is empty or not |
| git_matrix | A matrix of changed files, each with a `Name` (change type) and `Filename` |
| git_matrix_empty | A boolean true/false if the matrix is empty or not |

See the [examples](https://github.com/vsoch/uptodate/tree/main/.github/examples) folder for a more detailed example.
