<?php
// Owned differential fixture: bounded empty($arrayVariable). The compiler
// accepts only a variable operand; each non-empty case below is an array whose
// contents deliberately must not affect the result.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function dynamic_value($value) {
    return $value;
}

function empty_dynamic($value): bool {
    return EmPtY($value);
}

function empty_typed(array $items): bool {
    return empty($items);
}

function first_or_null(array $items) {
    if (empty($items)) {
        return null;
    }
    return $items[0];
}

function forward_empty(array $items, callable $predicate): bool {
    return $predicate($items);
}

function empty_flag($items): bool {
    return empty($items);
}

function local_write_flag(array $items): bool {
    $was_empty = empty($items);
    $items[] = 'local';
    return $was_empty && !empty($items);
}

function marker_array(array $items): bool {
    echo '<marker>';
    return empty($items);
}

class EmptyProbe {
    public $value = 'object';
}

$empty = [];
$null_item = [null];
$false_item = [false];
$zero_item = [0];
$zero_text_item = ['0'];
$blank_item = [''];
$array_item = [[]];
$object_item = [new EmptyProbe()];
$arrow_item = [fn($value) => $value];
$nested_item = [[false, 0, '0', '', []]];

$dynamic_empty = empty_dynamic(dynamic_value($empty));
$dynamic_one = empty_dynamic(dynamic_value($null_item));
$typed_empty = empty_typed($empty);
$typed_one = empty_typed($false_item);
$first_empty = first_or_null($empty);
$first_one = first_or_null(['first']);
$forward_empty = forward_empty($empty, fn($items) => empty($items));
$forward_one = forward_empty($zero_item, fn($items) => empty($items));
$mapped = array_map('empty_flag', [[], [null], [false], [0]]);
$local_write = local_write_flag($empty);
$lazy_empty = empty($empty) || marker_array($empty);
$lazy_one = empty($null_item) || marker_array($null_item);

echo 'arrays=' . bool_word(EMPTY($empty)) . ':' . bool_word(empty($null_item)) . ':' . bool_word(empty($false_item)) . ':' . bool_word(empty($zero_item)) . ':' . bool_word(empty($zero_text_item)) . ':' . bool_word(empty($blank_item)) . ':' . bool_word(empty($array_item)) . ':' . bool_word(empty($object_item)) . ':' . bool_word(empty($arrow_item)) . ':' . bool_word(empty($nested_item)) . "\n";
echo 'params=' . bool_word($dynamic_empty) . ':' . bool_word($dynamic_one) . ':' . bool_word($typed_empty) . ':' . bool_word($typed_one) . "\n";
echo 'first=' . bool_word(is_null($first_empty)) . ':' . $first_one . "\n";
echo 'callable=' . bool_word($forward_empty) . ':' . bool_word($forward_one) . "\n";
echo 'mapped=' . bool_word($mapped[0]) . ':' . bool_word($mapped[1]) . ':' . bool_word($mapped[2]) . ':' . bool_word($mapped[3]) . "\n";
echo 'copy=' . bool_word($local_write) . ':' . bool_word(empty($empty)) . ':' . count($empty) . "\n";
echo 'lazy=' . bool_word($lazy_empty) . ':' . bool_word($lazy_one) . "\n";
