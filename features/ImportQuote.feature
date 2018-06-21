Feature:
  As a user
  I want to upload stocks
  so that I can operate with the stocks

  Background:
    Given that the following exchanges are stored:

      | id                                   | name                        | symbol |
      | 515e0d12-0b9f-412a-9136-d7875ce718f2 | New York Stock Exchange     | NYSE   |
      | 85e11c4e-cc5a-4cbf-956b-48037be9158f | Bolsas y Mercados Españoles | BME    |
      | 8936b1e8-9955-4950-a2f0-3f022f64b7a7 | Frankfurter Wertpapierbörse | FRA    |
      | a9787ff0-fe23-4453-b625-875d12d17450 | NASDAQ COMPOSITE            | NASDAQ |
      | ba21f439-f0f3-4598-a77f-eb3210142774 | Borsa Italiana S.p. A       | BIT    |

    And that the following markets are stored:
      | id                                   | name           | display_name   |
      | 18dc5008-a066-4f83-9a1a-be75b04756a1 | cryptocurrency | Cryptocurrency |
      | 8cc8f7ca-43f2-4f4f-84d7-aac4efdbc173 | stock          | Stock          |

  Scenario: Import stocks
    When I add a new csv file "01_stocks.csv" to the "stock" import folder with the following lines

      | Name                                  | Exchange | Symbol | Type   | Sector          | Industry                   |
      | STARBUCKS CORPORATION                 | NASDAQ   | SBUX   | COMMON | SERVICES        | SPECIALTY EATERIES         |
      | INTEL CORPORATION - CO                | NASDAQ   | INTC   | COMMON | TECHNOLOGY      | SEMICONDUCTOR - BROAD LINE |
      | SCOTTS MIRACLE-GRO COMPANY            | NYSE     | SMG    | COMMON | BASIC MATERIALS | AGRICULTURAL CHEMICALS     |
      | WALMART INC                           | NYSE     | WMT    | COMMON | SERVICES        | DISCOUNT; VARIETY STORES   |
      | ELI LILLY AND COMPANY                 | NYSE     | LLY    | COMMON | HEALTHCARE      | DRUG MANUFACTURERS - MAJOR |
      | QUALCOMM INCORPORATED                 | NASDAQ   | QCOM   | COMMON | TECHNOLOGY      | COMMUNICATION EQUIPMENT    |
      | BROOKFIELD INFRASTRUCTURE PARTNERS LP | NYSE     | BIP    | MLP    | UTILITIES       | DIVERSIFIED UTILITIES      |

    And I run a command "market-manager" with args "purchase import quote"
    Then following stock info should be stored:
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

      | id | name                                  | exchange_id                          | symbol | type | sector | industry |
      | 1  | STARBUCKS CORPORATION                 | a9787ff0-fe23-4453-b625-875d12d17450 | SBUX   | 1    | 3      | 8        |
      | 2  | INTEL CORPORATION - CO                | a9787ff0-fe23-4453-b625-875d12d17450 | INTC   | 1    | 4      | 9        |
      | 3  | SCOTTS MIRACLE-GRO COMPANY            | 515e0d12-0b9f-412a-9136-d7875ce718f2 | SMG    | 1    | 5      | 10       |
      | 4  | WALMART INC                           | 515e0d12-0b9f-412a-9136-d7875ce718f2 | WMT    | 1    | 3      | 11       |
      | 5  | ELI LILLY AND COMPANY                 | 515e0d12-0b9f-412a-9136-d7875ce718f2 | LLY    | 1    | 6      | 12       |
      | 6  | QUALCOMM INCORPORATED                 | a9787ff0-fe23-4453-b625-875d12d17450 | QCOM   | 1    | 4      | 13       |
      | 7  | BROOKFIELD INFRASTRUCTURE PARTNERS LP | 515e0d12-0b9f-412a-9136-d7875ce718f2 | BIP    | 2    | 7      | 14       |
