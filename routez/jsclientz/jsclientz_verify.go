package jsclientz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/stretchr/testify/require"
)

// RunJsFetch executa o código JS gerado com o Node contra um servidor real, e devolve o
// resultado — é como um teste confirma que o SDK gerado realmente funciona, não só compila.
func RunJsFetch(t testing.TB, fnName, code, baseUrl string, params map[string]any) map[string]any {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	call, err := jsCallExpr(fnName, fmt.Sprintf("%q", baseUrl), params)
	require.NoError(t, err)

	driver := fmt.Sprintf("%s\n%s.then(r => console.log(JSON.stringify(r))).catch(e => { console.error(e); process.exit(1); });\n",
		code, call)
	scriptPath := filepath.Join(t.TempDir(), "client.mjs")
	require.NoError(t, filez.WriteFile(scriptPath, []byte(driver)))

	out, err := exec.Command("node", scriptPath).CombinedOutput()
	require.NoError(t, err, string(out))

	var result map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(out), &result))
	return result
}
