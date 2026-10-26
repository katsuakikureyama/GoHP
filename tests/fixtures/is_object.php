<?php
// is_object recognizes only generated PHP-object markers: generated classes
// (including empty classes) and capture-free generated arrows, never host-like
// values or arrays/scalars.
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

function marked_value($value) {
    echo '<marked>';
    return $value;
}

function object_flag($value): bool {
    return Is_ObJeCt($value);
}

function object_values(array $items): array {
    $objects = [];
    foreach ($items as $key => $value) {
        if (is_object($value)) {
            $objects[] = $value;
        }
    }
    return $objects;
}

function name_after_object_guard($value) {
    if (is_object($value) && $value->name === 'box') {
        return $value->name;
    }
    return 'other';
}

class Box {
    public $name = 'box';
}

class EmptyBox {}

$box = new Box();
$empty = new EmptyBox();
$arrow = fn($value) => is_object($value);
$items = ['box' => $box, 'empty' => $empty, 'arrow' => $arrow, 'array' => [], 'text' => 'text'];
$values = array_values($items);
$objects = object_values($items);
$mapped = array_map('object_flag', $items);
$dynamic_box = object_flag(dynamic_value($box));
$dynamic_text = object_flag(dynamic_value('text'));
$marker = object_flag(marked_value($empty));
$arrow_box = $arrow($box);
$arrow_array = $arrow([]);
$guard_box = name_after_object_guard($box);
$guard_array = name_after_object_guard([]);

echo 'direct=' . bool_word(is_object($box)) . ':' . bool_word(is_object($empty)) . ':' . bool_word(is_object($arrow)) . ':' . bool_word(is_object([])) . ':' . bool_word(is_object('text')) . ':' . bool_word(is_object(null)) . "\n";
echo 'dynamic=' . bool_word($dynamic_box) . ':' . bool_word($dynamic_text) . ':' . bool_word($marker) . "\n";
echo 'mapped=' . bool_word($mapped['box']) . ':' . bool_word($mapped['empty']) . ':' . bool_word($mapped['arrow']) . ':' . bool_word($mapped['array']) . ':' . bool_word($mapped['text']) . "\n";
echo 'values=' . bool_word(is_object($values[0])) . ':' . bool_word(is_object($values[1])) . ':' . bool_word(is_object($values[2])) . ':' . bool_word(is_object($values[3])) . ':' . bool_word(is_object($values[4])) . "\n";
echo 'objects=' . count($objects) . ':' . bool_word(is_object($objects[0])) . ':' . bool_word(is_object($objects[1])) . ':' . bool_word(is_object($objects[2])) . "\n";
echo 'arrow=' . bool_word($arrow_box) . ':' . bool_word($arrow_array) . "\n";
echo 'guard=' . $guard_box . ':' . $guard_array . "\n";
echo 'original=' . $items['box']->name . ':' . $box->name . "\n";
