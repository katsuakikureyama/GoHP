<?php
function describe_value($value) {
    if (is_array($value)) {
        $description = 'array:' . count($value) . '[';
        foreach ($value as $key => $child) {
            $description = $description . $key . '=' . describe_value($child) . ',';
        }
        return $description . ']';
    }
    if ($value === null) {
        return 'null';
    }
    if ($value === null) {
        return 'null';
    }
    if ($value === false) {
        return 'false';
    }
    if ($value === 0) {
        return 'zero';
    }
    if ($value === '') {
        return 'empty';
    }
    if ($value === '0') {
        return 'string-zero';
    }
    if (is_string($value)) {
        return 'string:' . $value;
    }
    return 'int:' . $value;
}

function dump_entries($label, $items) {
    foreach ($items as $key => $value) {
        echo $label . ':' . (is_string($key) ? 'string' : 'int') . ':' . $key . '=' . describe_value($value) . ',';
    }
    echo "\n";
    return '';
}

function select_set_entries($input, $keys) {
    $slice = [];
    foreach ($keys as $key) {
        if (isset($input[$key])) {
            $slice[$key] = $input[$key];
        }
    }
    return $slice;
}

function bool_word($value) {
    return $value ? 'true' : 'false';
}

function guarded_probe($input, $key) {
    return false && isset($input[$key]);
}

$input = [
    'null' => null,
    'false' => false,
    'zero' => 0,
    'empty' => '',
    'string-zero' => '0',
    'empty-array' => [],
    'nested' => ['child'],
    'text' => 'plain',
    'seven' => 7,
    2 => 'two',
    '02' => 'leading-zero',
];
$input['-5'] = 'negative';
$keys = ['text', 'missing', 'null', 'false', 'zero', 'empty', 'string-zero', 'empty-array', 'nested', 'seven', '2', '02', '-5', 'text', 'missing'];
$selected = select_set_entries($input, $keys);

echo 'direct:' . bool_word(isset($input['missing'])) . ':' . bool_word(isset($input['null'])) . ':' . bool_word(IsSeT($input['false'])) . ':' . bool_word(isset($input['zero'])) . ':' . bool_word(isset($input['empty'])) . ':' . bool_word(isset($input['string-zero'])) . ':' . bool_word(isset($input['empty-array'])) . "\n";
echo 'guarded:' . bool_word(guarded_probe([], true)) . "\n";
echo dump_entries('selected', $selected);
echo dump_entries('input-after', $input);
