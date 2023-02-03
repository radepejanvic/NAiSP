package readpath

import (
	bloomfilter "NAiSP/Structures/Bloomfilter"
	configreader "NAiSP/Structures/ConfigReader"
	lru "NAiSP/Structures/LRUcache"
	memtable "NAiSP/Structures/Memtable"
	sstable "NAiSP/Structures/Sstable"
	writepath "NAiSP/Structures/WritePath"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
)

const (
	DATAPATH = "./Data/Data"
)

type ReadPath struct {
	MemTable     *memtable.MemTable
	Lru          *lru.LRUCache
	BloomFilter  *bloomfilter.BloomFilter
	ConfigReader *configreader.ConfigReader
}

func (rp *ReadPath) Read(key string) []byte {

	filepath := DATAPATH + rp.ConfigReader.DataFileStructure + "/" + rp.ConfigReader.Compaction + "/"
	// First check in MemTable
	record := rp.MemTable.Find(key)
	if record != nil {
		rp.Lru.AddElement(record.GetKey(), record.GetValue())
		return record.GetValue()
	}

	// Next check in Cache
	value := rp.Lru.GetElement(key)
	if value != nil {
		return value
	}

	// Next check Bloom Filter
	found := rp.BloomFilter.Find(key)
	if !found {
		// If not in bloom we can be sure it's not there
		return nil
	}

	// Opening directory that contains data files
	folder, err := ioutil.ReadDir(filepath + "Toc")
	if err != nil {
		fmt.Println("Greska kod citanja direktorijuma: ", err)
		log.Fatal(err)
	}

	for i := 0; i < rp.ConfigReader.LSMLevelMax; i++ {
		files := getFiles(folder, i, filepath)
		if rp.ConfigReader.Compaction == "Size_tiered" || i == 0 {
			for j := len(files) - 1; j >= 0; j-- {
				Sstable := sstable.NewSStableFromTOC(files[j])
				if rp.ConfigReader.DataFileStructure == "Multiple" {
					record := Sstable.Search(key)
					if record != nil {
						return record.GetValue()
					}
				} else {
					record := Sstable.SearchOneFile(key)
					if record != nil {
						return record.GetValue()
					}
				}
			}

		} else {
			for j := 0; j < len(files); j++ {
				Sstable := sstable.NewSStableFromTOC(files[j])
				if rp.ConfigReader.DataFileStructure == "Multiple" {
					record := Sstable.Search(key)
					if record != nil {
						return record.GetValue()
					}
				} else {
					record := Sstable.SearchOneFile(key)
					if record != nil {
						return record.GetValue()
					}
				}
			}
		}
	}
	return nil
}

func getFiles(folder []fs.FileInfo, level int, filepath string) []string {
	stringlist := make([]string, 0)
	for _, file := range folder {
		if writepath.GetLevel(file.Name()) == level {
			filePath := filepath + "Toc/" + file.Name()
			stringlist = append(stringlist, filePath)
		}
		if writepath.GetLevel(file.Name()) > level {
			break
		}

	}
	return stringlist

}
