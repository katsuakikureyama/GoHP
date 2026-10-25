<?php
// Owned composition fixture: not a WordPress original or a Gossip pin upgrade.
function classify($value): array {
    if (in_array($value, ['image/heic', 'image/heif'], true)) {
        return ['kind' => 'image', 'value' => $value];
    } elseif (in_array($value, [0, false, null], true)) {
        return ['kind' => 'sentinel', 'value' => $value];
    } else if (in_array($value, ['0', ''], true)) {
        return ['kind' => 'text', 'value' => $value];
    } else {
        return ['kind' => 'other', 'value' => $value];
    }
}

function select_members($input, $allowed): array {
    $result = [];
    foreach ($input as $key => $value) {
        if (!isset($input[$key])) {
            $result[$key] = 'null';
        } elseif (in_array($value, $allowed, true)) {
            $result[$key] = $value;
        } else {
            $result[$key] = 'miss';
        }
    }
    return $result;
}

function observe($label, $value) {
    echo $label . "\n";
    return $value;
}

function short_circuit($skip, $haystack): array {
    if ($skip || in_array('x', $haystack, true)) {
        return ['kind' => 'first'];
    } elseif (in_array('later', $haystack, true)) {
        return ['kind' => 'later'];
    } else {
        return ['kind' => 'fallback'];
    }
}

$input = [-3 => 'image/heic', 'int' => 0, 'false' => false, 'null' => null, 'text' => '0', 'empty' => '', 'upper' => 'IMAGE/HEIC'];
$mapped = array_map('classify', $input);
foreach ($mapped as $key => $entry) {
    echo $key . ':' . $entry['kind'] . ':' . ($entry['value'] === $input[$key] ? 'same' : 'changed') . "\n";
}
$selected = select_members($input, ['image/heic', 0, '0']);
$copy = $selected;
$copy['int'] = 'changed';
$selected[] = 'tail';
echo 'copy:' . $selected['int'] . ':' . $copy['int'] . ':' . $input['int'] . ':' . $selected[-2] . ':' . count($input) . "\n";
echo 'selected:' . $selected['false'] . ':' . $selected['null'] . ':' . $selected['text'] . ':' . $selected['empty'] . "\n";
echo 'case:' . (0 === strcasecmp('IMAGE/HEIC', $input['upper']) ? 'folded' : 'wrong') . ':' . (in_array($input['upper'], ['image/heic'], true) ? 'wrong' : 'strict') . "\n";
echo 'skip:' . short_circuit(true, [['outside-domain']])['kind'] . "\n";
echo 'later:' . short_circuit(false, ['later'])['kind'] . ':' . short_circuit(false, [])['kind'] . "\n";
if (in_array(observe('needle-once', 'x'), observe('haystack-once', ['x']), true)) {
    echo 'matched' . "\n";
} elseif (in_array(observe('must-not-run', 'later'), [], true)) {
    echo 'wrong' . "\n";
}
