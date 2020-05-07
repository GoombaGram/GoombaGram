package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TL sources:
// https://core.telegram.org/schema
// https://github.com/telegramdesktop/tdesktop/tree/dev/Telegram/Resources/tl

func parser (file *os.File, version string) error {
	// Remove old schema
	_ = os.Remove("result/schema.go")

	// Create a new output file
	output, err := os.OpenFile("result/schema.go", os.O_APPEND|os.O_RDWR|os.O_CREATE, 666)
	if err != nil {
		return err
	}
	defer output.Close()

	// Header
	_, err = output.Write([]byte("package TL\n\nimport \"errors\"\n\nconst(\n"))
	if err != nil {
		return err
	}

	// Append tl schema version to constants
	_, err = output.Write([]byte(fmt.Sprintf("	%s = %s\n", "TL_layer", version)))
	if err != nil {
		return err
	}

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// If it's an useless line, skip it
		if line == "" || line == "---functions---" || line == "---types---" || (line[0] == '/' && line[1] == '/') {
			continue
		}

		// Replace points with an underscore to save names as a variable
		line = strings.Replace(line, ".", "_", -1)
		// Split filed name and CRC-32 value
		lineArr := strings.Split(line, "#")

		// Write name and crc value as constants
		_, err = output.Write([]byte(fmt.Sprintf("	CRC_%s = 0x%s\n", lineArr[0], lineArr[1][:8])))
		if err != nil {
			return err
		}
	}

	_, err = output.Write([]byte(")\n\n"))
	if err != nil {
		return err
	}

	return nil
	// Variable that indicate if we are parsing struct or functions in every moment
	functionParse := false

	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// If it's a comment or an empty line, skip
		if (line[0] == '/' && line[1] == '/') || line == "" {
			continue
		}

		if line == "---functions---" {
			functionParse = true
			continue
		} else if line == "---types---" {
			functionParse = false
			continue
		}

		if functionParse == false {

		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	// Get file name as input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Insert .tl schema file name: ")
	scanner.Scan()
	fileName := scanner.Text()

	fmt.Print("Insert TL layer version: ")
	scanner.Scan()
	tlVersion := scanner.Text()

	// Open file and close at end
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Parse TL to Go struct and functions
	err = parser(file, tlVersion)
	if err != nil {
		panic(err)
	}

}