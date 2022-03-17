#!/bin/bash

START=$(date +%s);

FILELIMIT="$1";

TIMEOUT="$2";

MEM="$3";

FILENAME="$4";

INPUT="$5";

OUTPUT="$6";

COMPARE="$7";

RESULT="$8";


if [[ -z "$FILENAME" && -z "$EXT" ]]; then
	echo "Invalid argument"
	echo "example: ./compile filename extension output_filename"
	exit;
fi


$FILELIMIT -t $TIMEOUT -m $MEM $FILENAME  < $INPUT > $OUTPUT &
sleep $TIMEOUT
if [ $? != 0 ]
then
	echo "$(cat $RESULT)E" >  $RESULT
	exit
fi
if cmp $OUTPUT $COMPARE; then
	echo "$(cat $RESULT)P" >  $RESULT
else
	echo "$(cat $RESULT)F" >  $RESULT
fi
