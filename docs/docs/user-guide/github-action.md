# GitHub Action


To use the GitHub action, you can basically checkout your repository,
and then select a parser and root to use:

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
      dockerhierarchy_matrix: ${{ steps.dh_check.outputs.dockerhierarchy_matrix }}
      dockerfilelist_matrix: ${{ steps.dl_check.outputs.dockerfilelist_matrix }}
    steps:
    - name: Checkout Actions Repository
      uses: actions/checkout@v2
    - name: Find and Update Dockerfiles in root
      uses: vsoch/uptodate@main
      id: dockerfile_check
      with: 
        parser: dockerfile

    - name: Find and Update Docker Hierarchy in root
      uses: vsoch/uptodate@main
      id: dh_check
      with: 
        parser: dockerhierarchy

    - name: List all Dockerfile
      uses: vsoch/uptodate@main
      id: df_list
      with: 
        parser: dockerfilelist
```

As the action is currently under development, we default to the main branch,
shown above. When the action has releases we will use a release that also
has a pre-built image to make the action run faster.
You might then do something with this output later in the run, such as view it
(or more advanced, pipe into matrix, which is the intended usage):

```yaml
  view:
    needs:
      - test
    runs-on: ubuntu-latest
    steps:
      - name: Check Dockerfile result
        env:
          result: ${{ needs.test.outputs.dockerfile_matrix }}
        run: echo ${result}

      - name: Check Docker Hierarchy Result
        env:
          result: ${{ needs.test.outputs.dockerhierarchy_matrix }}
        run: echo ${result}
```

If you want to use it in a matrix you might do:

```yaml
jobs:
  update:
    runs-on: ubuntu-latest
    outputs:
      dockerfile_matrix: ${{ steps.dockerfile_check.outputs.dockerfile_matrix }}
    steps:
    - name: Checkout Actions Repository
      uses: actions/checkout@v2

    - name: Add new Dockerfiles for build
      uses: vsoch/uptodate@main
      with: 
        parser: dockerhierarchy

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

The following inputs and outputs are provided by the action:

#### Inputs

| Name | Description | Required | Default |
|------|-------------|----------|---------|
| root | Root path to provide to command. Can be a Dockerfile or directory | false | "" |
| parser | Parser to run, one of dockerfile, or dockerhierarchy | true | |
| dry_run | Do a dry run (don't write, but show changes) one of true or false, defaults to false | false | false |

For the root, if you leave it undefined, the root of your repository will be used, and discovery
of relevant files (e.g., Dockerfile) will be done from there. If `dry_run` is added, no outputs
are produced for next steps as no files are updated or created.

#### Outputs

| Name | Description |
|------|-------------|
| dockerfile_matrix | A matrix of Dockerfile changes with name and filename set to the Dockerfile name |
| dockerhiearchy_matrix |A matrix of new Dockerfiles and the corresponding tag (Name) |
| dockerfilelist_matrix | The result of Dockerfile list, akin to docker_file matrix but including all files |

See the [examples](https://github.com/vsoch/uptodate/tree/main/.github/examples) folder for a more detailed example.
