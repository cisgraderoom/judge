#!/bin/bash

START=$(date +%s);

FILENAME="$1";

EXT="$2";

OUTPUT="$3";

ERR="$4";

if [[ -z "$FILENAME" && -z "$EXT" ]]; then
	echo "Invalid argument"
	echo "example: ./compile filename extension output_filename"
fi

function execution_time
{
	END=$(date +%s);
	echo "compile_time: $((END-START))"
	exit 0
}

if [[ $EXT == "java" ]]; then
	javac $FILENAME >/dev/null 2>$ERR &
	
	pid="$!"
	wait "$pid"

	EXITCODE=$?

	echo "exit_code: $EXITCODE"
	execution_time
fi

if [[ $EXT == "cpp" ]]; then

	g++ $FILENAME -O2 -o $OUTPUT >/dev/null 2>$ERR &

	pid="$!"
	wait "$pid"

	EXITCODE=$?
	
	echo "exit_code: $EXITCODE"
	execution_time

fi

if [[ $EXT == "c" ]]; then

	gcc $FILENAME -O2 -o $OUTPUT >/dev/null 2>$ERR &
	
	pid="$!"
	wait "$pid"

	EXITCODE=$?

	echo "exit_code: $EXITCODE"
	execution_time

fi
