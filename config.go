package main

import (
	"bytes"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
	"github.com/siddontang/go/ioutil2"
)

type SchedulerConfig struct {
	PDAddrs       []string `toml:"pd"`
	ShuffleLeader bool     `toml:"shuffle-leader"`
	ShuffleRegion bool     `toml:"shuffle-region"`
}

// SuiteConfig is the configuration for all test cases.
type SuiteConfig struct {
	// Names contains all cases to be run later.
	Names []string `toml:"names"`
	// Concurrency is the concurrency to run all cases.
	Bank         BankCaseConfig         `toml:"bank"`
	Bank2        Bank2CaseConfig        `toml:"bank2"`
	Ledger       LedgerConfig           `toml:"ledger"`
	CRUD         CRUDCaseConfig         `toml:"crud"`
	Log          LogCaseConfig          `toml:"log"`
	BlockWriter  BlockWriterCaseConfig  `toml:"block_writer"`
	MVCCBank     BankCaseConfig         `toml:"mvcc_bank"`
	Sysbench     SysbenchCaseConfig     `toml:"sysbench"`
	SqllogicTest SqllogicTestCaseConfig `toml:"sqllogic_test"`
	SmallWriter  SmallWriterCaseConfig  `toml:"small_writer"`
}

// BankCaseConfig is for bank test case.
type BankCaseConfig struct {
	// NumAccounts is total accounts
	NumAccounts int `toml:"num_accounts"`
	TableNum    int `toml:"table_num"`
	Concurrency int `toml:"concurrency"`
}

// Bank2CaseConfig is for bank2 test case.
type Bank2CaseConfig struct {
	// NumAccounts is total accounts
	NumAccounts int    `toml:"num_accounts"`
	Contention  string `toml:"contention"`
	Concurrency int    `toml:"concurrency"`
}

// LedgerConfig is for ledger test case.
type LedgerConfig struct {
	NumAccounts int `toml:"num_accounts"`
	Concurrency int `toml:"concurrency"`
}

// CRUDCaseConfig is for CRUD test case.
type CRUDCaseConfig struct {
	UserCount int `toml:"user_count"`
	PostCount int `toml:"post_count"`
	// Insert/delete users every interval.
	UpdateUsers int `toml:"update_users"`
	// Insert/delete posts every interval.
	UpdatePosts int `toml:"update_posts"`
	Concurrency int `toml:"concurrency"`
}

// LogCaseConfig is for Log test case
type LogCaseConfig struct {
	MaxCount    int `toml:"max_count"`
	DeleteCount int `toml:"delete_count"`
	TableNum    int `toml:"table_num"`
	Concurrency int `toml:"concurrency"`
}

// BlockWriterCaseConfig is for block write test case
type BlockWriterCaseConfig struct {
	TableNum    int `toml:"table_num"`
	Concurrency int `toml:"concurrency"`
}

// SysbenchCaseConfig is for sysbench test case
type SysbenchCaseConfig struct {
	TableCount int    `toml:"table_count"`
	TableSize  int    `toml:"table_size"`
	Threads    int    `toml:"threads"`
	MaxTime    int    `toml:"max_time"`
	DBName     string `toml:"database"`
	LuaPath    string `toml:"lua_path"`
}

// SqllogictestCaseConfig is for sqllogic_test test case
type SqllogicTestCaseConfig struct {
	TestPath  string `toml:"test_path"`
	SkipError bool   `toml:"skipError"`
	Parallel  int    `toml:"parallel"`
	DBName    string `toml:"database"`
}

// SmallWriterCase is for small write test case
type SmallWriterCaseConfig struct {
	Concurrency int `toml:"concurrency"`
}

// Config is the configuration for the stability test.
type Config struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	PD       string `toml:"pd"`

	// Cluster ClusterConfig `toml:"cluster"`
	// Nemeses     NemesesConfig     `toml:"nemeses"`
	Suite     SuiteConfig     `toml:"suite"`
	Scheduler SchedulerConfig `toml:"scheduler"`
}

// ParseConfig parses the configuration file.
func ParseConfig(path string) (*Config, error) {
	cfg := new(Config)
	if err := parseConfig(path, &cfg); err != nil {
		return nil, errors.Trace(err)
	}

	// adjust Cluster.
	// cfg.Cluster.adjust()

	return cfg, nil
}

func parseConfig(path string, cfg interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Trace(err)
	}

	if err = toml.Unmarshal(data, cfg); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func writeConfig(path string, cfg *Config) error {
	var buf bytes.Buffer
	e := toml.NewEncoder(&buf)
	err := e.Encode(cfg)
	if err != nil {
		return errors.Trace(err)
	}

	err = ioutil2.WriteFileAtomic("./config.toml", buf.Bytes(), 0644)
	if err != nil {
		return errors.Trace(err)
	}

	return nil

}
