#!/bin/bash

set -e

printf "Found files in workspace:\n"
ls

printf "Looking for uptodate install...\n"
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
