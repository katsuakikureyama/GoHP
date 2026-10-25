<?php
function choose($value) {
    if ($value) {
        return 'then';
    } else {
        return 'else';
    }
}

function marker($name, $value) {
    echo '<' . $name . '>';
    return $value;
}

function marked_choice($value) {
    if (marker('condition', $value)) {
        return marker('then', 'T');
    } else {
        return marker('else', 'E');
    }
}

function nested_choice($value) {
    if ($value) {
        if (is_string($value)) {
            return 'string';
        } else {
            return 'truthy';
        }
    } else {
        if (is_array($value)) {
            return 'empty-array';
        } else {
            return 'falsey';
        }
    }
}

function branch_arrays($value) {
    $original = ['root'];
    $chosen = $original;
    if ($value) {
        $local = 'then';
        $chosen[] = $local;
        echo '<then:' . $chosen[1] . '>';
    } else {
        $local = 'else';
        $chosen['branch'] = $local;
        echo '<else:' . $chosen['branch'] . '>';
    }
    return count($original) . ':' . count($chosen) . ':' . $original[0];
}

function describe($value) {
    if (is_array($value)) {
        return 'array:' . count($value);
    } else {
        return 'scalar';
    }
}

function changed_value($value) {
    $current = [];
    if ($value) {
        $current = ['then'];
    } else {
        $current = 'else';
    }
    return describe($current);
}

function walk($items) {
    $out = '';
    foreach ($items as $key => $value) {
        if ($value) {
            $out = $out . 'T' . $key;
        } else {
            $out = $out . 'F' . $key;
        }
    }
    return $out;
}

$empty = [];
$items = [0, 'value'];
$empty_branch = 'steady';

echo 'choose=' . choose(null) . ',' . choose(false) . ',' . choose(true) . ',' . choose(0) . ',' . choose(-7) . ',' . choose('') . ',' . choose('0') . ',' . choose('value') . ',' . choose($empty) . ',' . choose($items) . "\n";
echo 'marked=' . marked_choice(false) . ':' . marked_choice(true) . "\n";
echo 'nested=' . nested_choice($empty) . ':' . nested_choice(null) . ':' . nested_choice('value') . ':' . nested_choice(-7) . "\n";
echo 'arrays=' . branch_arrays(true) . ':' . branch_arrays(false) . "\n";
echo 'types=' . changed_value(true) . ':' . changed_value(false) . "\n";
$changed = 0;
if (false) {
} else {
    $changed = ['else'];
}
echo 'changed-type=' . count($changed) . ':' . $changed[0] . "\n";
$sibling = [];
if (false) {
    $sibling = 'then-only-type';
} else {
    $sibling[] = 'independent';
}
echo 'sibling-type=' . count($sibling) . ':' . $sibling[0] . "\n";
echo 'walk=' . walk([0, 'yes', '', -1]) . "\n";

if (false) {
} else {
}
echo 'empty=' . $empty_branch . "\n";
