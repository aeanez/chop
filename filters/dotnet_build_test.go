package filters

import (
	"strings"
	"testing"
)

func TestFilterDotnetBuildSuccess(t *testing.T) {
	raw := `Microsoft (R) Build Engine version 17.8.3+195e7f5a3 for .NET
Copyright (C) Microsoft Corporation. All rights reserved.

  Determining projects to restore...
  All projects are up-to-date for restore.
  MyApp.Core -> C:\src\MyApp.Core\bin\Debug\net8.0\MyApp.Core.dll
  MyApp.Web -> C:\src\MyApp.Web\bin\Debug\net8.0\MyApp.Web.dll
  MyApp.Tests -> C:\src\MyApp.Tests\bin\Debug\net8.0\MyApp.Tests.dll

Build succeeded.

    0 Warning(s)
    0 Error(s)

Time Elapsed 00:00:04.52`

	got, err := filterDotnetBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Build succeeded") {
		t.Errorf("expected 'Build succeeded', got:\n%s", got)
	}

	// Token savings
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterDotnetBuildWithErrors(t *testing.T) {
	raw := `Microsoft (R) Build Engine version 17.8.3+195e7f5a3 for .NET
Copyright (C) Microsoft Corporation. All rights reserved.

  Determining projects to restore...
  All projects are up-to-date for restore.
  MyApp.Core -> C:\src\MyApp.Core\bin\Debug\net8.0\MyApp.Core.dll
  MyApp.Data -> C:\src\MyApp.Data\bin\Debug\net8.0\MyApp.Data.dll
  MyApp.Services -> C:\src\MyApp.Services\bin\Debug\net8.0\MyApp.Services.dll
  Restored C:\src\MyApp.Web\MyApp.Web.csproj (in 1.2 sec).
  Restored C:\src\MyApp.Core\MyApp.Core.csproj (in 0.3 sec).
  Restored C:\src\MyApp.Data\MyApp.Data.csproj (in 0.5 sec).
  Restored C:\src\MyApp.Services\MyApp.Services.csproj (in 0.4 sec).
  NuGet Config files used:
      C:\Users\dev\AppData\Roaming\NuGet\NuGet.Config
      C:\src\NuGet.Config
  Assets file has not changed. Skipping assets file writing. Path: C:\src\MyApp.Core\obj\project.assets.json
  Assets file has not changed. Skipping assets file writing. Path: C:\src\MyApp.Data\obj\project.assets.json
  3 project(s) in solution are up-to-date for restore.
  Generating code for MyApp.Data...
  Generating code for MyApp.Services...
Program.cs(12,5): error CS1002: ; expected [C:\src\MyApp.Web\MyApp.Web.csproj]
Program.cs(24,10): error CS0246: The type or namespace name 'Foo' could not be found [C:\src\MyApp.Web\MyApp.Web.csproj]
Startup.cs(8,1): warning CS0168: The variable 'ex' is declared but never used [C:\src\MyApp.Web\MyApp.Web.csproj]

Build FAILED.

    1 Warning(s)
    2 Error(s)

Time Elapsed 00:00:03.18`

	got, err := filterDotnetBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Build FAILED") {
		t.Errorf("expected 'Build FAILED', got:\n%s", got)
	}

	// Errors preserved
	if !strings.Contains(got, "CS1002") {
		t.Errorf("expected error CS1002 preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "CS0246") {
		t.Errorf("expected error CS0246 preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "Program.cs(12)") {
		t.Errorf("expected file/line reference preserved, got:\n%s", got)
	}

	// Warning preserved
	if !strings.Contains(got, "CS0168") {
		t.Errorf("expected warning CS0168 preserved, got:\n%s", got)
	}

	// Token savings
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterDotnetBuildEmpty(t *testing.T) {
	got, err := filterDotnetBuild("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestFilterDotnetBuildWarningsOnly(t *testing.T) {
	raw := `Microsoft (R) Build Engine version 17.8.3+195e7f5a3 for .NET
Copyright (C) Microsoft Corporation. All rights reserved.

  Determining projects to restore...
  All projects are up-to-date for restore.
  Restored C:\src\MyApp\MyApp.csproj (in 312 ms).
  Restored C:\src\MyApp.Core\MyApp.Core.csproj (in 0.2 sec).
  Restored C:\src\MyApp.Data\MyApp.Data.csproj (in 0.4 sec).
  NuGet Config files used:
      C:\Users\dev\AppData\Roaming\NuGet\NuGet.Config
  Assets file has not changed. Skipping assets file writing. Path: C:\src\MyApp\obj\project.assets.json
  Assets file has not changed. Skipping assets file writing. Path: C:\src\MyApp.Core\obj\project.assets.json
  3 project(s) in solution are up-to-date for restore.
  Generating code for MyApp.Data...
  MyApp.Core -> C:\src\MyApp.Core\bin\Debug\net8.0\MyApp.Core.dll
  MyApp.Data -> C:\src\MyApp.Data\bin\Debug\net8.0\MyApp.Data.dll
  MyApp -> C:\src\MyApp\bin\Debug\net8.0\MyApp.dll
Controllers\HomeController.cs(15,9): warning CS0219: The variable 'unused' is assigned but its value is never used [C:\src\MyApp\MyApp.csproj]
Models\User.cs(42,5): warning CS0414: The field 'User._legacy' is assigned but its value is never used [C:\src\MyApp\MyApp.csproj]

Build succeeded.

    2 Warning(s)
    0 Error(s)

Time Elapsed 00:00:02.87`

	got, err := filterDotnetBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Build succeeded") {
		t.Errorf("expected 'Build succeeded', got:\n%s", got)
	}
	if !strings.Contains(got, "CS0219") {
		t.Errorf("expected warning CS0219, got:\n%s", got)
	}
	if !strings.Contains(got, "CS0414") {
		t.Errorf("expected warning CS0414, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}
