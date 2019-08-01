package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/phoenixcoder/slack-golang-sdk/slashcmd"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	cmdNotRegErrFmt    = "Command is not registered. Command Name: '%s'"
	noArgsErrFmt       = "No arguments received."
	logRegUrl          = "Registry URL: %s\n"
	errDlRegFmt        = "Could not download registry contents. Error: %v\n"
	errReadRegFmt      = "Could not read registry contents. Error: %v\n"
	errFuncNotFoundFmt = "We're embarassed for you, but we don't know a '%s'."
)

type CommandNotFoundError error
type FunctionNotFoundError error
type ArgsNotFoundError error
type commandRegistry map[string]commandRecord
type functionRegistry map[string]functionRecord

func (cr *commandRegistry) UnmarshalJSON(text []byte) error {
	var tempMap map[string]commandRecord
	json.Unmarshal(text, &tempMap)

	*cr = make(commandRegistry)
	for key, val := range tempMap {
		(*cr)[strings.ToLower(key)] = val
	}

	return nil
}

func (fr *functionRegistry) UnmarshalJSON(text []byte) error {
	var tempMap map[string]functionRecord
	json.Unmarshal(text, &tempMap)
	*fr = make(functionRegistry)
	for key, val := range tempMap {
		(*fr)[strings.ToLower(key)] = val
	}

	return nil
}

type commandRecord struct {
	ReservedKeywords []string         `json:"reservedKeywords"`
	Functions        functionRegistry `json:"functions"`
}

type functionRecord struct {
	// Usage is a description of how to use the function with the command.
	Usage string `json: "usage"`
	// Description is a description of what the function does.
	Description string `json: "description"`
	// Manual is a location for additional information on the function.
	Manual string `json: "manual"`
}

func (cr *commandRegistry) getFunctionRecord(cmd *slashcmd.Info) (*functionRecord, error) {
	cmdRec, cmdRecOk := (*cr)[strings.ToLower(cmd.Command)]
	if !cmdRecOk {
		return nil, CommandNotFoundError(fmt.Errorf(cmdNotRegErrFmt, cmd.Command))
	}

	if len(cmd.Arguments) <= 0 {
		return nil, ArgsNotFoundError(errors.New(noArgsErrFmt))
	}
	funcName := cmd.Arguments[0]
	funcRec, funcRecOk := cmdRec.Functions[strings.ToLower(funcName)]
	if !funcRecOk {
		return nil, FunctionNotFoundError(fmt.Errorf(errFuncNotFoundFmt, funcName))
	}

	return &funcRec, nil
}

func NewCommandRegistry(contents []byte) (*commandRegistry, error) {
	var cr commandRegistry
	if err := json.Unmarshal(contents, &cr); err != nil {
		return nil, err
	}

	return &cr, nil
}

func NewCommandRegistryFromFile(location string) (*commandRegistry, error) {
	contents, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, err
	}

	return NewCommandRegistry(contents)
}

func NewCommandRegistryFromUrl(url string) (*commandRegistry, error) {
	log.Printf(logRegUrl, url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf(errDlRegFmt, err)
		return nil, err
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf(errReadRegFmt, err)
		return nil, err
	}

	return NewCommandRegistry(contents)
}
