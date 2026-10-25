<?php
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

function forwarded_int($value): bool {
    return is_int($value);
}

function guarded_add($value): bool {
    return is_int($value) && $value + 1 === 1;
}

function all_integer_keys(array $items, callable $callback): bool {
    foreach ($items as $key => $value) {
        if (!$callback($value, $key)) {
            return false;
        }
    }
    return true;
}

function forwarded_keys(array $items, callable $callback): bool {
    return all_integer_keys($items, $callback);
}

class IntProbe {
    public $value = 'object';
}

$maximum = 9223372036854775807;
$numeric_string = '42';
$empty = [];
$assoc = ['name' => 'value'];
$object = new IntProbe();
$arrow = fn($value) => $value;
$integer_keys = [-4 => 'negative', 7 => 'positive'];
$mixed_keys = [-4 => 'negative', 'name' => 'string'];
$observed_keys = array_keys($mixed_keys);

$dynamic_zero = forwarded_int(dynamic_value(0));
$dynamic_text = forwarded_int(dynamic_value($numeric_string));
$marker = is_int(marked_value(7));
$guard_zero = guarded_add(dynamic_value(0));
$guard_text = guarded_add(dynamic_value('not-a-number'));
$all_integer = all_integer_keys($integer_keys, fn($value, $key) => is_int($key));
$mixed_integer = forwarded_keys($mixed_keys, fn($value, $key) => is_int($key));
$empty_integer = all_integer_keys([], fn($value, $key) => is_int($key));

echo 'integers=' . bool_word(Is_InT(0)) . ':' . bool_word(is_int(-7)) . ':' . bool_word(is_int(42)) . ':' . bool_word(is_int($maximum)) . "\n";
echo 'non-integers=' . bool_word(is_int(false)) . ':' . bool_word(is_int(true)) . ':' . bool_word(is_int(null)) . ':' . bool_word(is_int($numeric_string)) . ':' . bool_word(is_int('0')) . ':' . bool_word(is_int('')) . "\n";
echo 'composites=' . bool_word(is_int($empty)) . ':' . bool_word(is_int($assoc)) . ':' . bool_word(is_int($object)) . ':' . bool_word(is_int($arrow)) . "\n";
echo 'dynamic=' . bool_word($dynamic_zero) . ':' . bool_word($dynamic_text) . "\n";
echo 'marker=' . bool_word($marker) . "\n";
echo 'guard=' . bool_word($guard_zero) . ':' . bool_word($guard_text) . "\n";
echo 'keys=' . bool_word(is_int($observed_keys[0])) . ':' . bool_word(is_int($observed_keys[1])) . ':' . ($observed_keys[1] === 'name' ? 'name' : 'wrong') . "\n";
echo 'callback=' . bool_word($all_integer) . ':' . bool_word($mixed_integer) . ':' . bool_word($empty_integer) . "\n";
