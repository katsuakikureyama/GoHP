<?php
// Owned differential fixture: is_null() is an actual-null predicate only.
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

function forwarded_null($value): bool {
    return is_null($value);
}

function null_or_safe_add($value): bool {
    return is_null($value) || $value + 1 === 1;
}

function null_ternary_safe_add($value): bool {
    return is_null($value) ? true : $value + 1 === 1;
}

function any_null(array $items, callable $callback): bool {
    foreach ($items as $key => $value) {
        if ($callback($value, $key)) {
            return true;
        }
    }
    return false;
}

function null_flag($value): bool {
    return is_null($value);
}

class NullProbe {
    public $value = 'object';
}

$empty = [];
$object = new NullProbe();
$arrow = fn($value) => $value;
$values = ['null' => null, 'false' => false, 'one' => 1];
$dynamic_null = forwarded_null(dynamic_value(null));
$dynamic_text = forwarded_null(dynamic_value('null'));
$marker = is_null(marked_value(null));
$lazy_or = null_or_safe_add(dynamic_value(null));
$lazy_ternary = null_ternary_safe_add(dynamic_value(null));
$callback_null = any_null([false, null], fn($value, $key) => is_null($value));
$callback_none = any_null([false, 'null'], fn($value, $key) => is_null($value));
$mapped = array_map('null_flag', [null, false, 'null']);
$present_null = array_key_exists('null', $values) && !isset($values['null']) && is_null($values['null']);
$missing = array_key_exists('missing', $values);

echo 'null=' . bool_word(Is_NuLl(null)) . "\n";
echo 'scalars=' . bool_word(is_null(false)) . ':' . bool_word(is_null(true)) . ':' . bool_word(is_null(0)) . ':' . bool_word(is_null(1)) . ':' . bool_word(is_null(-1)) . ':' . bool_word(is_null('')) . ':' . bool_word(is_null('0')) . ':' . bool_word(is_null('null')) . "\n";
echo 'composites=' . bool_word(is_null($empty)) . ':' . bool_word(is_null($object)) . ':' . bool_word(is_null($arrow)) . "\n";
echo 'dynamic=' . bool_word($dynamic_null) . ':' . bool_word($dynamic_text) . "\n";
echo 'marker=' . bool_word($marker) . "\n";
echo 'lazy=' . bool_word($lazy_or) . ':' . bool_word($lazy_ternary) . "\n";
echo 'array=' . bool_word($present_null) . ':' . bool_word(isset($values['null'])) . ':' . bool_word($missing) . ':' . bool_word(is_null($values['false'])) . "\n";
echo 'callback=' . bool_word($callback_null) . ':' . bool_word($callback_none) . "\n";
echo 'mapped=' . bool_word($mapped[0]) . ':' . bool_word($mapped[1]) . ':' . bool_word($mapped[2]) . "\n";
