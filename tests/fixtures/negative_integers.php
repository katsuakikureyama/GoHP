<?php
function bool_word($value) {
    if ($value === true) {
        return 'T';
    }
    return 'F';
}

function dump_entries($label, $items) {
    foreach ($items as $key => $value) {
        echo $label . ':' . (is_string($key) ? 'string' : 'int') . ':' . $key . '=' . $value . ',';
    }
    echo "\n";
    return '';
}

function negative_return() {
    return -7;
}

function pass_through($value) {
    return $value;
}

class NegativeDefaults {
    public $value = -9;
}

$defaults = new NegativeDefaults();
echo 'literals=' . -5 . ':' . -0 . ':' . - /* whitespace and comment are allowed */ 6 . "\n";
echo 'function=' . negative_return() . ':' . pass_through(-8) . "\n";
echo 'property=' . $defaults->value . "\n";
echo 'precedence=' . -5 + 6 . ':' . bool_word(-0 === 0) . ':' . (true ? -2 : -3) . "\n";

$short = [-5 => 'short-negative', 'short-implicit'];
$legacy = array(-5 => 'legacy-negative', 'legacy-implicit');
$short_keys = array_keys($short);
$legacy_keys = array_keys($legacy);
echo 'key-types=' . bool_word(is_string($short_keys[0])) . ':' . bool_word($short_keys[0] === -5) . ':' . bool_word($short_keys[1] === -4) . ':' . bool_word(is_string($legacy_keys[0])) . ':' . bool_word($legacy_keys[0] === -5) . ':' . bool_word($legacy_keys[1] === -4) . "\n";
echo dump_entries('short', $short);
echo dump_entries('legacy', $legacy);

$written = ['prefix' => 'string-key'];
$written[-5] = 'written-negative';
$written[] = 'first-append';
$written[-2] = 'higher-negative';
$written[] = 'second-append';
$written[-7] = 'lower-negative';
$written[] = 'zero-append';
echo dump_entries('written', $written);

$minimum = -9223372036854775807 + -1;
echo 'minimum=' . $minimum . "\n";
