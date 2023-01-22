# Beancount CSV Parser

This an importer tool is to help importing CSV files into beancount format. The tool was inspired by https://github.com/Sudneo/swed2beancount which was written for Swedbank CSV export.  I rewrote it to make a bit more generic and something I can use for converting downloads from various financial institutions to beancode ledger format.

## Features

* there are two applications:
    * csv2bean - imports simple cash account transactions.csv files and prepares an output.bean file in beancount format (that can be copied and pasted to a beancount journal file)
    * csv2inv - similar to csv2bean, but handles investement securities as well as cash transactions
* Both applications assume three .csv files:
    * "config.csv" - contains date format, primary "from" account, currency and a boolean if amounts need to have their sign reversed (typical for credit card account downloads)
    * "transactions.csv" or "transactions-inv.csv" - contains a minimum of date, amount and description fields, with a header row
    * "match.csv" or "match-inv.csv" - maps CSV transaction with Beancount accounts
* csv2bean provides a simple text match or partial match of details/description field
* csv2inv provides more complex matching to handle different investment securities
    * BUY or SELL match for investment transations
    * to and from account names based on "match-inv.csv"
* Both applications generate a beancount compatible file, "output.bean", which can be copied to a main ledger file

## How to Use

1. Prepare "transactions.csv" or "transactions-inv.csv" file
    - Download .csv file from Financial Institution (Bank, Brokerage, Credit Card) for the time period required.  
    - Open in a spreadsheet program like Google Sheets, although could use any spreadsheet program like Excel or Numbers, that can import a csv file.  
    - Modify or add single header row and add or rename the following fields descriptions: date,description, amount, symbol, quantity, price, commission, fee.  
    - For simple cash transactions, add "0" or "" entry for unused fields
    - Delete any unnecessary rows
    - Export revised csv worksheet to a file named "transactions.csv" or "transactions-inv.csv" in root beancount directory

2. Prepare "config.csv" File
    The program looks for a "config.csv" file which contains a single row with the following fields::
    - DateFormat: <date format> eg 2016-01-02
    - Default "fromAccount" eg Assets:BOA:Checking
    - currency: <currency> eg USD
    - reverse: "TRUE" if account amounts need to be opposite sign 

    This may be different for every institution, so I usually make a config file for each institution for which I have csv files. I find that the easiest way to do this is by using a spreadsheet program and then download it to a .csv file. Before running the importer, I then copy to a file named "config.csv" in my root beancount directory.

3. Prepare "match.csv" or "match-inv.csv" file:
    To help assigning accounts to transactions, I make a rules mapping file for each downloaded .csv file, which contains a headings row and a row for each matching condition.  For the simple case, only two columns are required: "account2" and "contains" description to match.  For more complex investment accounts, the following 4 columns are required:
    - "fromAccount": name of account from which funds to be drawn if different from the default account, in case of a match
    - "toAccount": name of the account to which funds will be added if a match record eg Expenses: Groceries which will overide the default to account
    - "trade": either "BUY" or "SELL" to determine trading direction ie whether the to and from Accounts contains the security in case of a match.  Assumes currency only if left blank
    - "contains": text string to "match" in record "description" field
    
    Again, this may be different for every institution, so I usually make a match file for each institution for which I regularly download csv files. I find that the easiest way to do this is by using a spreadsheet program and then download it to a .csv file. Before running the importer, I then copy to "match.csv" or "match-inv.csv" which is recognized by the program, and make sure it is in my beancount directory.

4. The program is written in "go" and is best compiled and placed in an accessible directory which is included in the users default PATH.  It may then be run from the beancount root directory and will produce an output file, "output.bean", in that directory.  You can then copy and paste to your journal file, on change its name and add an include statement in the journal.

## Download

Tool may be run directly from the binary file or compiled from source

```bash
git clone https://github.com/paulstek/csv2bc
cd csv2bc
# for simple cash downloads:
go install csv2bean.go
# for more complex investment account downloads:
go install csv2inv.go

```

Once binary file is properly installed, made executable and path set, the program may be executed as follows from root beancount directory:

```bash
csv2bean
# or
csv2inv
```
