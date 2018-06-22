Feature:
  As a user
  I want to upload transfer
  so that I can see how much are invested in the wallet

  Background:
    Given that the following bank accounts are stored:

      | id                                   | name  | account_no                  | alias  | account_no_type |
      | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | Bank1 | DE27 1007 7777 0209 2997 00 | Bank1  | iban            |
      | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd | Bank2 | DE11 5205 1373 5120 7101 31 | Wallet | iban            |

    And that the following wallets are stored:

      | id                                   | name   | url           |
      | b0c650ff-130e-4460-bcd9-423643ee314d | wallet | www.wallet.es |

    And that the following wallet bank accounts are stored:

      | wallet_id                            | bank_account_id                      |
      | b0c650ff-130e-4460-bcd9-423643ee314d | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd |

  Scenario: Import transfers
    When I add a new csv file "01_transfer.csv" to the "transfer" import folder with the following lines:

      | Date       | From  | To     | Amount |
      | 6/10/2017  | Bank1 | Wallet | 20,00  |
      | 15/10/2017 | Bank1 | Wallet | 600,00 |
      | 6/11/2017  | Bank1 | Wallet | 100,00 |

    And I run a command "market-manager" with args "banking import transfer":
    Then following transfers should be stored:

      | id | from_account                         | to_account                           | amount | date       |
      | 1  | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd | 20,00  | 6/10/2017  |
      | 2  | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd | 600,00 | 15/10/2017 |
      | 3  | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd | 100,00 | 6/11/2017  |

    And the wallets should have:

      | id                                   | invested |
      | b0c650ff-130e-4460-bcd9-423643ee314d | 720.00   |
