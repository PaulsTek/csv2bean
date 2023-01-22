// cdv2fid - specific to fidelity
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
	dateFormat 		string
	defaultAccount  string
	currency   		string
	reverse			bool	
}

// Transaction represents a financial transaction from the input CSV file.
type Transaction struct {
	date        time.Time
	description string
	symbol      string
	quantity    float64
	price       float64
	commission  float64
	fee         float64
	amount      float64
}

// Match represents a description-to-account match from the "match.csv" file.
type Match struct {
	fromAccount string
	toAccount   string
	trade       string
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
	transactions, err := readTransactions("transactions-inv.csv", config.dateFormat)
	if err != nil {
		log.Fatal(err)
	}
	// Write the transaction string to the output file in Beancount format
	if err := writeBeancount("output-inv.bean", transactions, config.currency, config.defaultAccount, config.reverse); err != nil {
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
		dateFormat: 		record[0],
		defaultAccount:   	record[1],
		currency:   		record[2],
		reverse:    		reverse,
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
		return nil, fmt.Errorf("failed to open transaction file: %s", err)
	}
	defer inFile.Close()
	reader := csv.NewReader(inFile)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction record: %s", err)
	}

	// Find the indexes of the Date, Description and Amount columns
	dateIndex := -1
	descriptionIndex := -1
	symbolIndex := -1
	quantityIndex := -1
	priceIndex := -1
	commissionIndex := -1
	feeIndex := -1
	amountIndex := -1

	for i, col := range header {
		switch col {
		case "date":
			dateIndex = i
		case "description":
			descriptionIndex = i
		case "symbol":
			symbolIndex = i
		case "quantity":
			quantityIndex = i
		case "price":
			priceIndex = i
		case "commission":
			commissionIndex = i
		case "fee":
			feeIndex = i
		case "amount":
			amountIndex = i
		}
	}

	// Make sure we found all the necessary columns
	if dateIndex == -1 || descriptionIndex == -1 || amountIndex == -1 || symbolIndex == -1 || priceIndex == -1 || commissionIndex == -1 || feeIndex == -1 {
		return nil, fmt.Errorf("required column not found in transaction header row")
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

		// Parse the Amount field using Float64 format
		stringAmount := strings.Replace(row[amountIndex], "$", "0", 1)
		stringAmount = strings.Replace(stringAmount, ",", "", 1)
		if stringAmount == "" {
			stringAmount = "0"
		}
		amount, err := strconv.ParseFloat(stringAmount, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing Amount: %s", err)
		}
		// Parse the Quantity field using Float64 format
		stringQuantity := strings.Replace(row[quantityIndex], "$", "0", 1)
		stringQuantity = strings.Replace(stringQuantity, ",", "", 1)
		if stringQuantity == "" {
			stringQuantity = "0"
		}
		quantity, err := strconv.ParseFloat(stringQuantity, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing price: %s", err)
		}

		// Parse the Price field using Float64 format
		stringPrice := strings.Replace(row[priceIndex], "$", "0", 1)
		stringPrice = strings.Replace(stringPrice, ",", "", 1)
		if stringPrice == "" {
			stringPrice = "0"
		}
		price, err := strconv.ParseFloat(stringPrice, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing price: %s", err)
		}

		// Parse the Commission field using Float64 format
		stringCommission := strings.Replace(row[commissionIndex], "$", "0", 1)
		stringCommission = strings.Replace(stringCommission, ",", "", 1)
		if stringCommission == "" {
			stringCommission = "0"
		}
		commission, err := strconv.ParseFloat(stringCommission, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing commission....: %s", err)
		}

		// Parse the Fee field using Float64 format
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
			quantity:    quantity,
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
func writeBeancount(filename string, transactions []*Transaction, currency string, defaultAccount string, reverse bool) error {
	// Read the matches from the "match.csv" file and create check slice
	matches, err := readMatches("match-inv.csv")
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
		// Reverse the transaction amount if necessary and calc price
		fromAmount := t.amount
		toAmount := -t.amount
		if reverse {
			fromAmount = -fromAmount
			toAmount = -toAmount
		}

		// Find the matching account for the transaction description
		fromAccount := defaultAccount
		toAccount := "Expenses:Unassigned"
		trade := ""
		for _, m := range matches {
			if strings.Contains(t.description, m.contains) {
				fromAccount = m.fromAccount
				toAccount = m.toAccount
				trade = m.trade
				break
			}
		}
		// write line 1 of beancount transaction record
		w := tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
		fmt.Fprintf(w, "%s * %q\n", t.date.Format("2006-01-02"), t.description)
		w.Flush()
		if trade == "BUY" {
			fromAmount = t.price * t.quantity
			// write line 2 of beancount transaction record
			w = tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
			fmt.Fprintf(w, "  %-40s\t%.2f  %s\n", fromAccount, -fromAmount, currency)
			w.Flush()
			// write line 3 of beancount transaction record
			w = tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
			fmt.Fprintf(w, "  %s%-15s\t%.3f  %s  {%.4f  %s}\n\n", toAccount, t.symbol, t.quantity, t.symbol, t.price, currency)
			w.Flush()
		} else if trade == "SELL" {
			toAmount = t.price * t.quantity
			// write line 2 of beancount transaction record
			w = tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
			fmt.Fprintf(w, "  %s%-15s\t%.3f  %s  {%.4f  %s}\n", fromAccount, t.symbol, t.quantity, t.symbol, t.price, currency)
			w.Flush()
			// write line 3 of beancount transaction record
			w = tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
			fmt.Fprintf(w, "  %-40s\t%.2f  %s\n\n", toAccount, -toAmount, currency)
			w.Flush()
		} else {
			// write line 2 of beancount transaction record
			w = tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
			fmt.Fprintf(w, "  %-40s\t%.2f  %s\n", fromAccount, fromAmount, currency)
			w.Flush()
			// write line 3 of beancount transaction record
			w = tabwriter.NewWriter(outFile, 0, 0, 1, ' ', 0)
			fmt.Fprintf(w, "  %-40s\t%.2f  %s\n\n", toAccount, toAmount, currency)
			w.Flush()
		}

	}
	return nil
}

// ******************************* func readMatches
// readMatches reads the matches from a CSV file.
func readMatches(filename string) ([]*Match, error) {
	// Open the input CSV file and reader (note ... no header row needed)
	inFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open match file: %s", err)
	}
	defer inFile.Close()
	reader := csv.NewReader(inFile)

	// Read the rows in the CSV file and check for match
	var matches []*Match
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		// Create a new match and add it to the slice
		match := &Match{
			fromAccount: row[0],
			toAccount:   row[1],
			trade:       row[2],
			contains:    row[3],
		}
		matches = append(matches, match)
	}

	return matches, nil
}
