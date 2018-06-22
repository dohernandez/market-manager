Feature:
  As a user
  I want to upload wallet
  so that I can operate with the wallet

  Background:
    Given that the following bank accounts are stored:

      | id                                   | name  | account_no                  | alias  | account_no_type |
      | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | Bank1 | DE27 1007 7777 0209 2997 00 | Wallet | iban            |

  Scenario: Import wallets
    When I add a new csv file "wallet.csv" to the "wallet" import folder with the following lines:

      | URL           | Name   |
      | www.wallet.es | Wallet |

    And I run a command "market-manager" with args "account import wallet":
    Then following wallets should be stored:
      | id | name   | url           |
      | 1  | wallet | www.wallet.es |
