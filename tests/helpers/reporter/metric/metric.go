// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metric

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

//GetEnvironment detects the envieronment and return its struct
func GetEnvironment() Environment {
	environment := Environment{}
	var osRelease = "/usr/lib/os-release"
	var id, versionID string
	if _, err := os.Stat(osRelease); err != nil {
		osRelease = "/etc/os-release"
	}

	scanner, inFile := readFile(osRelease)
	for scanner.Scan() {
		info := strings.Split(scanner.Text(), "=")
		switch value := info[0]; value {
		case "ID":
			id = info[1]
		case "VERSION_ID":
			versionID = info[1]
		}
	}
	defer inFile.Close()

	containerImage, _ :=
		os.Readlink("/usr/share/clear-containers/clear-containers.img")

	vm, _ :=
		os.Readlink("/usr/share/clear-containers/vmlinux.container")
	environment.Name =
		id + versionID + "-" + containerImage
	environment.Description =
		id + " " + versionID + " - " + containerImage + " " + vm
	os.Readlink("/usr/share/clear-containers/clear-containers.img")
	return environment
}

func readFile(logfile string) (*bufio.Scanner, *os.File) {
	inFile, err := os.Open(logfile)
	if err != nil {
		fmt.Println("error:", err)
	}
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	return scanner, inFile

}

func getValue(logFile, testRegex, ignore string) string {
	re := regexp.MustCompile(testRegex)
	scanner, inFile := readFile(logFile)
	for scanner.Scan() {
		if re.FindAllStringSubmatch(scanner.Text(), -1) != nil {
			result := strings.Replace(scanner.Text(), testRegex, "", -1)
			result = strings.Replace(result, ignore, "", -1)
			defer inFile.Close()
			return result
		}
	}
	defer inFile.Close()
	return ""
}

func compareReturnCode(returnCode string) bool {
	if strings.Compare(returnCode, "0") == 0 {
		return true
	}
	return false

}

//WriteJSON writes a JSON file, it receives the JSON filename and it's content
func WriteJSON(jsonFile string, jsonContent interface{}) {
	//Replace JSON file underscores and add .json extension
	jsonFile = strings.Replace(jsonFile, " ", "_", -1) + ".json"
	environment := GetEnvironment()
	directory := environment.Name
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0644)
	}

	identedJSON, err := json.MarshalIndent(jsonContent, "", "\t")
	if err != nil {
		fmt.Println("error:", err)
	}
	ioutil.WriteFile(directory+"/"+jsonFile, identedJSON, 0644)
}

func getCount(logFile, testRegex, ignoreString string) int {

	var counter = 0
	re := regexp.MustCompile(testRegex)
	scanner, inFile := readFile(logFile)
	for scanner.Scan() {
		if re.FindAllStringSubmatch(scanner.Text(), -1) != nil {
			if strings.Contains(scanner.Text(), ignoreString) {
				if len(ignoreString) > 0 {
					continue
				}
			}
			counter++

		}
	}
	defer inFile.Close()
	return counter
}

//GetTestResult query the received logfile and returns the number of occurences
func GetTestResult(tr TestRegex, testResult Test) Test {

	testResult.Params.Pass =
		getCount(tr.LogFile, tr.PassField, tr.SkipField)
	testResult.Params.Fail =
		getCount(tr.LogFile, tr.FailField, "")
	testResult.Params.Skip =
		getCount(tr.LogFile, tr.SkipField, "")
	return testResult

}

//GetCodeCoverage finds the code coverage from CodeRegex
func GetCodeCoverage(tr CodeRegex, codeCoverage Code) Code {

	codeCoverage.Params.Line, _ = strconv.ParseFloat(
		getValue(tr.LogFile, tr.LineCoverageField, "%"), 64)
	codeCoverage.Params.Function, _ = strconv.ParseFloat(
		getValue(tr.LogFile, tr.FunctionCoverageField, "%"), 64)

	return codeCoverage

}

//GetJSONReturnCode finds the return code from the regex
func GetJSONReturnCode(logFile, name, regex string) Build {
	build := Build{
		Name: name,
		Type: "boolean",
	}
	var returnCode = getValue(logFile, regex, "")
	build.Params.Success = compareReturnCode(returnCode)
	return build
}

//WriteJSONReturnCode writes the JSON return code File
func (build Build) WriteJSONReturnCode() {
	WriteJSON(build.Name, build)
}

//WriteJSONTestResult writes the JSON test result
func (testResult Test) WriteJSONTestResult() {
	WriteJSON(testResult.Name, testResult)
}

//WriteJSONEnvironment writes the JSON environment
func (environment Environment) WriteJSONEnvironment(logFile string) {
	WriteJSON(logFile, environment)
}

//WriteJSONCodeCoverage writes the JSON code coverage
func (codeCoverage Code) WriteJSONCodeCoverage() {
	WriteJSON(codeCoverage.Name, codeCoverage)
}
