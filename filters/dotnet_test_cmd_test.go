package filters

import (
	"strings"
	"testing"
)

func TestFilterDotnetTestAllPassing(t *testing.T) {
	raw := `Microsoft (R) Test Execution Command Line Tool Version 17.8.0 (x64)
Copyright (c) Microsoft Corporation.  All rights reserved.

Starting test execution, please wait...
A total of 1 test files matched the specified pattern.

Passed!  - Failed:     0, Passed:    28, Skipped:     0, Total:    28, Duration: 1.2 s - MyApp.Tests.dll (net8.0)`

	got, err := filterDotnetTestCmd(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "all 28 tests passed") {
		t.Errorf("expected 'all 28 tests passed', got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output: %s", got)
}

func TestFilterDotnetTestWithFailures(t *testing.T) {
	raw := `Microsoft (R) Test Execution Command Line Tool Version 17.8.0 (x64)
Copyright (c) Microsoft Corporation.  All rights reserved.

Starting test execution, please wait...
A total of 1 test files matched the specified pattern.

  Failed CreateUser_ShouldReturnBadRequest_WhenEmailInvalid [42 ms]
  Error Message:
   Assert.Equal() Failure
            Expected: 400
            Actual:   200
  Stack Trace:
     at MyApp.Tests.UserControllerTests.CreateUser_ShouldReturnBadRequest_WhenEmailInvalid() in C:\src\MyApp.Tests\UserControllerTests.cs:line 45

  Failed DeleteUser_ShouldReturn404_WhenNotFound [18 ms]
  Error Message:
   Assert.NotNull() Failure
            Expected: not null
            Actual:   null
  Stack Trace:
     at MyApp.Tests.UserControllerTests.DeleteUser_ShouldReturn404_WhenNotFound() in C:\src\MyApp.Tests\UserControllerTests.cs:line 72

Passed  GetUser_ShouldReturnOk [5 ms]
Passed  ListUsers_ShouldReturnAll [8 ms]
Passed  UpdateUser_ShouldReturnOk [12 ms]
Passed  GetUserById_ShouldReturnCorrectUser [3 ms]
Passed  SearchUsers_ShouldFilterByName [15 ms]
Passed  CreateUser_ShouldReturnCreated [22 ms]
Passed  GetRoles_ShouldReturnAll [4 ms]
Passed  AssignRole_ShouldWork [9 ms]

Failed!  - Failed:     2, Passed:     8, Skipped:     0, Total:    10, Duration: 1.8 s - MyApp.Tests.dll (net8.0)`

	got, err := filterDotnetTestCmd(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Failures should be shown
	if !strings.Contains(got, "CreateUser_ShouldReturnBadRequest_WhenEmailInvalid") {
		t.Errorf("expected first failure name, got:\n%s", got)
	}
	if !strings.Contains(got, "DeleteUser_ShouldReturn404_WhenNotFound") {
		t.Errorf("expected second failure name, got:\n%s", got)
	}
	if !strings.Contains(got, "Assert.Equal() Failure") {
		t.Errorf("expected error message preserved, got:\n%s", got)
	}

	// Summary
	if !strings.Contains(got, "2 failed") {
		t.Errorf("expected '2 failed' in summary, got:\n%s", got)
	}
	if !strings.Contains(got, "8 passed") {
		t.Errorf("expected '8 passed' in summary, got:\n%s", got)
	}

	// Passing test names should NOT appear
	if strings.Contains(got, "GetUser_ShouldReturnOk") {
		t.Errorf("expected passing test names stripped, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 30.0 {
		t.Errorf("expected >=30%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterDotnetTestEmpty(t *testing.T) {
	got, err := filterDotnetTestCmd("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestFilterDotnetTestWithSkipped(t *testing.T) {
	raw := `Microsoft (R) Test Execution Command Line Tool Version 17.8.0 (x64)
Copyright (c) Microsoft Corporation.  All rights reserved.

Starting test execution, please wait...
A total of 1 test files matched the specified pattern.

Passed!  - Failed:     0, Passed:    15, Skipped:     3, Total:    18, Duration: 2.1 s - MyApp.Tests.dll (net8.0)`

	got, err := filterDotnetTestCmd(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "all 18 tests passed") {
		t.Errorf("expected 'all 18 tests passed', got:\n%s", got)
	}

	t.Logf("output: %s", got)
}
