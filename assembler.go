package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	register  = iota // регистровая
	addr             // прямая
	indirect         // косвенная регистровая
	immediate        // непосредственная
)

type AddresationType uint

type CommandType struct {
	Name string
	AddresationType
}

var commands = map[CommandType]string{
	{"ld", register}:    "0000",
	{"ld", indirect}:    "0001",
	{"st", indirect}:    "0010",
	{"add", register}:   "0011",
	{"shra", register}:  "0100",
	{"nand", register}:  "0101",
	{"shl", register}:   "0110",
	{"ld", addr}:        "0111",
	{"ld", immediate}:   "1000",
	{"st", addr}:        "1001",
	{"add", immediate}:  "1010",
	{"shra", immediate}: "1011",
	{"nand", immediate}: "1100",
	{"shl", immediate}:  "1101",
	{"jmp", addr}:       "1110",
	{"jz", addr}:        "1111",
}

func main() {
	// Открываем файл для чтения
	file, err := os.Open("program.asm")
	if err != nil {
		log.Fatalf("Не удалось открыть файл: %v", err)
	}
	defer file.Close()

	// Создаем сканнер для чтения строк
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Ошибка при чтении файла: %v", err)
	}

	binaryLines := convert(lines)
	save(binaryLines)
}

func convert(lines []string) []string {
	binaryLines := make([]string, 0)
	for i, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(strings.Replace(line, " ", "~", 1), "~")

		if strings.ToLower(parts[0]) == "ld" {
			if strings.Contains(parts[1], ",") {
				ops := strings.Split(parts[1], ",")
				commandType := CommandType{Name: strings.ToLower(parts[0])}
				if strings.Contains(ops[1], "#") {
					commandType.AddresationType = addr
				} else if strings.Contains(ops[1], "[R") {
					commandType.AddresationType = indirect
				} else if strings.Contains(ops[1], "R") {
					commandType.AddresationType = register
				} else {
					commandType.AddresationType = immediate
				}

				kop, ok := commands[commandType]
				if !ok {
					log.Fatalf("Не найден код операции в строке %v", i)
				}

				fmt.Println(ops[0], ronCode(ops[0]))
				binaryLines = append(binaryLines, kop+ronCode(ops[0])+ronCode(ops[1]))
				if commandType.AddresationType == addr || commandType.AddresationType == immediate {
					val, err := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(ops[1], "#", "")))
					if err != nil {
						log.Fatalf("Ошибка операции %v: %v", i, err)
					}

					binaryLines = append(binaryLines, fmt.Sprintf("%08b", uint8(val)))
				}
			} else {
				log.Fatalf("Не указаны операторы LD в строке %v", i)
			}
		} else if strings.ToLower(parts[0]) == "st" {
			if strings.Contains(parts[1], ",") {
				ops := strings.Split(parts[1], ",")

				commandType := CommandType{strings.ToLower(parts[0]), addr}
				if strings.Contains(ops[1], "[R") {
					commandType.AddresationType = indirect
				}

				kop, ok := commands[commandType]
				if !ok {
					log.Fatalf("Не найден код операции в строке %v", i)
				}

				binaryLines = append(binaryLines, kop+ronCode(ops[0])+ronCode(ops[1]))

				if commandType.AddresationType == addr {
					addr, err := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(ops[1], "#", "")))
					if err != nil {
						log.Fatalf("Ошибка операции %v: %v", i, err)
					}

					binaryLines = append(binaryLines, fmt.Sprintf("%08b", uint8(addr)))
				}
			} else {
				log.Fatalf("Не указаны операторы ST в строке %v", i)
			}
		} else if strings.ToLower(parts[0]) == "add" || strings.ToLower(parts[0]) == "shra" || strings.ToLower(parts[0]) == "nand" || strings.ToLower(parts[0]) == "shl" {
			if strings.Contains(parts[1], ",") {
				ops := strings.Split(parts[1], ",")
				commandType := CommandType{strings.ToLower(parts[0]), register}
				if !strings.Contains(ops[1], "R") {
					commandType.AddresationType = immediate
				}

				kop, ok := commands[commandType]
				if !ok {
					log.Fatalf("Не найден код операции в строке %v", i)
				}

				binaryLines = append(binaryLines, kop+ronCode(ops[0])+ronCode(ops[1]))

				if commandType.AddresationType == immediate {
					imm8, err := strconv.Atoi(strings.TrimSpace(ops[1]))
					if err != nil {
						log.Fatalf("Ошибка операции %v: %v", i, err)
					}

					binaryLines = append(binaryLines, fmt.Sprintf("%08b", uint8(imm8)))
				}
			} else {
				log.Fatalf("Не указаны операторы для АЛУ в строке %v", i)
			}
		} else if strings.ToLower(parts[0]) == "jmp" || strings.ToLower(parts[0]) == "jz" {
			if strings.Contains(parts[1], "#") {
				commandType := CommandType{strings.ToLower(parts[0]), addr}
				kop, ok := commands[commandType]
				if !ok {
					log.Fatalf("Не найден код операции в строке %v", i)
				}

				addr, err := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(parts[1], "#", "")))
				if err != nil {
					log.Fatalf("Ошибка прыжка %v: %v", i, err)
				}

				binaryLines = append(binaryLines, kop+"0000")
				binaryLines = append(binaryLines, fmt.Sprintf("%08b", uint8(addr)))
			} else {
				log.Fatalf("Не указан адрес прыжка в строке %v", i)
			}
		} else {
			log.Fatalf("Ошибка в строке %v", i)
		}
	}

	return binaryLines
}

func ronCode(reg string) string {
	str := strings.TrimSpace(reg)
	str = strings.ReplaceAll(str, "[", "")
	str = strings.ReplaceAll(str, "]", "")

	switch str {
	case "R1":
		return "01"
	case "R2":
		return "10"
	case "R3":
		return "11"
	default:
		return "00"
	}
}

const header = `
WIDTH=8;
DEPTH=256;

ADDRESS_RADIX=UNS;
DATA_RADIX=BIN;

CONTENT BEGIN`

func save(binaryLines []string) {
	var content string
	var last int
	for i, line := range binaryLines {
		content += fmt.Sprintf("	%v  :   %v;\n", i, line)
		last = i
	}

	last++
	content += fmt.Sprintf("	[%v..255]  :   00000000;\n", last)

	data := header + "\n" + content + "END;"
	err := os.WriteFile("ram.mif", []byte(data), 0644)
	if err != nil {
		log.Fatalf("Ошибка записи файла: %v", err)
	}
}
