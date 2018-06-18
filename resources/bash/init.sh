#!/bin/bash

# Run database migration
market-manager migrate up
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import stocks
market-manager purchase import quote
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import wallet
market-manager account import wallet
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import transfer
market-manager banking import transfer
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import operation
market-manager account import operation
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Update stock price
market-manager purchase update price
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Update stock 52 week high - low price
market-manager purchase update highLow52week
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import dividend
market-manager purchase import dividend
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

echo SUCCESS: Init finished
