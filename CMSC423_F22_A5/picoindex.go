package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

// type tallyNode struct {
// 	// A int
// 	// C int
// 	// G int
// 	// T int
// 	Node map[byte]int
// }

type outData struct {
	SuffArr    []int
	BWT        []byte
	FArr       []byte
	TallyBWT   [][]int
	FullGenome string
}

func main() {

	// timeCheck("Program begins")

	args := os.Args
	reference := args[1]
	output := args[2]

	//gets raw text file with $ appended
	text := ReadInFA(reference) + "$"

	//creates suffix array from
	suffArr := CreateSuffixArray(reference)

	// timeCheck("Suffix Array Created")

	//f-index and L-index which is really just BWT
	// var fArr []byte
	// var BWT []byte
	fArr := make([]byte, len(suffArr))
	BWT := make([]byte, len(suffArr))

	tallyBWT := make([][]int, len(text))

	//fill BWT array by referencing the suffix array and going to the previous character
	//additionaly builds tally array along the way
	// var currTallyNode tallyNode
	for i := 0; i < len(suffArr); i++ {
		var char byte
		if suffArr[i] != 0 {
			char = text[suffArr[i]-1]
			BWT[i] = char
		} else {
			char = '$'
			BWT[i] = char
		}

		// currTallyNode = addElem(tallyBWT, char, alphabet)
		// tallyBWT = append(tallyBWT, currTallyNode)

		switch char {
		case '$':
			if i == 0 {
				tallyBWT[i] = []int{1, 0, 0, 0, 0}
			} else {
				tallyBWT[i] = make([]int, len(tallyBWT[i-1]))
				copy(tallyBWT[i], tallyBWT[i-1])
				// fmt.Printf("%d", tallyBWT[i][0])
				tallyBWT[i][0] = tallyBWT[i][0] + 1
			}
		case 'A':
			if i == 0 {
				tallyBWT[i] = []int{0, 1, 0, 0, 0}
			} else {
				tallyBWT[i] = make([]int, len(tallyBWT[i-1]))
				copy(tallyBWT[i], tallyBWT[i-1])

				// fmt.Printf("%d", tallyBWT[i][0])
				tallyBWT[i][1] = tallyBWT[i][1] + 1
			}
		case 'C':
			if i == 0 {
				tallyBWT[i] = []int{0, 0, 1, 0, 0}
			} else {
				tallyBWT[i] = make([]int, len(tallyBWT[i-1]))
				copy(tallyBWT[i], tallyBWT[i-1])
				// fmt.Printf("%d", tallyBWT[i][0])
				tallyBWT[i][2] = tallyBWT[i][2] + 1
			}
		case 'G':
			if i == 0 {
				tallyBWT[i] = []int{0, 0, 0, 1, 0}
			} else {
				tallyBWT[i] = make([]int, len(tallyBWT[i-1]))
				copy(tallyBWT[i], tallyBWT[i-1])
				// fmt.Printf("%d", tallyBWT[i][0])
				tallyBWT[i][3] = tallyBWT[i][3] + 1
			}
		case 'T':
			if i == 0 {
				tallyBWT[i] = []int{0, 0, 0, 0, 1}
			} else {
				tallyBWT[i] = make([]int, len(tallyBWT[i-1]))
				copy(tallyBWT[i], tallyBWT[i-1])
				// fmt.Printf("%d", tallyBWT[i][0])
				tallyBWT[i][4] = tallyBWT[i][4] + 1
			}
		}
	}

	// timeCheck("BWT and BWT tally generated")

	//fills F array by referencing the first cahracter of each suff arr indice
	//also builds tally along the way
	for i := 0; i < len(suffArr); i++ {
		char := text[suffArr[i]]
		fArr[i] = char
	}

	// timeCheck("F and F tally generated")

	createOutput2(suffArr, BWT, fArr, tallyBWT, text, output)
	// timeCheck("buildfm complete")
	// testOutput(suffArr, BWT, fArr, tallyBWT, output)

}

func testOutput(suffixArr []int, BWT []byte, fArr []byte, tallyBWT [][]int, output string) {
	fTest, err := os.Create("testBuildFm")
	Check(err)
	defer fTest.Close()

	push := bufio.NewWriter(fTest)

	//Suffix Array
	fmt.Fprint(push, fmt.Sprintf("Suffix Array\n"))
	for i := 0; i < len(suffixArr); i++ {
		fmt.Fprint(push, fmt.Sprintf(strconv.Itoa(suffixArr[i])+"\n"))
	}

	//F
	fmt.Fprint(push, fmt.Sprintf("\nF Array \n"))
	for i := 0; i < len(fArr); i++ {
		fmt.Fprint(push, fmt.Sprintf(string(fArr[i])+"\n"))
	}

	//BWT
	fmt.Fprint(push, fmt.Sprintf("\n BWT Array \n"))
	for i := 0; i < len(BWT); i++ {
		fmt.Fprint(push, fmt.Sprintf(string(BWT[i])+"\n"))
	}

	//tally BWT
	fmt.Fprint(push, "$ A C G T")
	fmt.Fprint(push, fmt.Sprintf("\n"))
	for i := 0; i < len(tallyBWT); i++ {
		fmt.Fprintf(push, "%d %d %d %d %d", tallyBWT[i][0], tallyBWT[i][1], tallyBWT[i][2], tallyBWT[i][3], tallyBWT[i][4])
		fmt.Fprintf(push, fmt.Sprintf("\n"))

	}
	push.Flush()

}

/*
PURPOSE:
Creates designated encoder output, I really hope this GOB shit works
*/
func createOutput2(suffixArr []int, BWT []byte, fArr []byte, tallyBWT [][]int, genome string, output string) {
	data := outData{SuffArr: suffixArr, BWT: BWT, FArr: fArr, TallyBWT: tallyBWT, FullGenome: genome}

	f, err := os.Create(output)
	Check(err)
	defer f.Close()

	encoder := gob.NewEncoder(f)
	err = encoder.Encode(data)
	Check(err)
}

// cant be bothered to learn how to code

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func ReadInFA(ref string) string {
	reference := ref

	f, err := os.Open(reference)
	Check(err)
	defer f.Close()

	//processes header line
	s := bufio.NewScanner(f)
	s.Scan()

	//genome contains entire genome
	genome := ""
	for s.Scan() {
		genome = genome + s.Text()
	}
	// genome = genome + "$"

	return genome
}

func readInFA(ref string) string {
	reference := ref

	f, err := os.Open(reference)
	Check(err)
	defer f.Close()

	//processes header line
	s := bufio.NewScanner(f)
	s.Scan()

	//genome contains entire genome
	genome := ""
	for s.Scan() {
		genome = genome + s.Text()
	}
	// genome = genome + "$"

	return genome
}

func numerizeString(text string) []int {
	numerize := make(map[string]int)
	numerize["A"] = 2
	numerize["C"] = 3
	numerize["G"] = 4
	numerize["T"] = 5
	numerize["$"] = 1

	numerize["a"] = 2
	numerize["b"] = 3
	numerize["n"] = 4
	// numerize["$"] = 5

	// numerize["A"] = 1
	// numerize["G"] = 2
	// numerize["T"] = 3

	// numerize["t"] = 4
	// numerize["r"] = 3
	// numerize["o"] = 2
	// numerize["l"] = 1

	retString := []int{}
	for i := 0; i < len(text); i++ {
		number := numerize[string(text[i])]
		retString = append(retString, number)
	}

	return retString
}

type Suffix2 struct {
	position int
	content1 int
	content2 int
	rank     int
}

type SuffixArr2 []Suffix2

func (s SuffixArr2) Len() int {
	return len(s)
}
func (s SuffixArr2) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// if content of i is less than content of j return true
func (s SuffixArr2) Less(i, j int) bool {
	// content1 := s[i].content
	// content2 := s[j].content

	// for (int(math.Log10(float64(content1))) != int(math.Log10(float64(content2)))) && (content1 != 0 && content2 != 0) {
	// 	if content1 < content2 {
	// 		content1 = content1 * 10
	// 	} else {
	// 		content2 = content2 * 10
	// 	}
	// }

	if s[i].content1 < s[j].content1 {
		return true
	} else if s[i].content1 > s[j].content1 {
		return false
	} else { //switch to checking second parameter
		if s[i].content2 < s[j].content2 {
			return true
		} else {
			return false
		}
	}

}

func rankify2(text []int, iteration int) []int {

	//populates suffix array
	suffixArray := []Suffix2{}
	for i := 0; i < len(text); i++ {
		newSuffix := Suffix2{}
		newSuffix.position = i
		if i+int(math.Pow(2, float64(iteration))) < len(text) { //space to jump for second val
			// mergeStrings := strconv.Itoa(text[i]) + strconv.Itoa(text[i+int(math.Pow(2, float64(iteration)))])
			// newSuffix.content, _ = strconv.Atoi(mergeStrings)

			newSuffix.content1 = text[i]
			newSuffix.content2 = text[i+int(math.Pow(2, float64(iteration)))]
		} else {
			newSuffix.content1 = text[i]
			newSuffix.content2 = -1
		}
		newSuffix.rank = 0
		suffixArray = append(suffixArray, newSuffix)
	}

	// printSuffArr(suffixArray)

	//Theoretically should sort the array in ascending order
	// printSuffArr(suffixArray)

	sort.Sort(SuffixArr2(suffixArray))
	// printSuffArr(suffixArray)

	//set rank
	rank := -1
	for i := 0; i < len(suffixArray); i++ {
		if i != 0 && suffixArray[i].content1 == suffixArray[i-1].content1 && suffixArray[i].content2 == suffixArray[i-1].content2 {
			suffixArray[i].rank = rank
		} else {
			rank++
			suffixArray[i].rank = rank

		}
	}

	//return array
	rankArray := make([]int, len(text))
	for i := 0; i < len(text); i++ {
		rankArray[suffixArray[i].position] = suffixArray[i].rank
	}

	// printPosition(rankArray)

	return rankArray

}

func CreateSuffixArray(ref string) []int {

	genome := readInFA(ref) + "$"

	text := numerizeString(genome)

	// printPosition(text)

	var newText []int
	newText = text
	for i := 0; i < int(math.Ceil(math.Log(float64(len(text))))); i += 1 {
		newText = rankify2(newText, i)
		// printPosition(newText)
	}

	//sets up final suffix array
	finalSuffixArr := make([]int, len(newText))
	for i := 0; i < len(newText); i++ {
		finalSuffixArr[newText[i]] = i
	}
	// printPosition(finalSuffixArr)

	// viewSuffixArray(finalSuffixArr)
	//ngl Im not sure where the 9 came from but hey i guess it just be like that sometimes
	// return finalSuffixArr[1:]
	return finalSuffixArr

}

/*
RETURN:
names, patterns
*/
func QueryParser(queryFile string) ([]string, []string) {
	names := []string{}
	queries := []string{}
	f, err := os.Open(queryFile)
	check(err)

	s := bufio.NewScanner(f)

	var query string
	query = ""
	for s.Scan() {
		line := s.Text()
		if line[0] == '>' {
			if query != "" {
				queries = append(queries, query)
				query = ""
			}
			names = append(names, line[1:])
		} else {
			query = query + line
		}
	}
	if query != "" {
		queries = append(queries, query)
	}

	return names, queries
}

func timeCheck(msg string) {
	dt := time.Now()
	fmt.Println(msg)
	fmt.Println(dt.Format("01-02-2006 15:04:05"))
}
