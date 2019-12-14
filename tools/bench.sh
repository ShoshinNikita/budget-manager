#!/bin/bash
# This script makes requests to http://localhost:8080/years/${current_year}/months/${current_month}
# via 'vegeta' for 5 seconds with 25 workers

YEAR=$(date +"%Y")
MONTH=$(date +"%m")
URL="http://localhost:8080/years/${YEAR}/months/${MONTH}"
DURATION="5s"
#
RESULT_FILE="results.bin"

# Check whether 'vegeta' is installed
vegeta --version &> /dev/null
if [ "$?" != "0" ]; then
	echo "'vegeta' isn't installed. Installation options: https://github.com/tsenart/vegeta"
	exit 1
fi

# Check connection
curl -s "${URL}" > /dev/null
if [ "$?" != "0" ]; then
	echo "App is down"
	exit 1
fi

# Start benchmark
echo "Start benchmark for '${URL}'..."

echo "GET ${URL}" | vegeta attack -duration=${DURATION} -rate=0 -max-workers=25 > ${RESULT_FILE}

cat ${RESULT_FILE} | vegeta report
cat ${RESULT_FILE} | vegeta report -type="hist[25ms,50ms,100ms,200ms,400ms]"

# Cleanup
rm ${RESULT_FILE}
