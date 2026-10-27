<?php
// Differential coverage for the ValueSlot lane with exact scalar parameters.
// Only int/string/bool entries are admitted; typed and untyped array helpers
// below deliberately remain on the legacy lane for host-array compatibility.

function owned_parameter_trace(int $limit, string $label, bool $include) {
    $items = ['label' => $label];
    $copy = $items;
    if ($include) {
        $items['branch'] = 'yes';
    } else {
        $items['branch'] = 'no';
    }
    $i = 0;
    while ($i !== $limit) {
        $items[] = $i;
        $i = $i + 1;
    }
    $copy['copy'] = 'kept';
    return $items['label'] . ':' . $items['branch'] . ':' . count($items) . ':' . count($copy);
}

// The string parameter is read at entry, then reused as a local flat array
// before the loop. Its storage domain is therefore stable at every back edge.
function owned_parameter_reuse(int $steps, string $seed, bool $append) {
    $seed = ['seed' => $seed];
    $i = 0;
    while ($i !== $steps) {
        if ($append) {
            $seed[] = $i;
        } else {
            $seed['last'] = $i;
        }
        $i = $i + 1;
    }
    return $seed['seed'] . ':' . count($seed) . ':' . ($append ? $seed[0] : $seed['last']);
}

function owned_parameter_return(int $steps, string $label, bool $enabled): array {
    $items = ['label' => $label];
    $i = 0;
    while ($i !== $steps) {
        $items[] = $i;
        $i = $i + 1;
    }
    if ($enabled) {
        $items['flag'] = 'yes';
    }
    return $items;
}

// $unused is intentionally never read. $word is read-only after its exact
// entry guard, while $flag selects a branch without widening any local slot.
function owned_parameter_read_only(int $unused, string $word, bool $flag) {
    $items = ['word' => $word];
    if ($flag) {
        $items['branch'] = 'yes';
    } else {
        $items['branch'] = 'no';
    }
    return $items['word'] . ':' . $items['branch'] . ':' . count($items);
}

// The consumer calls this only to prove the third entry guard fires before a
// later body failure; it is intentionally absent from differential output.
function owned_parameter_guard_first(int $number, string $text, bool $flag) {
    $items = ['text' => $text];
    $edge = 9223372036854775807;
    return $edge + 1;
}

// These helpers retain the pre-existing general Array/Clone entry path.
function legacy_typed_array_identity(array $items) {
    return $items;
}

function legacy_untyped_identity($value) {
    return $value;
}

// A legacy stateful helper enables Execution wrapper coverage for the selected
// parameterized functions without making their bodies stateful.
function owned_parameter_state_start() {
    return strtok('first,second', ',');
}

function owned_parameter_state_next() {
    return strtok(',');
}

echo 'trace=' . owned_parameter_trace(2, 'seed', true) . ':' . owned_parameter_trace(0, 'zero', false) . "\n";
echo 'reuse=' . owned_parameter_reuse(2, 'flat', true) . ':' . owned_parameter_reuse(1, 'replace', false) . "\n";
$first = owned_parameter_return(2, 'owned', true);
$second = owned_parameter_return(2, 'owned', true);
$first[] = 'main';
echo 'returned=' . $first['label'] . ':' . $first[0] . ':' . $first[1] . ':' . $first['flag'] . ':' . $first[2] . ':' . count($second) . "\n";
echo 'readonly=' . owned_parameter_read_only(99, 'word', true) . ':' . owned_parameter_read_only(0, 'other', false) . "\n";
echo 'state=' . owned_parameter_state_start() . ':' . owned_parameter_state_next() . "\n";
