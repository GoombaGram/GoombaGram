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

	// Add version as first constant
	_, err = output.Write([]byte("\tTL_layer = " + version + "\n"))
	if err != nil {
		return err
	}

	// Buffer I/O Scanner from file
	scanner := bufio.NewScanner(file)

	// Result variables
	constants := make([][]byte, 0)
	structures := make([][]byte, 0)
	functions := make([][]byte, 0)


	// Parse every line
	for scanner.Scan() {
		// Line checker

		line := scanner.Text()
		// If this line is useless, skip it
		if line == "" || line == "---functions---" || line == "---types---" || (line[0] == '/' && line[1] == '/') {
			continue
		}

		// Constants parser
		// Parse CRC-32 values to Go constants

		// Replace points with an underscore to save names as a variable and split line struct name and CRC-32 value
		// Replace all the "." with an underscore to save the names as variables. Then split into constructor/function name and CRC-32 value.
		lineArr := strings.Split(strings.Replace(line, ".", "_", -1), "#")


		// If CRC length is less than 8, add the padding
		crcLength := len(strings.Split(lineArr[1], " ")[0])
		if crcLength < 8 {
			for i := crcLength; crcLength < 8; i++ {
				lineArr[1] = "0" + lineArr[1]
			}
		}

		// Save parsed constant as byte slice
		constants = append(constants, []byte("\tcrc_" + lineArr[0] + " = 0x"+lineArr[1][:8]+"\n"))


		// Struct parser
		// Parse TL to go struct

		// Vector type will be skipped (it is neither struct nor function, but only a constant)
		if lineArr[0] == "vector" {
			continue
		}

		// Add struct header to structures []byte
		structures = append(structures, []byte("type TL_" + lineArr[0] + " struct {\n"))

		// String that contains needed parameters
		paramsString := strings.Split(lineArr[1][9:], "=")[0]
		// paramsArray contains an array of single param
		paramsArray := strings.Split(paramsString, " ")
		// Remove last element (an empty string)
		paramsArray = paramsArray[:len(paramsArray) - 1]

		// If there are parameters, add them to struct, else skip
		if paramsString != "" {
			for _, param := range paramsArray {
				// Split single param into two parts 0: single param name, 1: single param type
				singleParamArray := strings.Split(param, ":")

				structures = append(structures, []byte(singleParamArray[0]+"\t"))

				switch singleParamArray[1] {
				case "int":
					structures = append(structures, []byte("int32"))
					break
				case "long":
					structures = append(structures, []byte("int64"))
					break
				case "string":
					structures = append(structures, []byte("string"))
					break
				case "double":
					structures = append(structures, []byte("float64"))
					break
				case "bytes":
					structures = append(structures, []byte("[]byte"))
					break
				case "Vector<int>":
					structures = append(structures, []byte("[]int32"))
					break
				case "Vector<long>":
					structures = append(structures, []byte("[]int64"))
					break
				case "Vector<string>":
					structures = append(structures, []byte("[]string"))
					break
				case "Vector<double>":
					structures = append(structures, []byte("[]float64"))
					break
				case "!X":
					structures = append(structures, []byte("TL"))
					break
				default:
					if strings.Contains(singleParamArray[1], "Vector<") {
						structures = append(structures, []byte("[]TL // " + singleParamArray[1][strings.LastIndex(singleParamArray[1], "Vector<")+len("Vector<"):]))
					} else {
						structures = append(structures, []byte("TL // " + singleParamArray[1]))
					}
				}
			}
		}

		// Close parsed struct
		structures = append(structures, []byte("}\n\n"))

		fmt.Println(paramsArray)

		// Functions parser
		// encode functions
		functions = append(functions, []byte("func (e TL_" + lineArr[0] + ") Encode() []byte {\n"))
		functions = append(functions, []byte("x := NewEncodeBuf(512)\n"))
		functions = append(functions, []byte("x.UInt(crc_" + lineArr[0] + ")\n"))

		// If there are parameters, add them to encode function, else skip
		if paramsString != "" {
			for _, param := range paramsArray {
				// Split single param into two parts 0: single param name, 1: single param type
				singleParamArray := strings.Split(param, ":")

				fmt.Println(paramsArray)

				switch singleParamArray[1] {
				case "int":
					functions = append(functions, []byte("x.Int(e."+singleParamArray[0]+")\n"))
					break
				case "long":
					functions = append(functions, []byte("x.Long(e."+singleParamArray[0]+")\n"))
					break
				case "string":
					functions = append(functions, []byte("x.String(e."+singleParamArray[0]+")\n"))
					break
				case "double":
					functions = append(functions, []byte("x.Double(e."+singleParamArray[0]+")\n"))
					break
				case "bytes":
					functions = append(functions, []byte("x.StringBytes(e."+singleParamArray[0]+")\n"))
					break
				case "Vector<int>":
					functions = append(functions, []byte("x.VectorInt(e."+singleParamArray[0]+")\n"))
					break
				case "Vector<long>":
					functions = append(functions, []byte("x.VectorLong(e."+singleParamArray[0]+")\n"))
					break
				case "Vector<string>":
					functions = append(functions, []byte("x.VectorString(e."+singleParamArray[0]+")\n"))
					break
				case "!X":
					functions = append(functions, []byte("x.Bytes(e."+singleParamArray[0]+")\n"))
					break
				case "Vector<double>":
					panic("Unsupported " + singleParamArray[1])
				default:
					if strings.Contains(singleParamArray[1], "Vector<") {
						functions = append(functions, []byte("x.Vector(e."+singleParamArray[0]+")\n"))
					} else {
						functions = append(functions, []byte("x.Bytes(e."+singleParamArray[0]+".encode())\n"))
					}
				}
			}
		}

		functions = append(functions, []byte("return x.buf\n}\n\n"))

		functions = append(functions, []byte("func (m *DecodeBuf) ObjectGenerated(constructor uint32) (r TL) {\n\t switch constructor {"))

		for _, key := range _order {
			c := _cons[key]
			fmt.Printf("case crc_%s:\n", c.predicate)
			fmt.Printf("r = TL_%s{\n", c.predicate)
			for _, t := range c.params {
				switch t._type {
				case "int":
					fmt.Print("m.Int(),\n")
				case "long":
					fmt.Print("m.Long(),\n")
				case "string":
					fmt.Print("m.String(),\n")
				case "double":
					fmt.Print("m.Double(),\n")
				case "bytes":
					fmt.Print("m.StringBytes(),\n")
				case "Vector<int>":
					fmt.Print("m.VectorInt(),\n")
				case "Vector<long>":
					fmt.Print("m.VectorLong(),\n")
				case "Vector<string>":
					fmt.Print("m.VectorString(),\n")
				case "!X":
					fmt.Print("m.Object(),\n")
				case "Vector<double>":
					panic(fmt.Sprintf("Unsupported %s", t._type))
				default:
					var inner string
					n, _ := fmt.Sscanf(t._type, "Vector<%s", &inner)
					if n == 1 {
						fmt.Print("m.Vector(),\n")
					} else {
						fmt.Print("m.Object(),\n")
					}
				}
			}
			fmt.Print("}\n\n")
		}

		fmt.Println(`
	default:
		m.err = fmt.Errorf("Unknown constructor: \u002508x", constructor)
		return nil
	}
	if m.err != nil {
		return nil
	}`)
	} // End of for cycle

	/*
	// Decode function
	functions = append(functions, []byte("func (m *DecodeBuf) ObjectGenerated(constructor uint32) (r TL) {\n\t switch constructor {"))

	for _, key := range _order {
		c := _cons[key]
		fmt.Printf("case crc_%s:\n", c.predicate)
		fmt.Printf("r = TL_%s{\n", c.predicate)
		for _, t := range c.params {
			switch t._type {
			case "int":
				fmt.Print("m.Int(),\n")
			case "long":
				fmt.Print("m.Long(),\n")
			case "string":
				fmt.Print("m.String(),\n")
			case "double":
				fmt.Print("m.Double(),\n")
			case "bytes":
				fmt.Print("m.StringBytes(),\n")
			case "Vector<int>":
				fmt.Print("m.VectorInt(),\n")
			case "Vector<long>":
				fmt.Print("m.VectorLong(),\n")
			case "Vector<string>":
				fmt.Print("m.VectorString(),\n")
			case "!X":
				fmt.Print("m.Object(),\n")
			case "Vector<double>":
				panic(fmt.Sprintf("Unsupported %s", t._type))
			default:
				var inner string
				n, _ := fmt.Sscanf(t._type, "Vector<%s", &inner)
				if n == 1 {
					fmt.Print("m.Vector(),\n")
				} else {
					fmt.Print("m.Object(),\n")
				}
			}
		}
		fmt.Print("}\n\n")
	}

	fmt.Println(`
	default:
		m.err = fmt.Errorf("Unknown constructor: \u002508x", constructor)
		return nil
	}
	if m.err != nil {
		return nil
	}`)
	 */

// Close constants declaration
	constants = append(constants, []byte(")\n\n\n"))

	// Save parsed constants to file
	for _, constantToFile := range constants {
		_, err = output.Write(constantToFile)

		if err != nil {
			return nil
		}
	}

	// Save parsed structures to file
	for _, structToFile := range structures {
		_, err = output.Write(structToFile)

		if err != nil {
			return nil
		}
	}

	// Save parsed functions to file
	for _, funcToFile := range functions {
		_, err = output.Write(funcToFile)

		if err != nil {
			return nil
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