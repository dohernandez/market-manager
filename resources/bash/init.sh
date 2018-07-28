#!/bin/bash

# Run database migration
market-manager migrate up
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import stocks
echo INFO: importing stock
market-manager purchase import stock
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import wallet
echo INFO: importing wallet
market-manager account import wallet
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import transfer
echo INFO: importing transfer
market-manager banking import transfer
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

# Import operation
echo INFO: importing operation
market-manager account import operation
if [ $? -ne 0 ]; then
    echo ERROR: Init failed
    exit 1
fi

echo SUCCESS: Init finished
