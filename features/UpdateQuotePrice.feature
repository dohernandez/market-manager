Feature:
  As a user
  I want to update the price of all the stocks
  so that I can operate with the wallet with the stock price updated

  Background:
    Given that the following stocks info are stored:
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

      | id                                   | name                                  | exchange_symbol | market_name | symbol | value | type                                 | sector                               | industry                             | last_price_update | high_low_52_week_update |
      | 5b273d3c-9a0f-4abc-9606-fe1165ee1d79 | STARBUCKS CORPORATION                 | NASDAQ          | stock       | SBUX   | 52.22 | 17635c1f-9058-4cd2-8f8b-e193db69cccf | b914a9a8-9bf5-4ecb-a587-c300e477363a | 64038365-888d-42a6-8ec5-3f868e374b53 | 04/1/2018         | 04/1/2018               |
      | 719e2ce6-0a9e-4602-ba23-16dcae2fe963 | INTEL CORPORATION - CO                | NASDAQ          | stock       | INTC   | 53.46 | 17635c1f-9058-4cd2-8f8b-e193db69cccf | 508b0f62-81a8-42ae-998a-d236a3eb1040 | 1c5b334d-2b3e-430a-8342-30203a8ea956 | 04/1/2018         | 04/1/2018               |
      | 74099645-a1c0-4931-8193-366ef14cbf20 | SCOTTS MIRACLE-GRO COMPANY            | NYSE            | stock       | SMG    | 79.74 | 17635c1f-9058-4cd2-8f8b-e193db69cccf | 30fed46b-94ca-4ca7-9419-1d66a3f67e03 | 8b8e6b11-e537-4d80-afdc-95dbedb48f2f | 04/1/2018         | 04/1/2018               |
      | a17bb930-64f0-4d02-987d-ce65f5ea2ddd | WALMART INC                           | NYSE            | stock       | WMT    | 83.61 | 17635c1f-9058-4cd2-8f8b-e193db69cccf | b914a9a8-9bf5-4ecb-a587-c300e477363a | dd6ae645-f324-4989-9c96-9e26f1ae298d | 04/1/2018         | 04/1/2018               |
      | bb19897e-49c9-40ba-900b-f966160a0ab1 | ELI LILLY AND COMPANY                 | NYSE            | stock       | LLY    | 86.28 | 17635c1f-9058-4cd2-8f8b-e193db69cccf | 34defa05-f0db-4808-8978-a022ecd7adea | 4b6f5da6-de17-4066-876f-3c158e0ce695 | 04/1/2018         | 04/1/2018               |
      | c1469b77-c18a-44e3-9ebd-c88221add9f8 | QUALCOMM INCORPORATED                 | NASDAQ          | stock       | QCOM   | 58.79 | 17635c1f-9058-4cd2-8f8b-e193db69cccf | 508b0f62-81a8-42ae-998a-d236a3eb1040 | ac9d5a3c-014a-4507-afb9-c1f44566d7ac | 04/1/2018         | 04/1/2018               |
      | e1c7a018-0107-4e70-9353-a043f8dd0a57 | BROOKFIELD INFRASTRUCTURE PARTNERS LP | NYSE            | stock       | BIP    | 39.69 | 1ed71ccc-b9ed-426c-93c4-25fac1ce5995 | c243d304-ddd9-4a4a-9a7a-5f0912930fb9 | 4fdfd84d-43cf-4234-ac0e-c42df60de5bb | 04/1/2018         | 04/1/2018               |

    And that we get the following stocks price from yahoo finance:

      | Stock | Date       | Open  | High  | Low   | Close | Adj Close | Volume   |
      | SBUX  | 2018-06-21 | 52.29 | 52.63 | 50.36 | 50.62 | 50.62     | 30723000 |
      | INTC  | 2018-06-21 | 54.38 | 54.53 | 51.94 | 52.19 | 52.19     | 44435600 |
      | SMG   | 2018-06-21 | 80.82 | 81.27 | 78.75 | 79.74 | 79.74     | 1593100  |
      | WMT   | 2018-06-21 | 79.73 | 81.69 | 79.67 | 81.48 | 81.48     | 724000   |
      | LLY   | 2018-06-21 | 85.82 | 86.28 | 85.19 | 86.07 | 86.07     | 2856400  |
      | QCOM  | 2018-06-21 | 59.21 | 59.21 | 58.39 | 58.75 | 58.75     | 6605400  |
      | BIP   | 2018-06-21 | 39.65 | 39.74 | 38.71 | 38.84 | 38.84     | 240000   |

  Scenario: Update price from yahoo finance
    When I run a command "market-manager" with args "purchase update price":
    Then following stocks should be stored:

      | id                                   | exchange_symbol | symbol | value | change |
      | 5b273d3c-9a0f-4abc-9606-fe1165ee1d79 | NASDAQ          | SBUX   | 50.62 | -1.67  |
      | 719e2ce6-0a9e-4602-ba23-16dcae2fe963 | NASDAQ          | INTC   | 52.19 | -2.19  |
      | 74099645-a1c0-4931-8193-366ef14cbf20 | NYSE            | SMG    | 79.74 | -1.08  |
      | a17bb930-64f0-4d02-987d-ce65f5ea2ddd | NYSE            | WMT    | 81.48 | 1.75   |
      | bb19897e-49c9-40ba-900b-f966160a0ab1 | NYSE            | LLY    | 86.07 | 0.25   |
      | c1469b77-c18a-44e3-9ebd-c88221add9f8 | NASDAQ          | QCOM   | 58.75 | -0.46  |
      | e1c7a018-0107-4e70-9353-a043f8dd0a57 | NYSE            | BIP    | 38.84 | -0.81  |
