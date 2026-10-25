<?php
// Owned cross-stack fixture, not a replacement for a WordPress original.
function charset_match($value) {
    if (!is_string($value)) {
        return false;
    }
    return 0 === strcasecmp('UTF-8', $value) || 0 === strcasecmp('UTF8', $value);
}

function shallow_unescape($value) {
    return is_string($value) ? stripslashes($value) : $value;
}

function select_charset_entries($input, $keys): array {
    $selected = [];
    foreach ($keys as $key) {
        if (isset($input[$key]) && (is_string($input[$key]) || is_array($input[$key]))) {
            $selected[$key] = $input[$key];
        }
    }
    return array_map('charset_match', array_map('shallow_unescape', $selected));
}

function forward_selected($input, $keys): array {
    echo 'forward-once' . "\n";
    return select_charset_entries($input, $keys);
}

function dump_flags($label, $items) {
    foreach ($items as $key => $value) {
        echo $label . ':' . (is_string($key) ? 's' : 'i') . ':' . $key . '=' . ($value === true ? 'T' : 'F') . "\n";
    }
    return '';
}

$input = ['compact' => 'uTf8', 'escaped' => 'UTF\\8', 'other' => 'latin1',
    'null' => null, 'false' => false, 'nested' => ['UTF-8'], -2 => 'UTF-8'];
$keys = ['missing', 'null', 'false', 'nested', -2, 'escaped', 'other', 'compact', 'escaped'];
$result = forward_selected($input, $keys);
$copy = $result;
$copy['compact'] = false;
echo dump_flags('result', $result);
echo 'copy:' . ($result['compact'] === true ? 'original-true' : 'changed') . ':' . ($copy['compact'] === false ? 'copy-false' : 'wrong') . "\n";
echo 'input:' . $input['escaped'] . ':' . count($input) . ':' . count($keys) . "\n";
echo 'empty:' . count(select_charset_entries($input, ['missing', 'null', 'false'])) . "\n";
echo 'guards:' . ((!isset($input['missing']) || charset_match($input['missing'])) ? 'skip' : 'wrong') . ':' . (isset($input['null']) === false ? 'null-false' : 'wrong') . "\n";
