package configreader

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

type ConfigReader struct {
	WalSize             int           `yaml:"wal_size"`
	WalBufferCapacity   int           `yaml:"wal_buffer_capacity"`
	MaxNumberOfSegments int           `yaml:"max_number_of_segments"`
	MemtableTrashold    float64       `yaml:"memtable_trashold"`
	CacheCapacity       int           `yaml:"cache_capacity"`
	MemtableSize        int           `yaml:"memtable_size"`
	MemtableStructure   string        `yaml:"memtable_structure"`
	Compaction          string        `yaml:"compaction"`
	LSMLevelMax         int           `yaml:"lsm_level_max"`
	LSMlevel0Number     int           `yaml:"lsm_max_l0_number"`
	LSMlevel1Number     int           `yaml:"lsm_max_l1_number"`
	LSMmultiplier       int           `yaml:"lsm_leveled_multiplier"`
	TokenBucketCapacity int           `yaml:"token_bucket_capacity"`
	TokenBucketDuration time.Duration `yaml:"token_bucket_duration"`
}

func (config *ConfigReader) ReadConfig() {
	configData, err := ioutil.ReadFile("./Data/ConfigurationFile/configuration.yaml")
	if err != nil || len(configData) == 0 {
		config.WalSize = 10
		config.WalBufferCapacity = 3
		config.MaxNumberOfSegments = 10
		config.MemtableSize = 10
		config.MemtableTrashold = 0.8
		config.MemtableStructure = "btree"
		config.CacheCapacity = 5
		config.Compaction = "leveled"
		config.LSMLevelMax = 4
		config.LSMlevel0Number = 8
		config.LSMlevel1Number = 10
		config.LSMmultiplier = 10
		config.TokenBucketCapacity = 3
		config.TokenBucketDuration = 3000000000
	} else {
		err := yaml.Unmarshal(configData, &config)
		if err != nil {
			log.Fatal(err)
		}
	}
	// fmt.Println(config.Compaction)
}