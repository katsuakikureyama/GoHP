<?php
// Bare continue targets only the innermost foreach, including when nested in
// an if. The outer separator proves that every outer iteration still advances.
function collect_values($items) {
    $out = '';
    foreach ($items as $item) {
        if ($item === 'skip') {
            CoNtInUe;
        }
        $out = $out . $item;
    }
    return $out;
}

function collect_nested($groups) {
    $out = '';
    foreach ($groups as $group) {
        foreach ($group as $value) {
            if ($value === 'skip') {
                continue;
            }
            $out = $out . $value;
        }
        $out = $out . '|';
    }
    return $out;
}

$first = collect_values(['a', 'skip', 'b', 'skip', 'c']);
$nested = collect_nested([['a', 'skip', 'b'], ['skip', 'c'], []]);

echo 'first=' . $first . "\n";
echo 'nested=' . $nested . "\n";
