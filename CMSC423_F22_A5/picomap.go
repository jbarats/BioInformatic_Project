package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"strconv"
)

type outData struct {
	SuffArr    []int
	BWT        []byte
	FArr       []byte
	TallyBWT   [][]int
	FullGenome string
}

type outputRecord struct {
	read_name string
	//num_aligments int (will always be setting this to one)
	ref_start int
	score     int
	CIGAR     string
}

func main() {
	args := os.Args
	index_file := args[1]
	read_file := args[2]
	mismatch_penalty, err := strconv.Atoi(args[3])
	check(err)
	gap_penalty, err := strconv.Atoi(args[4])
	check(err)
	output_file := args[5]

	//unpackage binary dile
	data := obtainData(index_file)

	//store genome text
	genome := data.FullGenome

	//process all reads into string lists
	names, read_list := processReads(read_file)

	//will contain start and end points for CIGAR comparison to original genome
	var start int
	var end int

	//will store all output records
	outRecords := []outputRecord{}

	//iterate through all reads and run heuristics
	for i := 0; i < len(read_list); i++ {
		outRecord := outputRecord{}

		//Sets first start to unreachable value, empty list of potential starts, lenght of seeds
		start = -1
		temp_starts := []int{}
		split_len := int(len(read_list[i]) / 4)

		//read currently being worked with
		curr_read := read_list[i]
		curr_name := names[i]

		//looks for complete match on full read, if full match exists process immediatly otherwise use seed heuristics
		fullHits, _ := c_query(read_list[i], data.FArr, data.BWT, data.TallyBWT, data.SuffArr)
		if len(fullHits) > 0 {
			start = fullHits[0]
			outRecord.read_name = curr_name
			outRecord.ref_start = fullHits[0]
			outRecord.score = 0
			outRecord.CIGAR = strconv.Itoa(len(curr_read)) + "="
			outRecords = append(outRecords, outRecord)
		} else {
			//not full match use heuristics

			//get starting positions of small segments
			for j := 0; j < 4; j++ {
				start_pos := int(split_len * j)
				end_pos := start_pos + (split_len - 1)
				curr_segment := curr_read[start_pos:end_pos]
				hits, _ := c_query(curr_segment, data.FArr, data.BWT, data.TallyBWT, data.SuffArr)
				//changes location of hits to where full string would have began
				for k := 0; k < len(hits); k++ {
					hits[k] = hits[k] - split_len*j
				}
				temp_starts = append(temp_starts, hits...)
			}

			//of all potential starting segments chooses ideal, once a start is selected should abort this process
			for i := 0; i < len(temp_starts); i++ {
				if start == -1 {
					for j := 0; j < len(temp_starts); j++ {
						if (i != j) && math.Abs(float64(temp_starts[i]-temp_starts[j])) <= 15 {
							start = temp_starts[i]
							break
						}
					}
				} else {
					break
				}
			}

			//obtain CIGAR score
			// end = start + len(curr_read) + 15
			end = start + len(curr_read)
			if start-15 > 0 {
				start = start - 15
			} else {
				start = 0
			}
			// fmt.Printf("\n\n\n Hey! You! yeah you! Watch out the start is %d \n\n\n", start)
			// fmt.Printf("\n\n\n Unfortunatly the max length of the genome is %d \n\n\n", len(genome))
			yString := genome[start:end]
			cigarInfo := info{name: curr_name, x: read_list[i], y: yString}
			cigData := fitting(mismatch_penalty, gap_penalty, cigarInfo)

			outRecord.read_name = curr_name
			outRecord.ref_start = start + 15
			outRecord.score = cigData.score
			outRecord.CIGAR = cigData.cigar
			outRecords = append(outRecords, outRecord)

		}

	}

	publishOutput(outRecords, output_file)

}

//-------Functions----------

func publishOutput(outRecords []outputRecord, output_file string) {
	f, err := os.Create(output_file)
	check(err)
	defer f.Close()

	push := bufio.NewWriter(f)

	for i := 0; i < len(outRecords); i++ {
		record := outRecords[i]
		fmt.Fprintf(push, fmt.Sprintf("%s\t1\n", record.read_name))
		fmt.Fprintf(push, fmt.Sprintf("%d\t%d\t%s\n", record.ref_start, record.score, record.CIGAR))
	}

	push.Flush()
}

// return: names, patterns
func processReads(read_file string) ([]string, []string) {
	names := []string{}
	read_list := []string{}

	f, err := os.Open(read_file)
	check(err)

	s := bufio.NewScanner(f)

	for s.Scan() {
		line := s.Text()
		if line[0] == '>' {
			names = append(names, line[1:])
		} else {
			read_list = append(read_list, line)
		}
	}

	return names, read_list
}

func obtainData(index string) outData {
	f, err := os.Open(index)
	Check(err)
	defer f.Close()

	decoder := gob.NewDecoder(f)

	inData := outData{}

	decoder.Decode(&inData)

	return inData
}

func c_query(patternBig string, F []byte, L []byte, lTally [][]int, suffArr []int) ([]int, int) {
	lastNode := lTally[len(lTally)-1]
	totalChar := []int{1, lastNode[1], lastNode[2], lastNode[3], lastNode[4]}

	//index of starting position of each character on F
	//last position is the next one minues 1
	startPos := make([]int, 5)
	startPos[0] = 0
	startPos[1] = 1
	startPos[2] = startPos[1] + totalChar[1]
	startPos[3] = startPos[2] + totalChar[2]
	startPos[4] = startPos[3] + totalChar[3]

	lChar := patternBig[len(patternBig)-1]
	//get top and bottom
	var top, bottom int
	switch lChar {
	case '$':
		top = 0
		bottom = 0
	case 'A':
		top = startPos[1]
		bottom = top + totalChar[1] - 1

	case 'C':
		top = startPos[2]
		bottom = top + totalChar[2] - 1

	case 'G':
		top = startPos[3]
		bottom = top + totalChar[3] - 1

	case 'T':
		top = startPos[4]
		bottom = top + totalChar[4] - 1
	}

	pattern := patternBig[0:(len(patternBig) - 1)]
	for len(pattern) > 0 {
		lChar = pattern[len(pattern)-1]
		var newTop, newBottom int

		switch lChar {
		case '$':
			newTop = lTally[top-1][0]
			newBottom = lTally[bottom][0]
		case 'A':
			newTop = lTally[top-1][1]
			newBottom = lTally[bottom][1]
		case 'C':
			newTop = lTally[top-1][2]
			newBottom = lTally[bottom][2]
		case 'G':
			newTop = lTally[top-1][3]
			newBottom = lTally[bottom][3]
		case 'T':
			newTop = lTally[top-1][4]
			newBottom = lTally[bottom][4]
		}

		//indicates that pattern actually exists
		if newBottom-newTop > 0 {
			//switch it to actual index
			switch lChar {
			case '$':
				top = 1
				bottom = 1
			case 'A':
				top = startPos[1] + newTop
				bottom = startPos[1] + newBottom - 1
			case 'C':
				top = startPos[2] + newTop
				bottom = startPos[2] + newBottom - 1
			case 'G':
				top = startPos[3] + newTop
				bottom = startPos[3] + newBottom - 1
			case 'T':
				top = startPos[4] + newTop
				bottom = startPos[4] + newBottom - 1
			}
			pattern = pattern[:(len(pattern) - 1)]
			// no full matches
		} else {
			empty := []int{}
			return empty, 0
		}

		// if bottom == top {
		// 	empty := []int{}
		// 	return empty, 0
		// }

	}

	hits := []int{}
	for i := top; i <= bottom; i++ {
		hits = append(hits, suffArr[i])
	}

	if len(hits) > 0 {
		return hits, len(patternBig)
	} else {
		return hits, 0
	}
}

// ---CIGAR function code
type info struct {
	name string
	x    string
	y    string
}

type Trace = int

const (
	Diag Trace = iota
	Up
	Left
)

// return data from Process
type outDataCigar struct {
	name    string
	x       string
	y       string
	score   int
	y_start int
	y_end   int
	cigar   string
}

func inputProcess(input string) []info {
	f, err := os.Open(input)
	check(err)
	defer f.Close()

	s := bufio.NewScanner(f)
	allData := []info{}

	for s.Scan() {
		var data info
		data.name = s.Text()
		s.Scan()
		data.x = s.Text()
		s.Scan()
		data.y = s.Text()

		allData = append(allData, data)
	}

	return allData
}

func fitting(mismatch_penalty int, gap_penalty int, data info) outDataCigar {
	//contains tracebacks
	direction := make([][]int, len(data.y)+1)
	for i := 0; i < len(data.y)+1; i++ {
		row := make([]int, len(data.x)+1)
		direction[i] = row
	}
	//scores of alignments
	manhattan := make([][]int, len(data.y)+1)
	for i := 0; i < len(data.y)+1; i++ {
		row := make([]int, len(data.x)+1)
		manhattan[i] = row
	}

	//initialize gap penalty along y-axis
	//for manhattan and direction
	for i := 0; i < len(data.y)+1; i++ {
		manhattan[i][0] = 0
		direction[i][0] = Up
	}

	//initialize gap penalty along x-axis
	//for manhattan and direction
	for i := 0; i < len(data.x)+1; i++ {
		manhattan[0][i] = gap_penalty * i * -1
		direction[0][i] = Left
	}

	//TESTING INITIAL BUILD OF ARRAY
	// displayManhattan(manhattan)

	//fill manhattan and direction row by row
	for i := 1; i < len(manhattan); i++ {
		for j := 1; j < len(manhattan[i]); j++ {
			traceBack := Up
			manhattan[i][j] = 0

			//looks at potential values
			fromAbove := manhattan[i-1][j] - gap_penalty
			fromLeft := manhattan[i][j-1] - gap_penalty
			fromDiag := manhattan[i-1][j-1]
			if data.x[j-1] != data.y[i-1] {
				fromDiag -= mismatch_penalty
			}

			//selects the highest value path
			highVal := fromAbove
			if fromLeft > highVal {
				highVal = fromLeft
				traceBack = Left
			}
			if fromDiag > highVal {
				highVal = fromDiag
				traceBack = Diag
			}

			manhattan[i][j] = highVal
			direction[i][j] = traceBack

			// displayManhattan(manhattan)
			// displayDirection(direction)

		}

		//TESTING ROW BUILD OF MANHATTAN
		// displayManhattan(manhattan)

		//TESTING ROW BUILD OF DIRECTION
		// displayDirection(direction)
	}

	//TESTING FINAL DIRECTION MATRIX
	// displayDirection(direction)

	//end Y_position
	endY := 0
	col_num := len(manhattan[0])
	for i := 1; i < len(manhattan); i++ {
		if manhattan[i][col_num-1] > manhattan[endY][col_num-1] {
			endY = i
		}
	}

	//create string which stores ideal path (fill from end)
	idealPath := ""
	row := endY
	col := col_num - 1

	for col != 0 {
		switch direction[row][col] {
		case Up:
			idealPath = "D" + idealPath
			row--
		case Left:
			idealPath = "I" + idealPath
			col--
		case Diag:
			if manhattan[row][col] == manhattan[row-1][col-1] {
				idealPath = "=" + idealPath
			} else {
				idealPath = "X" + idealPath
			}
			row--
			col--
		}
	}
	startY := row

	// //TESTING IDEAL PATH
	// fmt.Println(idealPath)

	// //convert string to integer character pairs
	idealPairs := ""
	symbol := idealPath[0]
	length := 1
	for i := 1; i < len(idealPath); i++ {
		if idealPath[i] == symbol {
			length++
		} else {
			idealPairs = idealPairs + strconv.Itoa(length) + string(symbol)
			symbol = idealPath[i]
			length = 1
		}
	}
	idealPairs = idealPairs + strconv.Itoa(length) + string(symbol)
	// fmt.Println(idealPairs)

	return outDataCigar{name: data.name, x: data.x, y: data.y,
		score: manhattan[endY][len(manhattan[0])-1], y_start: startY, y_end: endY, cigar: idealPairs}
}

func displayManhattan(manhattan [][]int) {
	fmt.Println()
	for i := 0; i < len(manhattan); i++ {
		for j := 0; j < len(manhattan[0]); j++ {
			fmt.Printf("%d ", manhattan[i][j])
		}
		fmt.Println()
	}
}

func displayDirection(direction [][]int) {
	fmt.Println()
	for i := 0; i < len(direction); i++ {
		for j := 0; j < len(direction[0]); j++ {
			switch direction[i][j] {
			case Up:
				fmt.Printf("U ")
			case Diag:
				fmt.Printf("D ")
			case Left:
				fmt.Printf("L ")
			}
		}
		fmt.Println()
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func Check(e error) {
	if e != nil {
		panic(e)
	}
}
