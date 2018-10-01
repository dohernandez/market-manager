# Market Manager

An application to manage the bull market allowing us to keep tracking the prices and dividends of the stocks we prefer, from any market in the world.
               
Also, allow us to manage our own wallet, storing all the operations we have made used.

## Table of Contents

* [Resources](#resources)
* [Getting started](#getting-started)
    * [Import tools](#import-tools)
        * [Stocks](#import-stocks)
        * [Wallets](#import-wallet)
        * [Transfers](#import-transfer)
* [Dependencies](#dependencies)
    * [Upstream](#upstream)


## Resources
<br />[[table of contents]](#table-of-contents)

## Getting started

### Import tools

#### Import stocks

Add stocks
    
* Add/Create `csv` file to `resources/import/stocks` with the stock(s) with the following format:
    
    | STOCK NAME         | EXCHANGE SYMBOL | SYMBOL | TYPE   | Sector     | Industry                    |
    |--------------------|-----------------|--------|--------|------------|-----------------------------|
    | NVIDIA CORPORATION | NASDAQ          | NVDA   | COMMON | TECHNOLOGY | SEMICONDUCTOR - SPECIALIZED |  

**Note:** The file **SHOULD** only contain the values, the header is just for better understanding.

* Run the command
    
    ```bash
    $ market-manager purchase import stock
    ```

<br />[[table of contents]](#table-of-contents)

#### Import wallet

Add/Create wallets

* Add/Create `csv` file to `resources/import/wallets` with the wallet(s) with the following format:
    
    | URL            |  NAME  |
    |----------------|--------|
    | www.wallet.es  | Degiro |  

**Note:** The file **SHOULD** only contain the values, the header is just for better understanding.

* Run the command
    
    ```bash
    $ market-manager account import wallet
    ```

<br />[[table of contents]](#table-of-contents)

#### Import transfer

Add/Create transfers

* Add/Create `csv` file to `resources/import/transfers` with the transfers(s) with the following format:
    
    | DATE      | FROM | TO     | AMOUNT |
    |-----------|------|--------|--------|
    | 6/10/2017 | Bank | Wallet | 20,00  |
    
    **FROM**: Name of the bank account. This values is used to match with our wallet in case you transfer money out from the wallet. 
    
    **TO**: Name of the bank account. This values is used to match with our wallet in case you transfer money in to the wallet.  

**Note:** The file **SHOULD** only contain the values, the header is just for better understanding.

* Run the command
    
    ```bash
    $ market-manager banking import transfer
    ```

<br />[[table of contents]](#table-of-contents)

## Dependencies
 
### Upstream

This application depends on the following sites:
* [finance.yahoo.com](https://finance.yahoo.com/): Used to scrape stock company profile and stock summary.
* [marketchameleon.com/](https://marketchameleon.com/:): Used to scrape stock dividends and stock volatility.
* [free.currencyconverterapi.com](http://free.currencyconverterapi.com): Used to get EUR to USD and EUR to Dollar Canadian exchange prices on real time.

<br />[[table of contents]](#table-of-contents)
