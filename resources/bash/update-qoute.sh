#!/bin/bash

QUOTE=$1
# Update stock price
market-manager purchase update price -s ${QUOTE}
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Update stock 52 week high - low price
market-manager purchase update highLow52week -s ${QUOTE}
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import dividend
market-manager purchase import dividend -s ${QUOTE}
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

echo SUCCESS: Update quote ${QUOTE} finished
