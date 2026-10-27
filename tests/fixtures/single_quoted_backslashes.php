<?php
// Owned differential fixture: PHP single-quoted strings only interpret \\ and
// \'. Every other backslash pair remains two literal bytes. This deliberately
// uses byte-oriented checks so the fixture does not depend on a regexp engine.

function bool_word($value) {
    if ($value === true) {
        return 'T';
    }
    return 'F';
}

function relay_string($value) {
    return $value;
}

$slash = '\\';
$quote = 'a\'b';
$unknown_s = '\s';
$unknown_slash = '\/';
$unknown_upper = '\F';
$unknown_vertical = '\v';
$unknown_escape = '\e';
$unknown_hex = '\x41';
$unknown_octal = '\101';
$unknown_unicode_escape = '\u{41}';
$unknown_dollar = '\$';

// Three source backslashes before s become two literal backslashes before s;
// a final escaped source backslash remains one byte in the value.
$parity = '\\\s';
$trailing = 'tail\\';

// A physical newline after a single-quoted backslash is not a line
// continuation. The UTF-8 character also remains adjacent to a literal slash.
$physical_newline = 'before\
after';
$unicode_adjacent = '\é';

echo 'known=' . bool_word($slash === "\\") . ':'
    . bool_word($quote === "a'b") . ':'
    . bool_word($parity === $slash . $slash . 's') . ':'
    . bool_word($trailing === 'tail' . $slash) . "\n";

echo 'unknown=' . bool_word($unknown_s === $slash . 's') . ':'
    . bool_word($unknown_slash === $slash . '/') . ':'
    . bool_word($unknown_upper === $slash . 'F') . ':'
    . bool_word($unknown_vertical === $slash . 'v') . ':'
    . bool_word($unknown_escape === $slash . 'e') . ':'
    . bool_word($unknown_hex === $slash . 'x41') . ':'
    . bool_word($unknown_octal === $slash . '101') . ':'
    . bool_word($unknown_unicode_escape === $slash . 'u{41}') . ':'
    . bool_word($unknown_dollar === $slash . '$') . "\n";

echo 'adjacent=' . bool_word($physical_newline === 'before' . $slash . "\n" . 'after') . ':'
    . strlen($physical_newline) . ':'
    . bool_word($unicode_adjacent === $slash . 'é') . ':'
    . strlen($unicode_adjacent) . "\n";

// Static literal keys exercise the decoded byte values through variable index
// reads, while relay_string covers a function argument and return path.
$by_escape = [
    '\\s' => 's',
    '\\/' => 'slash',
    '\\F' => 'upper',
    '\\v' => 'vertical',
    '\\e' => 'escape',
    '\\x41' => 'hex',
    '\\101' => 'octal',
    '\\u{41}' => 'unicode',
    '\\$' => 'dollar',
];
$relayed = relay_string($unknown_unicode_escape);
echo 'flow=' . $by_escape[$unknown_s] . ':' . $by_escape[$unknown_slash] . ':'
    . $by_escape[$unknown_hex] . ':' . $by_escape[$relayed] . ':'
    . $by_escape[$unknown_dollar] . ':'
    . bool_word($relayed === $unknown_unicode_escape) . "\n";
