package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// ******************************* DEFINE STRUCT's Config, Transaction and "Match
// Config represents the configuration for the program, including the date format
// for the input CSV file.
type Config struct {
	dateFormat string
	account1   string
	currency   string
	reverse    bool
}

// Transaction represents a financial transaction from the input CSV file.
type Transaction struct {
	date        time.Time
	description string
	symbol      string
	price       float64
	commission  float64
	fee         float64
	amount      float64
}

// Match represents a description-to-account match from the "match.csv" file.
type Match struct {
	account2    string
	contains    string
}

// ******************************* MAIN FUNCTION
func main() {
	// Read the configuration from the config CSV file
	config, err := readConfig("config.csv")
	if err != nil {
		log.Fatal(err)
	}
	// Read the transactions from the input CSV file, prepare tranaction string
	transactions, err := readTransactions("transactions.csv", config.dateFormat)
	if err != nil {
		log.Fatal(err)
	}
	// Write the transaction string to the output file in Beancount format
	if err := writeBeancount("output.bean", transactions, config.account1, config.currency, config.reverse); err != nil {
		log.Fatal(err)
	}
}

// ******************************* func readConfig
// readConfig reads the configuration from a single line CSV file.
func readConfig(filename string) (*Config, error) {
	// Open and read the configuration file; single row only
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Config file: %s", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	record, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read Config record: %s", err)
	}
	// Parse the "reverse" field as a boolean
	reverse := false
	if record[3] == "Y" {
		reverse = true
	}
	// Return the parsed Config struct
	return &Config{
		dateFormat: record[0],
		account1:   record[1],
		currency:   record[2],
		reverse:    reverse,
	}, nil
}

// ******************************* func readTransactions

// readTransactions reads the transaction records from a CSV file, assumes header row.
// CSV file contains at least Date, Description and Amount columns.
// Order doesnt matter, functions finds the record index for these three columns,

func readTransactions(filename, dateFormat string) ([]*Transaction, error) {
	// Open the input CSV file and read header - first row
	inFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()
	reader := csv.NewReader(inFile)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Find the indexes of the Date, Description and Amount columns
	dateIndex := -1
	descriptionIndex := -1
	amountIndex := -1
	symbolIndex := -1
	priceIndex := -1
	commissionIndex := -1
	feeIndex := -1

	for i, col := range header {
		switch col {
		case "date":
			dateIndex = i
		case "description":
			descriptionIndex = i
		case "amount":
			amountIndex = i
		case "symbol":
			symbolIndex = i
		case "price":
			priceIndex = i
		case "commission":
			commissionIndex = i
		case "fee":
			feeIndex = i
		}
	}

	// Make sure we found all the necessary columns
	if dateIndex == -1 || descriptionIndex == -1 || amountIndex == -1 || symbolIndex == -1 || priceIndex == -1 || commissionIndex == -1 || feeIndex == -1 {
		return nil, fmt.Errorf("required column not found transactions.csv in header row")
	}

	// Read each rows in the CSV file and create transaction string which
	// contains just the Date, Amount and Description records
	var transactions []*Transaction
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		// Parse the Date field using the date format
		date, err := time.Parse(dateFormat, row[dateIndex])
		if err != nil {
			return nil, fmt.Errorf("failed parsing date: %s", err)
		}

		// Parse the Amount, Price, Commission and Fee fields using Float64 format
		stringAmount := strings.Replace(row[amountIndex], "$", "0", 1)
		stringAmount = strings.Replace(stringAmount, ",", "", 1)
		if stringAmount == "" {
			stringAmount = "0"
		}
		amount, err := strconv.ParseFloat(stringAmount, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing Amount: %s", err)
		}

		stringPrice := strings.Replace(row[priceIndex], "$", "0", 1)
		stringPrice = strings.Replace(stringPrice, ",", "", 1)
		if stringPrice == "" {
			stringPrice = "0"
		}
		price, err := strconv.ParseFloat(stringPrice, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing price: %s", err)
		}
		stringCommission := strings.Replace(row[commissionIndex], "$", "0", 1)
		stringCommission = strings.Replace(stringCommission, ",", "", 1)
		if stringCommission == "" {
			stringCommission = "0"
		}
		commission, err := strconv.ParseFloat(stringCommission, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing commission....: %s", err)
		}

		stringFee := strings.Replace(row[feeIndex], "$", "0", 1)
		stringFee = strings.Replace(stringFee, ",", "", 1)
		if stringFee == "" {
			stringFee = "0"
		}
		fee, err := strconv.ParseFloat(stringFee, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing fee: %s", err)
		}

		// Create a new transaction and add it to the slice
		transaction := &Transaction{
			date:        date,
			description: row[descriptionIndex],
			symbol:      row[symbolIndex],
			price:       price,
			commission:  commission,
			fee:         fee,
			amount:      amount,
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// ******************************* func writeBeancount

// writeBeancount writes the transactions to a file in Beancount format.
func writeBeancount(filename string, transactions []*Transaction, account1 string, currency string, reverse bool) error {
	// Read the matches from the "match.csv" file and create check slice
	matches, err := readMatches("match.csv")
	if err != nil {
		return err
	}

	// Open the output file
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Write the transactions to the output file
	for _, t := range transactions {
		// Reverse the transaction amount if necessary
		amount := t.amount
		if reverse {
			amount = -amount
		}
		amount2 := -amount

		// Find the matching account for the transaction description
		account2 := "Expenses:Unassigned"
		for _, m := range matches {
			if strings.Contains(t.description, m.contains) {
				test := amount
				account2 = m.account2
				if test < 0 {
					test = -test
				}
				break
			}
		}
		// write line 1 of beancount transaction record
		w := tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
		fmt.Fprintf(w, "%s * %q\n", t.date.Format("2006-01-02"), t.description)
		w.Flush()
		// write line 2 of beancount transaction record
		w = tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
		fmt.Fprintf(w, "  %-40s\t%.2f  %s\n", account1, amount, currency)
		w.Flush()
		// write line 3 of beancount transaction record
		w = tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
		fmt.Fprintf(w, "  %-40s\t%.2f  %s\n\n", account2, amount2, currency)
		w.Flush()

	}
	return nil
}

// ******************************* func readMatches
// readMatches reads the matches from a CSV file.
func readMatches(filename string) ([]*Match, error) {
	// Open the input CSV file and read header - first row
	inFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()
	reader := csv.NewReader(inFile)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	// Find the indexes of the Account2, and contains columns
	account2Index := -1
	containsIndex := -1
	
	for i, col := range header {
		switch col {
		case "account2":
			account2Index = i
		case "contains":
			containsIndex = i
		}
	}

	// Make sure we found all the necessary columns
	if account2Index == -1 || containsIndex == -1 {
		return nil, fmt.Errorf("required column not found in match.csv header row")
	}

	// Read each rows in the CSV file and create match string which
	// contains just the fromAccount, toAccount, altAccount, contains and greaterThan records
	var matches []*Match
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		// Create a new transaction and add it to the slice
		match := &Match{
			account2:    row[account2Index],
			contains:    row[containsIndex],
		}
		matches = append(matches, match)
	}
	return matches, nil
}
