package main

import (
	"fmt"
	"github.com/phoenixcoder/slack-golang-sdk/slashcmd"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testRegJson = `{
                           "Command1": {
                               "reservedKeywords" : ["reservedKeywords1"],
                               "functions" : {
                                   "functions1" : {
                                       "usage" : "usage1",
                                       "description" : "description1",
                                       "manual" : "manual1"
                                   }
                               }
                           },
                           "command2": {
                               "reservedKeywords" : ["reservedKeywords2"],
                               "functions" : {
                                   "Functions2" : {
                                       "usage" : "usage2",
                                       "description" : "description2",
                                       "manual" : "manual2"
                                   }
                               }
                           }
                       }`
	testCommand     = "command"
	testResKeywords = "reservedKeywords"
	testFunctions   = "functions"
	testUsage       = "usage"
	testDesc        = "description"
	testMan         = "manual"
)

func TestGetFunctionRecord(t *testing.T) {
	fmtStr := "%s%d"
	cmdReg, err := NewCommandRegistry([]byte(testRegJson))
	assert.Nil(t, err)

	for i := 1; i <= 2; i++ {
		cmd := &slashcmd.Info{
			Command:   fmt.Sprintf(fmtStr, testCommand, i),
			Arguments: []string{fmt.Sprintf(fmtStr, testFunctions, i)},
		}
		assert.Contains(t, (*cmdReg)[cmd.Command].ReservedKeywords, fmt.Sprintf(fmtStr, testResKeywords, i))
		funcRec, err := cmdReg.getFunctionRecord(cmd)
		assert.Nil(t, err)
		assert.Equal(t, funcRec.Usage, fmt.Sprintf(fmtStr, testUsage, i))
		assert.Equal(t, funcRec.Description, fmt.Sprintf(fmtStr, testDesc, i))
		assert.Equal(t, funcRec.Manual, fmt.Sprintf(fmtStr, testMan, i))
	}
}

func TestLoadRegistryFromMalformedContents(t *testing.T) {
	testMalRegJson := `{
           malformed
        }`
	cmdReg, err := NewCommandRegistry([]byte(testMalRegJson))
	assert.Nil(t, cmdReg)
	assert.NotNil(t, err)
}

func TestCommandNotFoundError(t *testing.T) {
	cmdReg := make(commandRegistry)

	cmd := &slashcmd.Info{}
	funcRec, err := cmdReg.getFunctionRecord(cmd)
	assert.Nil(t, funcRec)
	assert.NotNil(t, err)

	_, ok := err.(CommandNotFoundError)
	assert.True(t, ok)
}

func TestFunctionNotFoundError(t *testing.T) {
	cmdReg, err := NewCommandRegistry([]byte(testRegJson))
	assert.Nil(t, err)
	cmd := &slashcmd.Info{
		Command:   testCommand,
		Arguments: []string{testFunctions},
	}
	funcRec, err := cmdReg.getFunctionRecord(cmd)
	assert.Nil(t, funcRec)
	assert.NotNil(t, err)

	_, ok := err.(FunctionNotFoundError)
	assert.True(t, ok)
}

func TestArgsNotFoundError(t *testing.T) {
	cmdReg := make(commandRegistry)

	cmd := &slashcmd.Info{
		Command: testCommand,
	}
	funcRec, err := cmdReg.getFunctionRecord(cmd)
	assert.Nil(t, funcRec)
	assert.NotNil(t, err)

	_, ok := err.(ArgsNotFoundError)
	assert.True(t, ok)
}
