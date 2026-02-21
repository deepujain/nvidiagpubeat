//******************************************************************
//Copyright 2018 eBay Inc.
//Architect/Developer: Deepak Vasthimal

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at

// https://www.apache.org/licenses/LICENSE-2.0

//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//******************************************************************

package nvidia

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func Test_Command_TestEnv(t *testing.T) {
	util := newUtilization()
	cmd := util.command("test", "myquery")

	if len(cmd.Args) != 1 {
		t.Errorf("Expected %d, Actual %d", 1, len(cmd.Args))
	}

	if cmd.Args[0] != "localnvidiasmi" {
		t.Errorf("Expected %s, Actual %s", "localnvidiasmi", cmd.Args[0])
	}
}

func Test_Command_ProdEnv(t *testing.T) {
	util := newUtilization()
	cmd := util.command("prod", "myquery")

	if len(cmd.Args) != 3 {
		t.Errorf("Expected %d, Actual %d", 3, len(cmd.Args))
	}

	if cmd.Args[0] != "nvidia-smi" {
		t.Errorf("Expected %s, Actual %s", "nvidia-smi", cmd.Args[0])
	}

	if cmd.Args[1] != "--query-gpu=myquery" {
		t.Errorf("Expected %s, Actual %s", "--query-gpu=myquery", cmd.Args[1])
	}

	if cmd.Args[2] != "--format=csv" {
		t.Errorf("Expected %s, Actual %s", "--format=csv", cmd.Args[2])
	}
}

func Test_Run_TestEnv(t *testing.T) {
	util := newUtilization()
	query := "utilization.gpu,utilization.memory,memory.total,memory.free,memory.used,temperature.gpu,pstate"
	cmd := util.command("test", query)
	os.Setenv("PATH", ".")
	output, _ := util.run(cmd, 4, query, NewLocal())
	if output == nil {
		//TODO fix unit test case
		//t.Errorf("output cannot be nil")
	}
}

func Test_Run_ProdEnv(t *testing.T) {
	util := newUtilization()
	query := "utilization.gpu,utilization.memory,memory.total,memory.free,memory.used,temperature.gpu,pstate"
	cmd := util.command("prod", query)
	t.Logf("Command: %s", cmd.Path)
	output, _ := util.run(cmd, 4, query, MockLocal{})

	for _, o := range output {
		if o == nil {
			t.Errorf("output cannot be nil.")
		}
	}
}

func Test_Event_Contains_Type_Field(t *testing.T) {
	util := newUtilization()
	query := "utilization.gpu,utilization.memory,memory.total,memory.free,memory.used,temperature.gpu,pstate"
	cmd := util.command("prod", query)
	t.Logf("Command: %s", cmd.Path)

	output, _ := util.run(cmd, 4, query, MockLocal{})
	t.Logf("Nr. of Events: %d", len(output))

	for _, o := range output {
		if o["type"] != "nvidiagpubeat" {
			t.Errorf("event does not contain 'type' field equal to 'nvidiagpubeat'")
		}
	}
}

// stringReaderMock is an Action that reads from a fixed string (for testing driver_version etc.).
type stringReaderMock struct{ content string }

func (m stringReaderMock) start(cmd *exec.Cmd) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(m.content))
}

// Test_DriverVersion_WithPatchVersion verifies driver_version with patch (e.g. 460.91.03) is stored as string (Fix #36).
func Test_DriverVersion_WithPatchVersion(t *testing.T) {
	util := newUtilization()
	query := "--query-gpu=name,driver_version,count"
	// Header line (skipped: contains gpu_uuid); then one data line with driver_version 460.91.03
	csvContent := "name, driver_version, count, gpu_uuid\nTesla P100, 460.91.03, 1\n"
	mock := stringReaderMock{content: csvContent}
	cmd := util.command("prod", query)
	events, err := util.run(cmd, 1, query, mock)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(events) != 1 || events[0] == nil {
		t.Fatal("expected one event")
	}
	ev := events[0]
	drv, ok := ev["driver_version"]
	if !ok {
		t.Fatal("event missing driver_version")
	}
	if s, ok := drv.(string); !ok || s != "460.91.03" {
		t.Errorf("driver_version should be string \"460.91.03\", got %T %v", drv, drv)
	}
}
