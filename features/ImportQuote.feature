Feature:
  As a user
  I want to upload stocks
  so that I can operate with the stocks

  Scenario: Import stocks
    When I add a new csv file "01_stocks.csv" to the "stock" import folder with the following lines:

      | Name                                  | Exchange | Symbol | Type   | Sector          | Industry                   |
      | STARBUCKS CORPORATION                 | NASDAQ   | SBUX   | COMMON | SERVICES        | SPECIALTY EATERIES         |
      | INTEL CORPORATION - CO                | NASDAQ   | INTC   | COMMON | TECHNOLOGY      | SEMICONDUCTOR - BROAD LINE |
      | SCOTTS MIRACLE-GRO COMPANY            | NYSE     | SMG    | COMMON | BASIC MATERIALS | AGRICULTURAL CHEMICALS     |
      | WALMART INC                           | NYSE     | WMT    | COMMON | SERVICES        | DISCOUNT; VARIETY STORES   |
      | ELI LILLY AND COMPANY                 | NYSE     | LLY    | COMMON | HEALTHCARE      | DRUG MANUFACTURERS - MAJOR |
      | QUALCOMM INCORPORATED                 | NASDAQ   | QCOM   | COMMON | TECHNOLOGY      | COMMUNICATION EQUIPMENT    |
      | BROOKFIELD INFRASTRUCTURE PARTNERS LP | NYSE     | BIP    | MLP    | UTILITIES       | DIVERSIFIED UTILITIES      |

    And I run a command "market-manager" with args "purchase import stock":
    Then following stocks info should be stored:
      | id | name                       | type     |
      | 1  | COMMON                     | type     |
      | 2  | MLP                        | type     |
      | 3  | SERVICES                   | sector   |
      | 4  | TECHNOLOGY                 | sector   |
      | 5  | BASIC MATERIALS            | sector   |
      | 6  | HEALTHCARE                 | sector   |
      | 7  | UTILITIES                  | sector   |
      | 8  | SPECIALTY EATERIES         | industry |
      | 9  | SEMICONDUCTOR - BROAD LINE | industry |
      | 10 | AGRICULTURAL CHEMICALS     | industry |
      | 11 | DISCOUNT; VARIETY STORES   | industry |
      | 12 | DRUG MANUFACTURERS - MAJOR | industry |
      | 13 | COMMUNICATION EQUIPMENT    | industry |
      | 14 | DIVERSIFIED UTILITIES      | industry |

    And following stocks should be stored:

      | id | name                                  | exchange_symbol | symbol | id_type | id_sector | id_industry |
      | 1  | STARBUCKS CORPORATION                 | NASDAQ          | SBUX   | 1       | 3         | 8           |
      | 2  | INTEL CORPORATION - CO                | NASDAQ          | INTC   | 1       | 4         | 9           |
      | 3  | SCOTTS MIRACLE-GRO COMPANY            | NYSE            | SMG    | 1       | 5         | 10          |
      | 4  | WALMART INC                           | NYSE            | WMT    | 1       | 3         | 11          |
      | 5  | ELI LILLY AND COMPANY                 | NYSE            | LLY    | 1       | 6         | 12          |
      | 6  | QUALCOMM INCORPORATED                 | NASDAQ          | QCOM   | 1       | 4         | 13          |
      | 7  | BROOKFIELD INFRASTRUCTURE PARTNERS LP | NYSE            | BIP    | 2       | 7         | 14          |
