<?php
// Original composition fixture: arrays select bounded string operations.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function format_values($input, $width) {
    $result = '';
    foreach (array_values($input) as $value) {
        if (is_string($value) && strlen(rtrim($value, '/')) === 4) {
            $result = $result . sprintf('%0' . $width . 's', strtolower(rtrim($value, '/'))) . ';';
        }
    }
    return $result;
}

function null_guard($input) {
    if (array_key_exists('none', $input) && is_string($input['none'])) {
        return rtrim($input['none'], '/');
    }
    return 'guarded';
}

$input = ArRaY(9 => 'WORD///', 'none' => null, 'bird' => 'BIRD/', 'number' => 42, 'last' => 'LAST/');
$width = 6;
$values = array_values($input);
$keys = array_keys($input);

echo 'formatted=' . format_values($input, $width) . "\n";
echo 'null=' . bool_word(array_key_exists('none', $input)) . ':' . null_guard($input) . "\n";
echo 'lists=' . $keys[0] . ':' . $keys[1] . ':' . bool_word(array_is_list($values)) . ':' . bool_word(array_is_list($keys)) . "\n";
echo 'format=' . '[' . sprintf('%00s', 'x') . ']' . ':' . '[' . sprintf('%04s', 'é') . ']' . ':' . '[' . sprintf('%0' . $width . 's', 42) . ']' . "\n";
echo 'replace=' . str_replace('', 'X', 'abc') . ':' . str_replace('a', 'aa', strtolower(rtrim('A///', '/'))) . "\n";
echo 'original=' . $input[9] . ':' . $input['bird'] . ':' . $input['last'] . "\n";
