<?php
// Owned differential fixture: empty($variable) admits only values that are
// strings or arrays. PHP strings are empty only when they are exactly '' or
// '0'; the other values below must remain non-empty byte sequences.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function typed_string_empty(string $value): bool {
    return empty($value);
}

function untyped_empty($value): bool {
    return empty($value);
}

function marker_empty(array $items): bool {
    echo '<marker>';
    return empty($items);
}

$strings = ['', '0', '00', '0.0', '-0', ' ', 'false', '日本語'];
$string_marks = '';
foreach ($strings as $text) {
    $string_marks = $string_marks . bool_word(empty($text));
}

$input = '0';
$empty_array = [];
$arrow_empty = fn($s) => empty($s);
$typed = typed_string_empty($input);
$untyped_string = untyped_empty($input);
$untyped_array = untyped_empty($empty_array);
$arrow = $arrow_empty($input);

// This is the original parser shape: a derived string is assigned to a bare
// local before empty() decides whether there is remaining markup.
$html = substr('body', 4);
$html_empty = empty($html);
$html = substr('body', 1);
$html_nonempty = empty($html);
$html_branch = '';
if (!empty($html)) {
    $html_branch = 'content';
} else {
    $html_branch = 'empty';
}

$changed = '';
$changed_first = empty($changed);
$changed = ['value'];
$changed_second = empty($changed);
$changed = '0';
$changed_third = empty($changed);

$if_widened = '';
if (empty($if_widened)) {
    $if_widened = ['value'];
} else {
    $if_widened = '0';
}
$if_mark = empty($if_widened);

$loop_widened = '';
$loop_marks = '';
foreach (['0', ['value'], ''] as $next) {
    $loop_widened = $next;
    $loop_marks = $loop_marks . bool_word(empty($loop_widened));
}

$null_item = [null];
$false_item = [false];
$zero_item = ['0'];
$nested_empty_item = [[]];

// marker_empty must not run: both short-circuit operands are valid domains,
// but the right-hand calls are unreachable under their respective operators.
$short_or = empty($empty_array) || marker_empty(['value']);
$short_and = !empty($input) && marker_empty($empty_array);

echo 'strings=' . $string_marks . "\n";
echo 'helpers=' . bool_word($typed) . ':' . bool_word($untyped_string) . ':' . bool_word($untyped_array) . ':' . bool_word($arrow) . ':input=' . $input . "\n";
echo 'html=' . bool_word($html_empty) . ':' . bool_word($html_nonempty) . ':' . $html_branch . "\n";
echo 'reassign=' . bool_word($changed_first) . ':' . bool_word($changed_second) . ':' . bool_word($changed_third) . "\n";
echo 'widened=' . bool_word($if_mark) . ':' . $loop_marks . "\n";
echo 'arrays=' . bool_word(empty($null_item)) . ':' . bool_word(empty($false_item)) . ':' . bool_word(empty($zero_item)) . ':' . bool_word(empty($nested_empty_item)) . "\n";
echo 'short=' . bool_word($short_or) . ':' . bool_word($short_and) . "\n";
