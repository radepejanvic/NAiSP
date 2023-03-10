package lsm

import (
	readpath "NAiSP/Structures/ReadPath"
	record "NAiSP/Structures/Record"
	sstable "NAiSP/Structures/Sstable"

	"math"
	"os"
)

func (lvl *Leveled) calculateCapacity(level int) int {
	duzinaNivoa := lvl.levels[level]
	return int(math.Ceil(float64(len(duzinaNivoa)) - PERCENT*float64(lvl.lsm.Config.LSMlevel1Number)*math.Pow(float64(lvl.lsm.Config.LSMmultiplier), float64(level-1))))
}

func (lvl *Leveled) BetweenLevels(from int, to int) {

	lvl.records = make(map[*os.File]*record.Record)

	firstLevel := readpath.SortFiles(lvl.levels[from])[:lvl.calculateCapacity(from)]
	secondLevel := readpath.SortFiles(lvl.levels[to])

	// TODO -> if second level is empty -> rename lvl.first level and break the loop

	iteratorFirst := 0
	iteratorSecond := 0

	// oppening the lvl.first file from each level
	lvl.first = lvl.lsm.OpenData(firstLevel[0])

	if secondLevel != nil {
		lvl.second = lvl.lsm.OpenData(secondLevel[0])
	} else {
		lvl.second = nil
	}

	// every newly generated sstable will be added to this list
	tempSSTables := make([]*sstable.SStable, 0)

	// file counters for levels
	counterFirst := len(firstLevel)

	var counterSecond int

	if lvl.second != nil {
		counterSecond = len(secondLevel)
	} else {
		counterSecond = 0
	}

	// SSTables for reading header
	SSTableFirst := lvl.NewSSTableFromFileName(lvl.first)

	var SSTableSecond *sstable.SStable

	if lvl.second != nil {
		SSTableSecond = lvl.NewSSTableFromFileName(lvl.second)
	} else {
		SSTableSecond = nil
	}

	SSTable := lvl.NewSSTable(0, to, true)
	files, writers, counter, offsetData, offsetIndex, bf, merkle := lvl.InitSSTable(SSTable)
	tempSSTables = append(tempSSTables, SSTable)

	// necessary for generating filename
	fileCounter := 0

	// necessary for checking if file is full
	currentCapacity := 0

	lvl.fromTo = make(map[*os.File][]string)

	lvl.fromTo[lvl.first] = lvl.lsm.ReadHeader(SSTableFirst)
	if lvl.second != nil {
		lvl.fromTo[lvl.second] = lvl.lsm.ReadHeader(SSTableSecond)
	}

	beginning := true

	// data file values
	// var recordFirst *record.Record
	// var recordSecond *record.Record

	// header values
	var firstHeaderRecord *record.Record
	var lastHeaderRecord *record.Record

	for {
		// level lvl.first got empied -> rename the rest of level two
		if counterFirst == 0 && counterSecond != 0 {
			// terminal condition 1

			// empty the remaining lvl.first file
			SSTable, tempSSTables = lvl.EmptyFile(lvl.second, SSTable, currentCapacity, firstHeaderRecord, lastHeaderRecord, files, writers,
				bf, merkle, &fileCounter, counter, offsetData, offsetIndex, &tempSSTables, to, false, lvl.records[lvl.second])

			// renaming reamining files to temp files

			if iteratorSecond+1 <= len(secondLevel)-1 {
				for _, file := range secondLevel[iteratorSecond+1:] {
					// rename element
					remainingSSTable := lvl.RenameFile(fileCounter, to, file)
					fileCounter++

					// append renamed to temp
					tempSSTables = append(tempSSTables, remainingSSTable)

				}
			}

			// renaming temo files
			lvl.RenameLevel(tempSSTables)

			lvl.GenerateLevels()
			// calls between levels
			if len(tempSSTables) > int(float64(lvl.lsm.Config.LSMlevel1Number)*(math.Pow(float64(lvl.lsm.Config.LSMmultiplier), float64(to-1)))) {

				if lvl.lsm.Config.LSMLevelMax > to+1 {
					lvl.BetweenLevels(to, to+1)
				}

			}
			return

		}
		// level second got emptied lvl.first -> rename the rest of level one
		if counterFirst != 0 && counterSecond == 0 {
			// terminal condition 2

			// empty the remaining lvl.first file
			// firstHeaderRecord, lastHeaderRecord, _, SSTable
			SSTable, tempSSTables = lvl.EmptyFile(lvl.first, SSTable, currentCapacity, firstHeaderRecord, lastHeaderRecord, files, writers,
				bf, merkle, &fileCounter, counter, offsetData, offsetIndex, &tempSSTables, to, false, lvl.records[lvl.first])

			// renaming reamining files to temp files

			// for index := range firstLevel {
			// 	fmt.Println(firstLevel[index])
			// }

			if iteratorFirst+1 <= len(firstLevel)-1 {
				for _, file := range firstLevel[iteratorFirst+1:] {

					// fmt.Println("File ", iteratorFirst, " za rename -> ", file)

					// fmt.Println("File ", iteratorFirst+1, " za rename -> ", file)
					// rename element
					remainingSSTable := lvl.RenameFile(fileCounter, to, file)
					fileCounter++

					// append renamed to temp
					tempSSTables = append(tempSSTables, remainingSSTable)

				}
			}

			// renaming temo files
			lvl.RenameLevel(tempSSTables)

			lvl.GenerateLevels()
			// calls between levels
			if len(tempSSTables) > int(float64(lvl.lsm.Config.LSMlevel1Number)*(math.Pow(float64(lvl.lsm.Config.LSMmultiplier), float64(to-1)))) {

				if lvl.lsm.Config.LSMLevelMax > to+1 {
					lvl.BetweenLevels(to, to+1)
				}

			}
			return

		}

		if beginning {
			lvl.MoveBeginning(&beginning, &fileCounter, to, &tempSSTables,
				&counterFirst, &counterSecond, &iteratorFirst, &iteratorSecond, &firstLevel, &secondLevel, SSTableFirst, SSTableSecond)
			if beginning {
				continue
			}
		}

		// sequential processing of both files
		minimumFile, minimumRecord := lvl.GetMinimumRecord(&counterFirst, &counterSecond, &iteratorFirst, &iteratorSecond, &firstLevel, &secondLevel)

		// fmt.Println("Minumum File posle GetMinimumRecord-a -> ", minimumFile.Name())

		currentCapacity += int(minimumRecord.GetSize())

		// if current sstable reached capacity -> make a new one
		if currentCapacity > lvl.lsm.Config.LSMDataCapacity {

			// completing formation of SSTable
			if lvl.lsm.Config.DataFileStructure == "Single" {
				SSTable.CopyExistingToSummary(firstHeaderRecord, lastHeaderRecord, files, writers)
				SSTable.EncodeHelpersOneFileWithoutTOC(bf, merkle)

				bloomSize, summarySize := SSTable.CalculateFileSizes(files)
				sizes := []uint64{bloomSize, summarySize, offsetIndex}

				SSTable.CopyAllandWriteHeader(sizes, files, writers)

				SSTable.CloseFiles(files)
			} else {
				SSTable.CopyExistingToSummary(firstHeaderRecord, lastHeaderRecord, files, writers)
				SSTable.EncodeHelpersWithoutTOC(bf, merkle)
				SSTable.CloseFiles(files)
			}
			// and closing it

			// Initialiazing new SSTABLE -> adding last record to new SSTABLE
			fileCounter++
			SSTable = lvl.NewSSTable(fileCounter, to, true)
			currentCapacity = int(minimumRecord.GetSize())
			// initializing all necesarry files
			files, writers, counter, offsetData, offsetIndex, bf, merkle = lvl.InitSSTable(SSTable)

			tempSSTables = append(tempSSTables, SSTable)

			firstHeaderRecord = minimumRecord

			// add record
			offsetData, offsetIndex = SSTable.AddRecord(counter, offsetData, offsetIndex, minimumRecord, bf, merkle, writers)
			counter++

		} else {
			// add record
			if firstHeaderRecord == nil {
				firstHeaderRecord = minimumRecord
			}

			offsetData, offsetIndex = SSTable.AddRecord(counter, offsetData, offsetIndex, minimumRecord, bf, merkle, writers)
			counter++
		}

		// fmt.Println("AFTER WRITING TO SSTABLE -> ", minimumRecord)

		// remembers the last record -> will be necessary in header later
		lastHeaderRecord = minimumRecord

		lvl.NextRecordBetweenLevels(minimumFile, &counterFirst, &counterSecond, &iteratorFirst, &iteratorSecond, &firstLevel, &secondLevel)

		// fmt.Println("AFTER NEXT RECORD -> ", minimumRecord)

	}

}

func (lvl *Leveled) GetMinimumRecord(counterFirst, counterSecond, iteratorFirst, iteratorSecond *int, firstLevel, secondLevel *[]string) (*os.File, *record.Record) {
	// initialiazing values for first and second record

	if lvl.records[lvl.first] == nil {
		lvl.NextRecordBetweenLevels(lvl.first, counterFirst, counterSecond, iteratorFirst, iteratorSecond, firstLevel, secondLevel)

	}
	if lvl.records[lvl.second] == nil {
		lvl.NextRecordBetweenLevels(lvl.second, counterFirst, counterSecond, iteratorFirst, iteratorSecond, firstLevel, secondLevel)
	}

	// ---------------------------------------------------------------

	// comparing first and second records
	if lvl.records[lvl.first].GetKey() < lvl.records[lvl.second].GetKey() {
		return lvl.first, lvl.records[lvl.first]

	} else if lvl.records[lvl.first].GetKey() > lvl.records[lvl.second].GetKey() {
		return lvl.second, lvl.records[lvl.second]

	} else {
		if lvl.records[lvl.first].GetTimeStamp() > lvl.records[lvl.second].GetTimeStamp() {

			// TODO MOVE TO THJE NEXT RECORD IN LOWER FILE
			lvl.NextRecordBetweenLevels(lvl.second, counterFirst, counterSecond, iteratorFirst, iteratorSecond, firstLevel, secondLevel)
			return lvl.first, lvl.records[lvl.first]

		} else if lvl.records[lvl.first].GetTimeStamp() < lvl.records[lvl.second].GetTimeStamp() {

			// TODO MOVE TO THJE NEXT RECORD IN LOWER FILE
			lvl.NextRecordBetweenLevels(lvl.first, counterFirst, counterSecond, iteratorFirst, iteratorSecond, firstLevel, secondLevel)
			return lvl.second, lvl.records[lvl.second]
		}
	}
	return nil, nil
}

func (lvl *Leveled) NextRecordBetweenLevels(minimumFile *os.File, counterFirst *int, counterSecond *int, iteratorFirst *int, iteratorSecond *int, firstLevel *[]string, secondLevel *[]string) {

	nextRecord, _ := record.ReadRecord(minimumFile)
	if nextRecord == nil {
		// deleting the file and removing it from the map
		if minimumFile == lvl.first {
			*counterFirst--
			*iteratorFirst++

			if *iteratorFirst > len(*firstLevel)-1 {
				lvl.first = nil
			} else {
				lvl.first = lvl.lsm.OpenData((*firstLevel)[*iteratorFirst])
				lvl.records[lvl.first], _ = record.ReadRecord(lvl.first)
				// read header -> update map
			}
		} else {
			*counterSecond--
			*iteratorSecond++

			if *iteratorSecond > len(*secondLevel)-1 {
				lvl.second = nil
			} else {
				lvl.second = lvl.lsm.OpenData((*secondLevel)[*iteratorSecond])
				lvl.records[lvl.second], _ = record.ReadRecord(lvl.second)
				// read header -> update map
			}
		}

		minimumFile.Close()
		delete(lvl.fromTo, minimumFile)
		delete(lvl.records, minimumFile)

		// remove all files (DATA, INDEX, SUMMARY, BF, TOC, MERKLE)
		lvl.RemoveFile(minimumFile)

	} else {
		if minimumFile == lvl.first {
			lvl.records[lvl.first] = nextRecord
		} else {
			lvl.records[lvl.second] = nextRecord
		}
	}

	// ---------------------------------------------------------
	return
}

func (lvl *Leveled) MoveBeginning(beginning *bool, fileCounter *int, to int, tempSSTables *[]*sstable.SStable,
	counterFirst, counterSecond, iteratorFirst, iteratorSecond *int, firstLevel, secondLevel *[]string, SSTableFirst, SSTableSecond *sstable.SStable) {

	if lvl.fromTo[lvl.first][0] < lvl.fromTo[lvl.second][0] && lvl.fromTo[lvl.first][1] < lvl.fromTo[lvl.second][0] {
		// renaming file from first level
		lvl.first.Close()
		*tempSSTables = append(*tempSSTables, lvl.RenameFile(*fileCounter, to, lvl.first.Name()))
		delete(lvl.fromTo, lvl.first)
		// open next file from first level
		*counterFirst--
		*fileCounter++
		*iteratorFirst++

		if *iteratorFirst > len(*firstLevel)-1 {
			// no more files from level 1
			return
		} else {

			lvl.first = lvl.lsm.OpenData((*firstLevel)[*iteratorFirst])

			SSTableFirst = lvl.NewSSTableFromFileName(lvl.first)
			// read header and update map
			lvl.fromTo[lvl.first] = lvl.lsm.ReadHeader(SSTableFirst)

		}

		// continue
	} else if lvl.fromTo[lvl.second][0] < lvl.fromTo[lvl.first][0] && lvl.fromTo[lvl.second][1] < lvl.fromTo[lvl.first][0] {
		lvl.second.Close()
		*tempSSTables = append(*tempSSTables, lvl.RenameFile(*fileCounter, to, lvl.second.Name()))
		delete(lvl.fromTo, lvl.second)
		// open next file from second level
		*counterSecond--
		*fileCounter++
		*iteratorSecond++

		if *iteratorSecond > len(*secondLevel)-1 {
			// no more files from level 2
			return
		} else {

			lvl.second = lvl.lsm.OpenData((*secondLevel)[*iteratorSecond])

			SSTableSecond = lvl.NewSSTableFromFileName(lvl.second)
			// read header and update map
			lvl.fromTo[lvl.second] = lvl.lsm.ReadHeader(SSTableSecond)

		}
	} else {
		*beginning = false
	}
}
