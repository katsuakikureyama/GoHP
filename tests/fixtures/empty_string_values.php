<?php
// Owned differential fixture for the deliberately small empty($definedBareLocal)
// extension: actual strings use PHP's exact ""/"0" result, while actual arrays
// keep their existing outer-entry-count behavior.  It does not exercise
// undefined addresses, properties, indexes, calls, or scalar coercion.

function bool_word($value) {
    return $value ? 'T' : 'F';
}

function empty_string(string $value): bool {
    return empty($value);
}

function empty_mixed($value): bool {
    return empty($value);
}

function forward_empty($value, callable $callback): bool {
    return $callback($value);
}

// This mirrors the measured parser shape: a typed HTML string is narrowed by
// substr before the same bare-local empty construct observes it.
function empty_substring(string $html, int $offset, int $length): bool {
    $html = substr($html, $offset, $length);
    return empty($html);
}

function nested_empty(string $value, bool $outer): bool {
    if ($outer) {
        if (empty($value)) {
            return true;
        }
    }
    return false;
}

function marker_empty($value): bool {
    echo '<marker>';
    return empty($value);
}

$blank = '';
$zero = '0';
$double_zero = '00';
$decimal_zero = '0.0';
$leading_zero = ' 0';
$trailing_zero = '0 ';
$space = ' ';
$negative_zero = '-0';
$utf8 = '雪';
// PHP single quotes retain \0; stripslashes then materializes one NUL byte.
$nul = stripslashes('a\0b');
$array_empty = [];
$array_one = [null];

$forward_blank = forward_empty($blank, fn($value) => empty($value));
$forward_one = forward_empty($array_one, fn($value) => empty($value));
$lazy_skip = empty($blank) || marker_empty($array_one);

echo 'strings=' . bool_word(empty_string($blank)) . ':' . bool_word(empty_string($zero)) . ':'
    . bool_word(empty_string($double_zero)) . ':' . bool_word(empty_string($decimal_zero)) . ':'
    . bool_word(empty_string($leading_zero)) . ':' . bool_word(empty_string($trailing_zero)) . ':'
    . bool_word(empty_string($space)) . ':' . bool_word(empty_string($negative_zero)) . ':'
    . bool_word(empty_string($utf8)) . ':' . bool_word(empty_string($nul)) . "\n";
echo 'mixed=' . bool_word(empty_mixed($array_empty)) . ':' . bool_word(empty_mixed($array_one)) . ':'
    . bool_word(empty_mixed($blank)) . ':' . bool_word(empty_mixed($zero)) . ':'
    . bool_word(empty_mixed($double_zero)) . "\n";
echo 'forward=' . bool_word($forward_blank) . ':' . bool_word($forward_one) . "\n";
echo 'substr=' . bool_word(empty_substring('x', 0, 0)) . ':'
    . bool_word(empty_substring('x0y', 1, 1)) . ':'
    . bool_word(empty_substring('x00y', 1, 2)) . "\n";
echo 'nested=' . bool_word(nested_empty('', true)) . ':' . bool_word(nested_empty('', false)) . "\n";
echo 'lazy=' . bool_word($lazy_skip) . ':';
$lazy_run = empty($array_one) || marker_empty($blank);
echo bool_word($lazy_run) . "\n";
echo 'preserve=' . bool_word(empty($zero)) . ':' . strlen($zero) . ':'
    . bool_word(empty($array_one)) . ':' . count($array_one) . "\n";
