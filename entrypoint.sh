#!/bin/bash

set -e

which uptodate

COMMAND="uptodate ${INPUT_PARSER} "

# dry run?
if [ "${INPUT_DRY_RUN}" == "true" ]; then
    COMMAND="${COMMAND} --dry-run"
fi

COMMAND="${COMMAND} ${INPUT_ROOT}"
echo "${COMMAND}"

${COMMAND}
echo $?
