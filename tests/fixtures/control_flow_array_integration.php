<?php
function bool_word($value) {
    if ($value === true) {
        return 'T';
    }
    if ($value === false) {
        return 'F';
    }
    return 'X';
}

function charset_match($value) {
    return 0 === strcasecmp('UTF-8', $value) || 0 === strcasecmp('UTF8', $value);
}

function select_mode($mode): array {
    if (0 === strcasecmp($mode, 'first')) {
        return ['mode' => 'first'];
    } elseif (0 === strcasecmp($mode, 'second')) {
        return ['mode' => 'second'];
    } else if (0 === strcasecmp($mode, 'third')) {
        return ['mode' => 'third'];
    } else {
        return ['mode' => 'fallback'];
    }
}

function guarded_headers($input): array {
    if (!isset($input['missing']) || charset_match($input['missing'])) {
        return ['guard' => 'missing'];
    } elseif (isset($input['null'])) {
        return ['guard' => 'null-set'];
    } else if (isset($input['charset']) && charset_match($input['charset'])) {
        return ['guard' => 'charset'];
    } else {
        return ['guard' => 'fallback'];
    }
}

function first_decision($items): array {
    foreach ($items as $key => $value) {
        if (is_string($value) && charset_match($value)) {
            return ['match' => $key];
        } elseif (is_array($value)) {
            return ['array' => $key];
        } else if ($value) {
            return ['truthy' => $key];
        } else {
            return ['falsey' => $key];
        }
    }
    return [];
}

function map_charset($value): array {
    if (is_string($value) && charset_match($value)) {
        return ['kind' => 'utf8', 'value' => $value];
    } elseif (is_string($value)) {
        return ['kind' => 'string', 'value' => $value];
    } else if (is_array($value)) {
        return ['kind' => 'array', 'value' => count($value)];
    } else {
        return ['kind' => 'other'];
    }
}

function forward_copy($items): array {
    return $items;
}

$negative = -3;
$tail_key = -2;
$input = ['charset' => 'uTf8', 'null' => null, -3 => 'UTF8', 'nested' => ['child' => 'value'], 'other' => 'latin1'];
$input[] = 'tail';
$selected = select_mode('SECOND');
$guarded = guarded_headers($input);
$charset_guarded = guarded_headers(['missing' => 'latin1', 'null' => null, 'charset' => 'UTF8']);
$first = first_decision($input);
$array_first = first_decision(['nested' => []]);
$mapped = array_map('map_charset', $input);
$returned = forward_copy($input);
$copy = $selected;
$copy['mode'] = 'changed';
$returned['charset'] = 'changed-input-copy';
$returned[] = 'return-tail';

echo 'modes=' . $selected['mode'] . ':' . select_mode('first')['mode'] . ':' . select_mode('third')['mode'] . ':' . select_mode('other')['mode'] . "\n";
echo 'guards=' . $guarded['guard'] . ':' . $charset_guarded['guard'] . ':' . bool_word(isset($input[$negative])) . ':' . bool_word(isset($input['missing'])) . ':' . bool_word(isset($input['null'])) . "\n";
echo 'foreach=' . $first['match'] . ':' . $array_first['array'] . "\n";
echo 'foreach-rest=' . first_decision(['truthy' => 1])['truthy'] . ':' . first_decision(['falsey' => false])['falsey'] . ':' . count(first_decision([])) . "\n";
echo 'mapped=' . $mapped['charset']['kind'] . ':' . $mapped[$negative]['value'] . ':' . $mapped['nested']['kind'] . ':' . $mapped[$tail_key]['kind'] . "\n";
echo 'keys=' . count($input) . ':' . $input[$negative] . ':' . $input[$tail_key] . "\n";
echo 'copies=' . $selected['mode'] . ':' . $copy['mode'] . ':' . $input['charset'] . ':' . $returned['charset'] . "\n";
echo 'returned-append=' . count($input) . ':' . count($returned) . ':' . $returned[-1] . "\n";
