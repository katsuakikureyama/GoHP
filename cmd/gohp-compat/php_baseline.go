package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/katsuakikureyama/GoHP/internal/compat"
)

// Fixed allowlists only: never collect phpinfo(), environment variables,
// configuration paths or arbitrary extension settings into public artifacts.
func phpBaselineScript() string {
	ini, _ := json.Marshal(compat.PHPINIKeys())
	locales, _ := json.Marshal(compat.PHPLocaleKeys())
	return fmt.Sprintf(`
$ini = [];
foreach (json_decode('%s', true, 512, JSON_THROW_ON_ERROR) as $key) {
    $value = ini_get($key);
    $ini[$key] = $value === false ? null : $value;
}
$locales = [];
foreach (json_decode('%s', true, 512, JSON_THROW_ON_ERROR) as $key) {
    $value = defined($key) ? setlocale(constant($key), 0) : false;
    $locales[$key] = $value === false ? null : $value;
}
$extensions = function ($zend) {
    $names = get_loaded_extensions($zend);
    sort($names, SORT_STRING);
    $result = [];
    foreach ($names as $name) {
        $version = phpversion($name);
        $result[] = ['name' => $name, 'version' => $version === false ? null : $version];
    }
    return $result;
};
echo json_encode([
    'scope' => 'cli-startup-baseline',
    'php_version' => PHP_VERSION,
    'sapi' => PHP_SAPI,
    'int_size' => PHP_INT_SIZE,
    'timezone' => date_default_timezone_get(),
    'ini' => $ini,
    'locales' => $locales,
    'extensions' => $extensions(false),
    'zend_extensions' => $extensions(true)
], JSON_THROW_ON_ERROR);
`, ini, locales)
}

func probePHPBaseline() (*compat.PHPBaseline, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "php", "-r", phpBaselineScript())
	// The suite inherits the same environment. No -n/-d/locale overrides here.
	stdout, stderr := &boundedProbeOutput{}, &boundedProbeOutput{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil || stderr.Len() != 0 {
		// Do not expose arbitrary PHP startup output (potentially secret paths).
		return nil, fmt.Errorf("PHP startup baseline probe failed or emitted diagnostics")
	}
	var baseline compat.PHPBaseline
	decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&baseline); err != nil {
		return nil, fmt.Errorf("invalid PHP startup baseline JSON")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("unexpected trailing PHP startup baseline output")
	}
	if err := baseline.Validate(baseline.PHPVersion); err != nil {
		return nil, err
	}
	return &baseline, nil
}

type boundedProbeOutput struct{ buffer bytes.Buffer }

func (b *boundedProbeOutput) Len() int      { return b.buffer.Len() }
func (b *boundedProbeOutput) Bytes() []byte { return b.buffer.Bytes() }

func (b *boundedProbeOutput) Write(p []byte) (int, error) {
	if len(p) > (128<<10)-b.Len() {
		return 0, fmt.Errorf("PHP startup probe output exceeds limit")
	}
	return b.buffer.Write(p)
}
