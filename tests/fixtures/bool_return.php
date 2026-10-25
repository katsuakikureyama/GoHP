<?php
// Owned bool return-boundary fixture. The only type declarations are named
// global function returns, deliberately including mixed-case BOOL.
function literal_true(): BOOL {
    return true;
}

function literal_false(): bool {
    return false;
}

function first_take($items): bool {
    foreach ($items as $item) {
        if ($item === 'take') {
            return true;
        }
    }
    return false;
}

function forwarded_bool($value): bool {
    return marker_bool($value);
}

function marker_bool($value): bool {
    echo '<marker:' . $value . '>';
    return $value === 'yes';
}

function identity_bool($value): bool {
    return $value;
}

function map_yes($value): bool {
    return $value === 'yes';
}

function explicit_bool($value): bool {
    return (bool) $value;
}

function bool_flags($items): array {
    return array_map('map_yes', $items);
}

function nested_bool($items): bool {
    $flags = bool_flags($items);
    return $flags[0] === true && $flags[1] === false;
}

$source = ['yes', 'no'];
$mapped = array_map('map_yes', $source);

function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

echo 'literal:' . bool_word(literal_true()) . ':' . bool_word(literal_false()) . "\n";
echo 'identity:' . bool_word(identity_bool(true)) . ':' . bool_word(identity_bool(false)) . "\n";
echo 'early:' . bool_word(first_take(['skip', 'take'])) . ':' . bool_word(first_take([])) . "\n";
echo 'forward:' . bool_word(forwarded_bool('yes')) . "\n";
echo 'cast:' . bool_word(explicit_bool('')) . ':' . bool_word(explicit_bool('value')) . "\n";
echo 'mapped:' . bool_word($mapped[0]) . ':' . bool_word($mapped[1]) . ':' . $source[0] . ':' . $source[1] . "\n";
echo 'nested:' . bool_word(nested_bool($source)) . "\n";
