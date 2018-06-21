Feature:
  As a user
  I want to upload wallet
  so that I can operate with the wallet

  Scenario: Import wallets
    When I add a new csv file "wallet.csv" to the "wallet" import folder with the following lines

      | URL           | Name   |
      | www.degiro.es | Degiro |

    And I run a command "market-manager" with args "account import wallet"
    Then following wallets should be stored:
      | id | name   | url           |
      | 1  | wallet | www.degiro.es |
