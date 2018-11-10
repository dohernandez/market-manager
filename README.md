# Market Manager

An application to manage the bull market allowing us to keep tracking the prices and dividends of the stocks we prefer, from any market in the world.
               
Also, allow us to manage our own wallet, storing all the operations we have made used.

## Table of Contents

* [Resources](#resources)
* [Commands](#commands)
    * [Import tools](#import-tools)
        * [Stocks](#import-stocks)
        * [Wallets](#import-wallet)
        * [Transfers](#import-transfer)
        * [Operations](#import-operations)
        * [Retentions](#import-retention)
    * [Add tools](#add-tools)
        * [Stock](#add-stock)
        * [Wallet](#add-wallet)
        * [Transfer](#add-transfer)
        * [Operation](#add-operation)
            * [Buy](#add-operation-buy)
            * [Sell](#add-operation-sell)
            * [Dividend](#add-operation-dividend)
            * [Interest](#add-operation-interest)
        * [Retention](#add-retention)
* [Getting started](#getting-started)
    * [Prerequisites](#prerequisites)
    * [Testing](#testing)
* [Dependencies](#dependencies)
    * [Upstream](#upstream)
* [Copyright](#copyright)


## Resources
<br />[[table of contents]](#table-of-contents)

## Commands

### Import tools

#### Import stocks

    ```bash
    market-manager purchase import stock -h
    ```
    
* Add/Create `xx_stocks.csv` file to `resources/import/stocks` with the stock(s) with the following format:
    
    | STOCK NAME         | EXCHANGE SYMBOL | SYMBOL | TYPE   | Sector     | Industry                    |
    |--------------------|-----------------|--------|--------|------------|-----------------------------|
    | NVIDIA CORPORATION | NASDAQ          | NVDA   | COMMON | TECHNOLOGY | SEMICONDUCTOR - SPECIALIZED |  

**Note:** The file **SHOULD** only contain the values, the header is just for better understanding.

* Run the command
    
    ```bash
    market-manager purchase import stock
    ```

<br />[[table of contents]](#table-of-contents)

#### Import wallet

    ```bash
    market-manager account import wallet -h
    ```

* Add/Create `xx_ourwallet.csv` file to `resources/import/wallets` with the wallet(s) with the following format:
    
    | URL            |  NAME  |
    |----------------|--------|
    | www.wallet.es  | Degiro |  

**Note:** The file **SHOULD** only contain the values, the header is just for better understanding.

* Run the command
    
    ```bash
    market-manager account import wallet
    ```

<br />[[table of contents]](#table-of-contents)

#### Import transfer

    ```bash
    market-manager banking import transfer -h
    ```

* Add/Create `xx_transfer.csv` file to `resources/import/transfers` with the transfers(s) with the following format:
    
    | DATE      | FROM | TO     | AMOUNT |
    |-----------|------|--------|--------|
    | 6/10/2017 | Bank | Wallet | 20,00  |
    
    **FROM**: Name of the bank account. This values is used to match with our wallet in case you transfer money out from the wallet. 
    
    **TO**: Name of the bank account. This values is used to match with our wallet in case you transfer money in to the wallet.  

**Note:** The file **SHOULD** only contain the values, the header is just for better understanding.

* Run the command
    
    ```bash
    market-manager banking import transfer
    ```

<br />[[table of contents]](#table-of-contents)

#### Import retention

    ```bash
    market-manager account import stock-retention -h
    ```

* Add/Create `xx_ourwallet.csv` file to `resources/import/retentions` with the retention(s) with the following format:
    
    | DATE      | FROM | TO     | AMOUNT |
    |-----------|------|--------|--------|
    | 6/10/2017 | Bank | Wallet | 20,00  |
    
    **FROM**: Name of the bank account. This values is used to match with our wallet in case you transfer money out from the wallet. 
    
    **TO**: Name of the bank account. This values is used to match with our wallet in case you transfer money in to the wallet.  

**Note:** The file **SHOULD** only contain the values, the header is just for better understanding.

* Run the command
    
    ```bash
    market-manager account import stock-retention -w ourwallet
    ```

<br />[[table of contents]](#table-of-contents)

#### Import operations

    ```bash
    market-manager account import operation -h
    ```

* Add/Create `xx_ourwallet.csv` file to `resources/import/accounts` with the operation(s) with the following format:
    
    | TRADE # | DATE      | STOCK NAME       | TYPE    | AMOUNT | PRICE    | EXCHANGE | E. FEES | PAYED   | FEES   |
    |---------|-----------|------------------|---------|--------|----------|----------|---------|---------|--------|
    | 1       | 6/10/2017 | TESLA MOTORS INC | Comprar | 2      | "332,69" | "1,1848" | "0,16"  | "88,25" | "0,51" |
    
    **EXCHANGE**: Exchange currency to EUR. 
    
**Note:** The file **SHOULD** only contain the values, the header is just for better understanding.

* Run the command
    
    ```bash
    market-manager account import operation -w ourwallet
    ```

<br />[[table of contents]](#table-of-contents)

### Add tools

#### Add stock

    ```bash
    market-manager account add stock -h
    ```
    
*Example of used

    ```bash
        market-manager purchase add stock -s ET -e NYSE
    ```

<br />[[table of contents]](#table-of-contents)

#### Add operation

    ```bash
    market-manager account add operation -h
    ```

##### Add operation buy

    ```bash
    market-manager account add operation buy -h
    ```
    
*Example of used

    ```bash
        market-manager account add operation buy -w ourwallet -t 83 -d 16/10/2018 -s DRAD -a 150 -p 1.4 -pc 1.1579 -pcc 0.16 -v 181.54 -c 1.02
    ```

**Note:** The option trade is use to match with the sell operation in order to track the performance of a single operation.

<br />[[table of contents]](#table-of-contents)

##### Add operation sell

    ```bash
    market-manager account add operation sell -h
    ```
    
*Example of used

    ```bash
        market-manager account add operation sell -w ourwallet -t 12 -d 04/10/2018 -s JD -a 2 -p 24.78 -pc 1.1526 -pcc 0.16 -v 42.96 -c 0.51
    ```

**Note:** The option trade is use to match with the buy operation in order to track the performance of a single operation.

<br />[[table of contents]](#table-of-contents)

##### Add operation dividend

    ```bash
    market-manager account add operation dividend -h
    ```
    
*Example of used

    ```bash
        market-manager account add operation dividend -w ourwallet -d 28/09/2018 -s HUN -v 2.37
    ```

<br />[[table of contents]](#table-of-contents)

##### Add operation interest

    ```bash
    market-manager account add operation interest -h
    ```
    
*Example of used

    ```bash
        market-manager account add operation interest -w ourwallet -d 30/09/2018 -v 13.92
    ```

<br />[[table of contents]](#table-of-contents)

#### Add retention

    ```bash
    market-manager account add dividend-retention -h
    ```
    
*Example of used

    ```bash
        market-manager account add dividend-retention-w ourwallet -s HUN -r 0.37
    ```

<br />[[table of contents]](#table-of-contents)

## Getting started

<br />[[table of contents]](#table-of-contents)

## Dependencies
 
### Upstream

This application depends on the following sites:
* [finance.yahoo.com](https://finance.yahoo.com/): Used to scrape stock company profile and stock summary.
* [marketchameleon.com/](https://marketchameleon.com/:): Used to scrape stock dividends and stock volatility.
* [free.currencyconverterapi.com](http://free.currencyconverterapi.com): Used to get EUR to USD and EUR to Dollar Canadian exchange prices on real time.

<br />[[table of contents]](#table-of-contents)

### Copyright

Copyright (C) 2018 by Darien Hernandez <dohernandez@gmail.com>.

Market Manager released under MIT License.
See [LICENSE](https://github.com/dohernandez/market-manager/blob/master/LICENSE) for details.
