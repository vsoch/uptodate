# GitHub Action


To use the GitHub action, you can basically checkout your repository,
and then select a parser and root to use.

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

      - name: Check Docker Hierarchy Result
        env:
          result: ${{ needs.test.outputs.dockerhierarchy_matrix }}
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

**under development**

#### Inputs

The following inputs are provided by the action:


| Name | Description | Required | Default |
|------|-------------|----------|---------|
| root | Root path to provide to command. Can be a Dockerfile or directory | false | "" |
| parser | Parser to run, one of dockerfile, or dockerhierarchy | true | |
| dry_run | Do a dry run (don't write, but show changes) one of true or false, defaults to false | false | false |

For the root, if you leave it undefined, the root of your repository will be used, and discovery
of relevant files (e.g., Dockerfile) will be done from there. If `dry_run` is added, no outputs
are produced for next steps as no files are updated or created.

#### Outputs

The following outputs are provided by the action:


| Name | Description |
|------|-------------|
| dockerfile_matrix | A matrix of Dockerfile changes with name and filename set to the Dockerfile name |
| dockerhiearchy_matrix |A matrix of new Dockerfiles and the corresponding tag (Name) |
| dockerfilelist_matrix | The result of Dockerfile list, akin to docker_file matrix but including all files |

See the [examples](https://github.com/vsoch/uptodate/tree/main/.github/examples) folder for a more detailed example.
