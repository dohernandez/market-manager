Feature:
  As a user
  I want to upload transfer
  so that I can see how much are invested in the wallet

  Background:
    Given that the following wallets are stored:

      | id                                   | name   | url           |
      | b0c650ff-130e-4460-bcd9-423643ee314d | wallet | www.degiro.es |

  Scenario: Import transfers
    When I add a new csv file "01_transfer.csv" to the "transfer" import folder with the following lines

      | Date       | From    | To     | Amount |
      | 6/10/2017  | Triodos | Degiro | 20,00  |
      | 15/10/2017 | Triodos | Degiro | 600,00 |
      | 6/11/2017  | Triodos | Degiro | 100,00 |

    And I run a command "market-manager" with args "banking import transfer"
    Then following transfers should be stored:
      | id | name   | url           |
      | 1  | wallet | www.degiro.es |
