package filters

import (
	"strings"
	"testing"
)

func TestGitDiffMultiFile(t *testing.T) {
	raw := `diff --git a/src/app.ts b/src/app.ts
index 1234567..abcdefg 100644
--- a/src/app.ts
+++ b/src/app.ts
@@ -1,5 +1,8 @@
 import express from 'express';
+import cors from 'cors';
+import helmet from 'helmet';

 const app = express();
+app.use(cors());
+app.use(helmet());

-app.listen(3000);
+app.listen(process.env.PORT || 3000);
diff --git a/src/auth/login.ts b/src/auth/login.ts
index 2345678..bcdefgh 100644
--- a/src/auth/login.ts
+++ b/src/auth/login.ts
@@ -10,8 +10,6 @@
 export async function login(email: string, password: string) {
-  const user = await db.query('SELECT * FROM users WHERE email = ?', [email]);
-  if (!user) throw new Error('User not found');
-  const valid = await bcrypt.compare(password, user.password);
+  const user = await findUserByEmail(email);
+  const valid = await verifyPassword(password, user);
   if (!valid) throw new Error('Invalid credentials');
   return generateToken(user);
 }
diff --git a/package.json b/package.json
index 3456789..cdefghi 100644
--- a/package.json
+++ b/package.json
@@ -5,6 +5,8 @@
   "dependencies": {
     "express": "^4.18.0",
+    "cors": "^2.8.5",
+    "helmet": "^7.1.0",
     "bcrypt": "^5.1.0"
   }
 }
`

	got, err := filterGitDiff(raw)
	if err != nil {
		t.Fatal(err)
	}

	// Should show 3 files
	if !strings.Contains(got, "3 files changed") {
		t.Errorf("expected '3 files changed', got: %s", got)
	}

	// Check per-file stats
	if !strings.Contains(got, "src/app.ts:") {
		t.Errorf("expected src/app.ts in output, got: %s", got)
	}
	if !strings.Contains(got, "package.json:") {
		t.Errorf("expected package.json in output, got: %s", got)
	}

	// Verify +/- counts appear
	lines := strings.Split(strings.TrimSpace(got), "\n")
	lastLine := lines[len(lines)-1]
	if !strings.HasPrefix(lastLine, "3 files changed") {
		t.Errorf("expected summary line at end, got: %s", lastLine)
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - (float64(filteredTokens)/float64(rawTokens))*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
}

func TestGitDiffShortPassthrough(t *testing.T) {
	raw := `diff --git a/README.md b/README.md
index 1234567..abcdefg 100644
--- a/README.md
+++ b/README.md
@@ -1,3 +1,3 @@
-# Old Title
+# New Title

 Some content.`

	got, err := filterGitDiff(raw)
	if err != nil {
		t.Fatal(err)
	}

	// Short diff (<10 lines) should pass through
	if got != strings.TrimSpace(raw) {
		t.Errorf("short diff should pass through, got: %s", got)
	}
}

func TestGitDiffEmpty(t *testing.T) {
	got, err := filterGitDiff("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty output for empty input, got: %s", got)
	}
}
