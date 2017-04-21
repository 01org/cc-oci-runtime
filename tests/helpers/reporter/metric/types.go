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

//Environment struct for the environment JSON
type Environment struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

//Result struct for the tests JSON
type Result struct {
	Pass int `json:"pass"`
	Fail int `json:"fail"`
	Skip int `json:'skip"`
}

//Test struct for the tests JSON
type Test struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Params      Result
}

//TestRegex struct to find regex from the logfile
type TestRegex struct {
	LogFile   string
	SkipField string
	FailField string
	PassField string
}

//CodeRegex struct to find regex from the logfile
type CodeRegex struct {
	LogFile               string
	LineCoverageField     string
	FunctionCoverageField string
}

//ReturnCode struct for  the build JSON
type ReturnCode struct {
	Success bool `json:"succes"`
}

//Build struct for the build JSON
type Build struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Params ReturnCode
}

//Coverage struct for the code JSON
type Coverage struct {
	Line     float64 `json:"line"`
	Function float64 `json:"function"`
}

//Code struct for the code JSON
type Code struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Params Coverage
}
