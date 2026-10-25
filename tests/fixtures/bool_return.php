<?php
// Owned differential fixture: bounded global named :bool returns only.
function literal_true(): BoOl {
    return TRUE;
}

function literal_false(): bool {
    return false;
}

function forward_flag($value): bool {
    return predicate_flag($value);
}

function predicate_flag($value): bool {
    return is_bool($value) && $value === true;
}

function branch_flag($first, $second): bool {
    if ($first) {
        return true;
    } elseif ($second) {
        return false;
    } else if (is_bool($second)) {
        return $second === true;
    } else {
        return false;
    }
}

function first_yes($items): bool {
    foreach ($items as $value) {
        if ($value === 'yes') {
            return true;
        }
    }
    return false;
}

function map_flag($value): BOOL {
    return $value === 'yes';
}

function mark_bool($value): bool {
    echo '<return>';
    return $value;
}

function marked_return(): bool {
    return mark_bool(true);
}

function skipped_branch(): bool {
    if (true) {
        return true;
    } elseif (mark_bool(false)) {
        return false;
    } else if (mark_bool(false)) {
        return false;
    } else {
        return false;
    }
}

$yes = literal_true();
$no = literal_false();
$input = ['yes', 'no', 'yes'];
$mapped = array_map('map_flag', $input);
$copy = $mapped;
$copy[0] = false;
$marked = marked_return();
$skipped = skipped_branch();

echo 'literals:' . (is_bool($yes) ? 'bool' : 'wrong') . ':' . ($yes === true ? 'true' : 'wrong') . ':' . ($no === false ? 'false' : 'wrong') . "\n";
echo 'forward:' . (forward_flag(true) === true ? 'true' : 'wrong') . ':' . (forward_flag(false) === false ? 'false' : 'wrong') . "\n";
echo 'branches:' . (branch_flag(true, true) ? 'first' : 'wrong') . ':' . (branch_flag(false, true) === false ? 'elseif' : 'wrong') . ':' . (branch_flag(false, false) === false ? 'else-if' : 'wrong') . ':' . (branch_flag(false, null) === false ? 'else' : 'wrong') . "\n";
echo 'foreach:' . (first_yes(['no', 'yes']) ? 'early' : 'wrong') . ':' . (first_yes(['no']) === false ? 'default' : 'wrong') . "\n";
echo 'map:' . (is_bool($mapped[0]) ? 'bool' : 'wrong') . ':' . ($mapped[0] === true ? 'true' : 'wrong') . ':' . ($mapped[1] === false ? 'false' : 'wrong') . ':' . $input[0] . ':' . ($copy[0] === false ? 'copy' : 'wrong') . "\n";
echo 'returns:' . ($marked === true ? 'marked' : 'wrong') . ':' . ($skipped === true ? 'skipped' : 'wrong') . "\n";
