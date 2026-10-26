<?php
// Differential coverage for pure scalar built-ins on the scalar-frame path.
// The legacy helpers below deliberately keep an array or an untyped value.

function scalar_trace(int $limit, string $text, bool $trim) {
    $copy = $text;
    $text = 'local';
    $lower = StRtOlOwEr($copy);
    $piece = substr(strtolower($copy), 0, 3);
    $index = 0;
    $marks = '';
    while ($index !== $limit) {
        $marks = $marks . $index;
        $index = $index + 1;
    }
    if ($trim) {
        return rtrim($piece . '///', '/') . ':' . strlen($copy) . ':' . $marks;
    }
    return $piece . ':' . strcasecmp($lower, $copy) . ':' . $marks;
}

function early_builtin(int $offset, bool $early) {
    if ($early) {
        return substr('abc', $offset, 1);
    }
    return rtrim('stay///', '/');
}

function scalar_predicates() {
    $empty = '';
    $zero = 0;
    $false_value = false;
    $nothing = null;
    return (is_array($empty) ? 'A' : 'a') . ':' . (is_bool($false_value) ? 'B' : 'b') . ':'
        . (is_int($zero) ? 'I' : 'i') . ':' . (is_null($nothing) ? 'N' : 'n') . ':'
        . (is_string($empty) ? 'S' : 's') . ':' . (is_object($nothing) ? 'O' : 'o');
}

function standalone_builtin() {
    StRtOlOwEr('IGNORED');
    return str_starts_with('prefix', 'pre') && str_contains(strtolower('PREFIX'), 'fix') ? 'start' : 'wrong';
}

function lazy_scalar(string $text, bool $run) {
    $value = $run ? $text : false;
    return $run ? strlen($value) : 'skipped';
}

function lazy_legacy($value, bool $run) {
    if ($run) {
        return strlen($value);
    }
    return 'skipped';
}

function legacy_array_probe(array $items) {
    $copy = $items;
    $copy['local'] = 'copy';
    return (is_array($copy) ? 'array' : 'wrong') . ':' . (is_bool($copy['flag']) ? 'bool' : 'wrong') . ':'
        . (is_int($copy['zero']) ? 'int' : 'wrong') . ':' . (is_null($copy['none']) ? 'null' : 'wrong') . ':'
        . (is_string($copy['text']) ? 'string' : 'wrong') . ':' . (is_object($copy['text']) ? 'object' : 'plain') . ':'
        . (isset($items['local']) ? 'changed' : 'clean');
}

$source = ['flag' => false, 'zero' => 0, 'none' => null, 'text' => 'text'];

echo 'trace=' . scalar_trace(3, 'AbCd', false) . ':' . scalar_trace(2, '雪', true) . "\n";
echo 'early=' . early_builtin(1, true) . ':' . early_builtin(0, false) . "\n";
echo 'predicates=' . scalar_predicates() . "\n";
echo 'standalone=' . standalone_builtin() . "\n";
echo 'lazy=' . lazy_legacy($source, false) . "\n";
echo 'lazy-owned=' . lazy_scalar('abc', true) . ':' . lazy_scalar('', false) . "\n";
echo 'legacy=' . legacy_array_probe($source) . ':' . (isset($source['local']) ? 'changed' : 'clean') . "\n";
