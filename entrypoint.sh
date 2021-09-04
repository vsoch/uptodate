#!/bin/bash

set -e

which uptodate

COMMAND="uptodate ${INPUT_PARSER} "

# dry run?
if [ "${INPUT_DRY_RUN}" == "true" ]; then
    COMMAND="${COMMAND} --dry-run"
fi

# Add parser specific flags and the root
COMMAND="${COMMAND} ${INPUT_FLAGS} ${INPUT_ROOT}"
echo "${COMMAND}"

${COMMAND}
echo $?
