<?php
// Native reference only: not compiler input or a production endpoint.
// Each invocation is an independent CLI execution; never emulate isolation
// with PHP Fibers (they share process-level PCRE error state).
if (PHP_VERSION !== '8.3.36' || PHP_SAPI !== 'cli' || PHP_INT_SIZE !== 8) {
    throw new RuntimeException('requires pinned PHP8.3.36 64-bit CLI');
}
if (setlocale(LC_CTYPE, 'C') !== 'C') {
    throw new RuntimeException('requires C LC_CTYPE');
}
ini_set('pcre.jit', '0');
ini_set('pcre.recursion_limit', '10000');

function typed_value($value) {
    if (is_string($value)) return ['type' => 'string', 'base64' => base64_encode($value)];
    if (is_int($value)) return ['type' => 'int', 'value' => $value];
    if (is_bool($value)) return ['type' => 'bool', 'value' => $value];
    if ($value === null) return ['type' => 'null', 'value' => null];
    $entries = [];
    foreach ($value as $key => $item) {
        $entries[] = ['key' => typed_value($key), 'value' => typed_value($item)];
    }
    return ['type' => 'array', 'entries' => $entries];
}

function observe($label, $pattern, $subject, $limit) {
    $before = preg_last_error();
    ini_set('pcre.backtrack_limit', (string) $limit);
    $matches = null;
    $returned = preg_match($pattern, $subject, $matches, PREG_OFFSET_CAPTURE, 0);
    $after = preg_last_error();
    return ['label' => $label, 'before' => $before, 'after' => $after,
            'return' => typed_value($returned), 'matches' => typed_value($matches)];
}

function nested_observe($label, $pattern, $subject, $limit) {
    return observe($label, $pattern, $subject, $limit);
}

$scenario = $argv[1] ?? '';
$initial = preg_last_error();
if ($scenario === 'a') {
    $long = str_repeat('a', 100) . '!';
    $steps = [
        observe('failure', '/^(a+)+$/s', $long, 1),
        nested_observe('nested_success', '/(?<letter>a)(b)?/s', 'a', 100000),
        observe('failure_again', '/^(a+)+$/s', $long, 1),
        observe('no_match_reset', '/z/s', 'a', 100000),
        nested_observe('nested_failure', '/^(a+)+$/s', $long, 1),
    ];
} elseif ($scenario === 'b') {
    $steps = [observe('empty_capture', '/()/s', '', 100000),
              observe('no_match', '/z/s', 'a', 100000)];
} else {
    throw new RuntimeException('unknown scenario');
}
$final = preg_last_error();
echo json_encode(['schema' => 'gohp.preg83-state-reference/v1',
    'php_version' => PHP_VERSION, 'pcre_version' => PCRE_VERSION,
    'jit' => ini_get('pcre.jit'), 'depth' => ini_get('pcre.recursion_limit'),
    'locale' => setlocale(LC_CTYPE, 0), 'scenario' => $scenario,
    'initial' => $initial, 'final' => $final, 'steps' => $steps], JSON_THROW_ON_ERROR), "\n";
