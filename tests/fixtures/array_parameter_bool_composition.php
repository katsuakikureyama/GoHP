<?php
// Owned differential fixture: named array parameters composed with named :bool
// returns. Nested arrays enter a string callback through array_map.
function bool_word($value) {
    if (is_bool($value) && $value === true) {
        return 'T';
    }
    if (is_bool($value) && $value === false) {
        return 'F';
    }
    return 'X';
}

function forward_group_flag(array $items): bool {
    return group_flag($items);
}

function group_flag(array $items): bool {
    echo '<group>';
    // Mutating the array parameter must not mutate the caller's nested value.
    $items[] = 'local';

    if ($items[0] === 'yes') {
        return true;
    } elseif ($items[0] === 'cast') {
        return (bool) $items[1];
    } else if ($items[0] === 'scan') {
        foreach ($items as $value) {
            if ($value === 'yes') {
                return true;
            }
        }
        return false;
    }

    return false;
}

$groups = [
    ['yes'],
    ['cast', ''],
    ['scan', 'no', 'yes'],
    ['scan', 'no'],
    ['other'],
];
$flags = array_map('forward_group_flag', $groups);
$empty = array_map('forward_group_flag', []);

echo "\nflags:"
    . bool_word($flags[0])
    . ':' . bool_word($flags[1])
    . ':' . bool_word($flags[2])
    . ':' . bool_word($flags[3])
    . ':' . bool_word($flags[4]) . "\n";
echo 'source:'
    . count($groups[0])
    . ':' . count($groups[1])
    . ':' . count($groups[2])
    . ':' . count($groups[3])
    . ':' . count($groups[4]) . "\n";
echo 'empty:' . count($empty) . "\n";
