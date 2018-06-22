Feature:
  As a user
  I want to upload operations
  so that I can see wallet's details

  Background:
    Given that the following bank accounts are stored:

      | id                                   | name  | account_no                  | alias  | account_no_type |
      | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | Bank1 | DE27 1007 7777 0209 2997 00 | Bank1  | iban            |
      | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd | Bank2 | DE11 5205 1373 5120 7101 31 | Wallet | iban            |

    And that the following wallets are stored:

      | id                                   | name   | url           | invested | funds |
      | b0c650ff-130e-4460-bcd9-423643ee314d | wallet | www.wallet.es | 720      | 720   |

    And that the following wallet bank accounts are stored:

      | wallet_id                            | bank_account_id                      |
      | b0c650ff-130e-4460-bcd9-423643ee314d | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd |

    And that the following transfers are stored:

      | id                                   | from_account                         | to_account                           | amount | date       |
      | 43e3d1d9-f60f-414b-9ade-a6f2f78db28a | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd | 20,00  | 6/10/2017  |
      | a65ab78e-006a-475c-889d-3919ede5dab2 | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd | 600,00 | 15/10/2017 |
      | e1e4124f-056d-4342-b4b8-8ec965f38691 | f8e0d0b7-8b5f-46e2-a65a-aae8a7ef5b63 | cf9ef982-a9c1-4c7b-bbd3-011ed81f1bbd | 100,00 | 6/11/2017  |

    And that the following stocks info are stored:
      | id                                   | name                       | type     |
      | 17635c1f-9058-4cd2-8f8b-e193db69cccf | COMMON                     | type     |
      | 1ed71ccc-b9ed-426c-93c4-25fac1ce5995 | MLP                        | type     |
      | b914a9a8-9bf5-4ecb-a587-c300e477363a | SERVICES                   | sector   |
      | 508b0f62-81a8-42ae-998a-d236a3eb1040 | TECHNOLOGY                 | sector   |
      | 30fed46b-94ca-4ca7-9419-1d66a3f67e03 | BASIC MATERIALS            | sector   |
      | 34defa05-f0db-4808-8978-a022ecd7adea | HEALTHCARE                 | sector   |
      | c243d304-ddd9-4a4a-9a7a-5f0912930fb9 | UTILITIES                  | sector   |
      | 64038365-888d-42a6-8ec5-3f868e374b53 | SPECIALTY EATERIES         | industry |
      | 1c5b334d-2b3e-430a-8342-30203a8ea956 | SEMICONDUCTOR - BROAD LINE | industry |
      | 8b8e6b11-e537-4d80-afdc-95dbedb48f2f | AGRICULTURAL CHEMICALS     | industry |
      | dd6ae645-f324-4989-9c96-9e26f1ae298d | DISCOUNT; VARIETY STORES   | industry |
      | 4b6f5da6-de17-4066-876f-3c158e0ce695 | DRUG MANUFACTURERS - MAJOR | industry |
      | ac9d5a3c-014a-4507-afb9-c1f44566d7ac | COMMUNICATION EQUIPMENT    | industry |
      | 4fdfd84d-43cf-4234-ac0e-c42df60de5bb | DIVERSIFIED UTILITIES      | industry |

    And that the following stocks are stored:

      | id                                   | name                                  | exchange_symbol | market_name | symbol | type                                 | sector                               | industry                             | last_price_update | high_low_52_week_update |
      | 5b273d3c-9a0f-4abc-9606-fe1165ee1d79 | STARBUCKS CORPORATION                 | NASDAQ          | stock       | SBUX   | 17635c1f-9058-4cd2-8f8b-e193db69cccf | b914a9a8-9bf5-4ecb-a587-c300e477363a | 64038365-888d-42a6-8ec5-3f868e374b53 | 04/1/2018         | 04/1/2018               |
      | 719e2ce6-0a9e-4602-ba23-16dcae2fe963 | INTEL CORPORATION - CO                | NASDAQ          | stock       | INTC   | 17635c1f-9058-4cd2-8f8b-e193db69cccf | 508b0f62-81a8-42ae-998a-d236a3eb1040 | 1c5b334d-2b3e-430a-8342-30203a8ea956 | 04/1/2018         | 04/1/2018               |
      | 74099645-a1c0-4931-8193-366ef14cbf20 | SCOTTS MIRACLE-GRO COMPANY            | NYSE            | stock       | SMG    | 17635c1f-9058-4cd2-8f8b-e193db69cccf | 30fed46b-94ca-4ca7-9419-1d66a3f67e03 | 8b8e6b11-e537-4d80-afdc-95dbedb48f2f | 04/1/2018         | 04/1/2018               |
      | a17bb930-64f0-4d02-987d-ce65f5ea2ddd | WALMART INC                           | NYSE            | stock       | WMT    | 17635c1f-9058-4cd2-8f8b-e193db69cccf | b914a9a8-9bf5-4ecb-a587-c300e477363a | dd6ae645-f324-4989-9c96-9e26f1ae298d | 04/1/2018         | 04/1/2018               |
      | bb19897e-49c9-40ba-900b-f966160a0ab1 | ELI LILLY AND COMPANY                 | NYSE            | stock       | LLY    | 17635c1f-9058-4cd2-8f8b-e193db69cccf | 34defa05-f0db-4808-8978-a022ecd7adea | 4b6f5da6-de17-4066-876f-3c158e0ce695 | 04/1/2018         | 04/1/2018               |
      | c1469b77-c18a-44e3-9ebd-c88221add9f8 | QUALCOMM INCORPORATED                 | NASDAQ          | stock       | QCOM   | 17635c1f-9058-4cd2-8f8b-e193db69cccf | 508b0f62-81a8-42ae-998a-d236a3eb1040 | ac9d5a3c-014a-4507-afb9-c1f44566d7ac | 04/1/2018         | 04/1/2018               |
      | e1c7a018-0107-4e70-9353-a043f8dd0a57 | BROOKFIELD INFRASTRUCTURE PARTNERS LP | NYSE            | stock       | BIP    | 1ed71ccc-b9ed-426c-93c4-25fac1ce5995 | c243d304-ddd9-4a4a-9a7a-5f0912930fb9 | 4fdfd84d-43cf-4234-ac0e-c42df60de5bb | 04/1/2018         | 04/1/2018               |

  Scenario: Import operations
    When I add a new csv file "01_wallet.csv" to the "accounts" import folder with the following lines:

      | # | Date       | Stock                 | Action | Amount | Price  | Price Change | Price Change Commission | Value  | Commission |
      | 1 | 04/12/2017 | STARBUCKS CORPORATION | Compra | 4      | 177,50 | 1,1840       | 0,16                    | 599,82 | 0,51       |
      | 2 | 04/1/2018  | QUALCOMM INCORPORATED | Compra | 8      | 44,50  | 1,1640       | 0,16                    | 399,82 | 0,54       |

    And I run a command "market-manager" with args "account import operation":
    Then the following wallets should have:

      | id                                   | invested | funds   | commission |
      | b0c650ff-130e-4460-bcd9-423643ee314d | 720.00   | -281.01 | 1.37       |

    And the following operations should be stored:
      | id | wallet_id                            | date       | stock_id                             | action | amount | price | price_change | price_change_commission | value  | commission |
      | 1  | b0c650ff-130e-4460-bcd9-423643ee314d | 04/1/2018  | c1469b77-c18a-44e3-9ebd-c88221add9f8 | buy    | 8      | 44.5  | 1.164        | 0.16                    | 399.82 | 0.54       |
      | 2  | b0c650ff-130e-4460-bcd9-423643ee314d | 04/12/2017 | 5b273d3c-9a0f-4abc-9606-fe1165ee1d79 | buy    | 4      | 177.5 | 1.184        | 0.16                    | 599.82 | 0.51       |

    And the following wallet items should be stored:

      | id | wallet_id                            | stock_id                             | amount | invested | dividend | buys   | sells | capital | capital_rate |
      | 1  | b0c650ff-130e-4460-bcd9-423643ee314d | c1469b77-c18a-44e3-9ebd-c88221add9f8 | 8      | 400.52   | 0.0      | 400.52 | 0.0   | 0.0     | 1.1654       |
      | 2  | b0c650ff-130e-4460-bcd9-423643ee314d | 5b273d3c-9a0f-4abc-9606-fe1165ee1d79 | 4      | 600.49   | 0.0      | 600.49 | 0.0   | 0.0     | 1.1654       |

