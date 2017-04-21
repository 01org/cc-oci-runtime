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

package main

import (
	"github.com/01org/cc-oci-runtime/tests/helpers/reporter/metric"
)

var LogPath string

func writeJSONEnvironment() {
	const JSONFile = "environment"
	environment := metric.GetEnvironment()
	environment.WriteJSONEnvironment(JSONFile)
}

func writeTestResult() {

	//Write JSON for Docker Test Results

	dockerTestRegex := metric.TestRegex{
		LogFile: LogPath + "docker-integration-tests.html",
		//		JSONFile:  "docker_test_result.json",
		SkipField: "# skip",
		FailField: ">not ok ",
		PassField: ">ok ",
	}

	dockerTestResult := metric.Test{
		Name:        "docker test result",
		Description: "docker test result for cc-oci-runtime",
		Type:        "integration tests",
	}

	dockerTest :=
		metric.GetTestResult(dockerTestRegex, dockerTestResult)
	dockerTest.WriteJSONTestResult()

	//Write JSON for Unit Test Results

	unitTestRegex := metric.TestRegex{
		LogFile: LogPath + "unit_tests.log",
		//		JSONFile:  "unit_test_result.json",
		SkipField: "^SKIP",
		FailField: "^FAIL",
		PassField: "^PASS",
	}

	unitTestResult := metric.Test{
		Name:        "unit test result",
		Description: "unit tests for cc-oci-runtime",
		Type:        "unit tests",
	}

	unitTest := metric.GetTestResult(unitTestRegex, unitTestResult)
	unitTest.WriteJSONTestResult()

	//Write JSON for Functional Test Results
	functionalTestsRegex := metric.TestRegex{
		LogFile: LogPath + "functional_tests.log",
		//		JSONFile:  "functional_test_result.json",
		SkipField: "# skip",
		FailField: "^not ok",
		PassField: "^ok ",
	}

	testResult := metric.Test{
		Name:        "functional test result",
		Description: "functional tests result for cc-oci-runtime",
		Type:        "functional tests",
	}

	functionalTest :=
		metric.GetTestResult(functionalTestsRegex, testResult)
	functionalTest.WriteJSONTestResult()

	//Write JSON for Valgrind test results
	valgrindRegex := metric.TestRegex{
		LogFile: LogPath + "valgrind_tests.log",
		//		JSONFile:  "valgrind_test_result.json",
		SkipField: "^SKIP:",
		FailField: "^FAIL:",
		PassField: "^PASS:",
	}

	valgrindResult := metric.Test{
		Name:        "valgrind test result",
		Description: "valgrind for cc-oci-runtime",
		Type:        "valgrind tests",
	}

	valgrind := metric.GetTestResult(valgrindRegex, valgrindResult)
	valgrind.WriteJSONTestResult()

	//Write JSON for code Coverage
	codeCoverageRegex := metric.CodeRegex{
		LogFile: LogPath + "test_summary.log",
		//	JsonFile:
		LineCoverageField:     "Lines Coverage: ",
		FunctionCoverageField: "Functions Coverage: ",
	}

	codeCoverageResult := metric.Code{
		Name: "code_coverage",
		Type: "Code coverage",
	}

	codeCoverage :=
		metric.GetCodeCoverage(codeCoverageRegex, codeCoverageResult)
	codeCoverage.WriteJSONCodeCoverage()
}

func writeReturnCode() {

	//Return codes are stored in the test_summary.log

	var LogFile = LogPath + "test_summary.log"
	const (
		CppcheckRE         = "cppcheck return code: "
		CppcheckName       = "ccpcheck"
		AutogenRE          = "autogen return code: "
		AutogenName        = "autogen"
		MakeRE             = "make return code: "
		MakeName           = "make"
		MakeInstallRE      = "make install return code: "
		MakeInstallName    = "make_install"
		ProxyTestName      = "check_proxy"
		ProxyTestRE        = "check-proxy return code: "
		UnitTestName       = "unit_test"
		UnitTestRE         = "unit tests return code: "
		FunctionalTestRE   = "functional tests return code: "
		FunctionalTestName = "functional_test"
		ValgrindName       = "valgrind"
		ValgrindRE         = "valgrind tests return code: "
		CoverageName       = "coverage"
		CoverageRE         = "coverage return code: "
		DockerTestName     = "docker_test"
		DockerTestRE       = "docker tests return code: "
	)

	//Write ccpcheck JSON file
	cppcheck :=
		metric.GetJSONReturnCode(LogFile, CppcheckName, CppcheckRE)
	cppcheck.WriteJSONReturnCode()

	//Write autogen JSON file
	autogen := metric.GetJSONReturnCode(LogFile, AutogenName, AutogenRE)
	autogen.WriteJSONReturnCode()

	//Write make JSON file
	make := metric.GetJSONReturnCode(LogFile, MakeName, MakeRE)
	make.WriteJSONReturnCode()

	//Write make install JSON file
	makeInstall :=
		metric.GetJSONReturnCode(
			LogFile, MakeInstallName, MakeInstallRE)
	makeInstall.WriteJSONReturnCode()

	//Write Proxy Test JSON file
	proxyTest :=
		metric.GetJSONReturnCode(LogFile, ProxyTestName, ProxyTestRE)
	proxyTest.WriteJSONReturnCode()

	//Write Unit Test Name JSON file
	unitTest :=
		metric.GetJSONReturnCode(LogFile, UnitTestName, UnitTestRE)
	unitTest.WriteJSONReturnCode()

	//Write Functional Test JSON file
	functionalTest :=
		metric.GetJSONReturnCode(
			LogFile, FunctionalTestName, FunctionalTestRE)
	functionalTest.WriteJSONReturnCode()

	//Write Valgrind JSON file
	valgrind :=
		metric.GetJSONReturnCode(LogFile, ValgrindName, ValgrindRE)
	valgrind.WriteJSONReturnCode()

	//Write Code Coverage JSON file
	coverage :=
		metric.GetJSONReturnCode(LogFile, CoverageName, CoverageRE)
	coverage.WriteJSONReturnCode()

	//Write Docker JSON file
	docker :=
		metric.GetJSONReturnCode(LogFile, DockerTestName, DockerTestRE)
	docker.WriteJSONReturnCode()

}

func main() {

	LogPath = "../test_logs/"
	writeJSONEnvironment()
	writeReturnCode()
	writeTestResult()
}
