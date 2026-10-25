<?php
function empty_result(): array {
    return [];
}

function associative_result(): ArRaY {
    return ['name' => 'value', 2 => false, 'empty' => '', 'null' => null];
}

function nested_result(): array {
    return ['nested' => ['child' => 'value'], 'falsy' => [false, 0, '', null]];
}

function guarded_result($items): array {
    foreach ($items as $key => $value) {
        if ($key === 'take') {
            $selected = [];
            $selected[$key] = $value;
            return $selected;
        }
    }
    return [];
}

function forward_result($value): array {
    return marker_result($value);
}

function marker_result($value): array {
    echo 'marker:' . $value . "\n";
    return ['value' => $value];
}

function copy_result($items): array {
    return $items;
}

$empty = empty_result();
$assoc = associative_result();
$nested = nested_result();
$selected = guarded_result(['skip' => 'no', 'take' => 'yes']);
$missing = guarded_result(['skip' => 'no']);
$forward = forward_result('once');
$original = ['value' => 'original'];
$copy = copy_result($original);
$copy['value'] = 'changed';

echo 'empty:' . count($empty) . "\n";
echo 'assoc:' . $assoc['name'] . ':' . ($assoc[2] ? 'true' : 'false') . ':' . $assoc['empty'] . ':' . ($assoc['null'] === null ? 'null' : 'not-null') . "\n";
echo 'nested:' . $nested['nested']['child'] . ':' . $nested['falsy'][1] . "\n";
echo 'guarded:' . $selected['take'] . ':' . count($missing) . "\n";
echo 'forward:' . $forward['value'] . "\n";
echo 'copy:' . $original['value'] . ':' . $copy['value'] . "\n";
